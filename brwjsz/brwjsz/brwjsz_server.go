package brwjsz

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/infinity6-ai/gox/brwjsz/brwclient"
	"github.com/infinity6-ai/gox/brwjsz/internal/handlers"
	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/execz"
	"github.com/infinity6-ai/gox/httpz/httpzserver"
)

// Options configures the Browser instance, tab count, command arguments, and HTTP server.
type Options struct {
	TabCount        int
	Command         execz.Options
	Server          httpzserver.Options
	tempUserDataDir string
}

func findChrome() string {
	for _, name := range []string{"google-chrome", "chromium", "chromium-browser", "chrome"} {
		if path, err := exec.LookPath(name); err == nil {
			return path
		}
	}
	return "google-chrome"
}

// Chrome configures default browser flags, anti-throttling options, headless mode, and isolated user profile.
// If binPath is empty, it discovers google-chrome/chromium in PATH.
// If userDataDir is empty, a temporary profile directory is created and cleaned up on Browser.Close().
func (o *Options) Chrome(binPath string, headless bool, userDataDir string) *Options {
	if binPath == "" {
		binPath = findChrome()
	}
	o.Command.Bin = binPath

	if userDataDir == "" {
		tmp, err := os.MkdirTemp("", "brwjsz-profile-*")
		errorz.Check(err)
		o.tempUserDataDir = tmp
		userDataDir = tmp
	}

	flags := []string{
		"--user-data-dir=" + userDataDir,
		"--no-sandbox",
		"--disable-dev-shm-usage",
		"--disable-gpu",
		"--no-first-run",
		"--no-default-browser-check",
		"--disable-sync",
		"--disable-background-timer-throttling",
		"--disable-backgrounding-occluded-windows",
		"--disable-renderer-backgrounding",
		"--disable-component-update",
		"--disable-features=Translate,OptimizationHints",
	}
	if headless {
		flags = append(flags, "--headless=new")
	}

	o.Command.Args = append(flags, o.Command.Args...)
	return o
}

// Browser manages an HTTP server, a spawned browser OS process, and tab controllers.
type Browser struct {
	tabCount        int
	tempUserDataDir string
	originalArgs    []string
	cmdOpts         execz.Options
	server          *httpzserver.Server
	hub             *handlers.Hub
	prc             *execz.Process
}

// New creates and configures a Browser instance.
func New(ctx context.Context, opts Options) *Browser {
	tabCount := opts.TabCount
	if tabCount <= 0 {
		tabCount = 1
	}
	hub := handlers.NewHub()
	srv := httpzserver.New(ctx, opts.Server)
	handlers.AddHandlers(ctx, srv, hub)

	ret := &Browser{
		tabCount:        tabCount,
		tempUserDataDir: opts.tempUserDataDir,
		server:          srv,
		hub:             hub,
		cmdOpts:         opts.Command,
		originalArgs:    opts.Command.Args,
	}
	return ret
}

// Close gracefully terminates the browser process, stops the HTTP server, and cleans up temporary profiles.
func (b *Browser) Close() error {
	var prcErr, serverErr, cleanErr error
	if b.prc != nil {
		prcErr = b.prc.Close()
	}
	if b.server != nil {
		serverErr = b.server.Close()
	}
	if b.tempUserDataDir != "" {
		cleanErr = os.RemoveAll(b.tempUserDataDir)
	}
	err := errors.Join(prcErr, serverErr, cleanErr)
	if err != nil {
		return fmt.Errorf("error closing browser: %w", err)
	}
	return nil
}

// Start launches the internal HTTP server, starts the browser process, and waits for all tabs to report ready.
func (b *Browser) Start(ctx context.Context) error {
	b.server.Listen()
	b.server.Start()

	b.fixArgs()
	b.prc = execz.New(ctx, b.cmdOpts)

	err := b.prc.Start(ctx)
	if err != nil {
		return fmt.Errorf("error starting browser process: %w", err)
	}

	err = b.hub.WaitReady(ctx, b.tabCount)
	if err != nil {
		return fmt.Errorf("error waiting for tabs ready: %w", err)
	}

	return nil
}

func (b *Browser) fixArgs() {
	tabCount := b.tabCount
	if tabCount <= 0 {
		tabCount = 1
	}

	tabUrls := make([]string, tabCount)
	for i := 0; i < tabCount; i++ {
		u := b.server.Base().MustJoinPathString(handlers.ControllerPath)
		u.Query = "tab=" + strconv.Itoa(i)
		tabUrls[i] = u.String()
	}

	b.cmdOpts.Args = make([]string, 0, len(b.originalArgs)+len(tabUrls))
	urlReplaced := false
	for _, arg := range b.originalArgs {
		if arg == "URL" {
			b.cmdOpts.Args = append(b.cmdOpts.Args, tabUrls...)
			urlReplaced = true
		} else {
			b.cmdOpts.Args = append(b.cmdOpts.Args, arg)
		}
	}
	if !urlReplaced {
		b.cmdOpts.Args = append(b.cmdOpts.Args, tabUrls...)
	}
}

// Process returns the underlying OS process handle.
func (b *Browser) Process() *execz.Process {
	return b.prc
}

// Tab returns a Client handle scoped to the specified tab index.
func (b *Browser) Tab(index int) *brwclient.Client {
	return &brwclient.Client{
		BaseUrl:    b.server.Base().MustJoinPathString(handlers.BasePath),
		Tab:        index,
		Dispatcher: b.hub,
	}
}

// Client returns a Client handle scoped to Tab 0.
func (b *Browser) Client() *brwclient.Client {
	return b.Tab(0)
}
