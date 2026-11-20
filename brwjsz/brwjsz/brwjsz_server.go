package brwjsz

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/infinity6-ai/gox/brwjsz/brwclient"
	"github.com/infinity6-ai/gox/brwjsz/internal/handlers"
	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/execz"
	"github.com/infinity6-ai/gox/commonz/pathz"
	"github.com/infinity6-ai/gox/httpz/httpzserver"
)

// SiteOptions configures serving static files from a disk directory on the internal server.
type SiteOptions = handlers.SiteOptions

// Options configures the Browser instance, command arguments, and HTTP server.
type Options struct {
	Site            SiteOptions
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

// Chrome configures default browser flags, logging, headless mode, and isolated user profile.
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
		"--enable-logging=stderr",
		"--no-first-run",
		"--no-default-browser-check",
		"--disable-sync",
		"--disable-component-update",
		"--disable-features=Translate,OptimizationHints",
	}
	if headless {
		flags = append(flags, "--headless", "--ozone-platform=headless")
	}

	o.Command.Args = append(flags, o.Command.Args...)
	return o
}

// Browser manages an HTTP server, a spawned browser OS process, and controller IPC.
type Browser struct {
	tempUserDataDir string
	originalArgs    []string
	cmdOpts         execz.Options
	server          *httpzserver.Server
	hub             *handlers.Hub
	prc             *execz.Process
}

// New creates and configures a Browser instance.
func New(ctx context.Context, opts Options) *Browser {
	hub := handlers.NewHub()
	srv := httpzserver.New(ctx, opts.Server)
	handlers.AddHandlers(ctx, srv, hub, opts.Site)

	ret := &Browser{
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

// Start launches the internal HTTP server, starts the browser process, and waits for the controller to report ready.
func (b *Browser) Start(ctx context.Context) error {
	b.server.Listen()
	b.server.Start()

	b.fixArgs()
	b.prc = execz.New(ctx, b.cmdOpts)

	err := b.prc.Start(ctx)
	if err != nil {
		return fmt.Errorf("error starting browser process: %w", err)
	}

	err = b.hub.WaitReady(ctx)
	if err != nil {
		return fmt.Errorf("error waiting for browser controller ready: %w", err)
	}

	return nil
}

func (b *Browser) fixArgs() {
	u := b.server.Base().Clone()
	u.Path = pathz.MustParse(handlers.ControllerPath)
	urlStr := u.String()

	b.cmdOpts.Args = make([]string, 0, len(b.originalArgs)+1)
	urlReplaced := false
	for _, arg := range b.originalArgs {
		if arg == "URL" {
			b.cmdOpts.Args = append(b.cmdOpts.Args, urlStr)
			urlReplaced = true
		} else {
			b.cmdOpts.Args = append(b.cmdOpts.Args, arg)
		}
	}
	if !urlReplaced {
		b.cmdOpts.Args = append(b.cmdOpts.Args, urlStr)
	}
}

// Process returns the underlying OS process handle.
func (b *Browser) Process() *execz.Process {
	return b.prc
}

// Server returns the underlying HTTP server instance.
func (b *Browser) Server() *httpzserver.Server {
	return b.server
}

// Client returns a Client handle to communicate with the browser controller.
func (b *Browser) Client() *brwclient.Client {
	u := b.server.Base().Clone()
	u.Path = pathz.MustParse(handlers.BasePath)
	return &brwclient.Client{
		BaseUrl:    u,
		Dispatcher: b.hub,
	}
}
