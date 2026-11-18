package httpzsite

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/infinity6-ai/gox/commonz/pathz"
	"github.com/infinity6-ai/gox/commonz/staticz"
	"github.com/infinity6-ai/gox/commonz/urlz"
	"github.com/infinity6-ai/gox/httpz/httpzserver"
	"github.com/infinity6-ai/gox/httpz/httpzserver/filters/httpzserverfilterauthbasic"
	"github.com/infinity6-ai/gox/httpz/httpzserver/httpzserverstatic"
	"github.com/spf13/cobra"
)

type Site struct {
	Context      context.Context
	LocalAddress string
	Pack         any
	BasicUsers   map[string]string

	server *httpzserver.Server
}

// ExtractTo extracts the site's static pack assets to the specified destination path.
func (s *Site) ExtractTo(ctx context.Context, path *pathz.Path) error {
	err := staticz.ExtractTo(ctx, path, s.Pack)
	if err != nil {
		return fmt.Errorf("failed to extract site pack to %s: %w", path, err)
	}
	return nil
}

func (s *Site) Listen() {
	if s.server == nil {
		s.server = httpzserver.New(s.Context, httpzserver.Options{
			LocalAddress: s.LocalAddress,
		})
	}
	s.server.Listen()
}

func (s *Site) AddHandlers() {
	if s.BasicUsers != nil {
		s.server.AddFilter(httpzserverfilterauthbasic.New(s.Context, httpzserverfilterauthbasic.Options{
			Verify: func(ctx context.Context, user string, pass string) (bool, error) {
				expectedPass, ok := s.BasicUsers[user]
				if !ok {
					return false, nil
				}
				return expectedPass == pass, nil
			},
		}))
	}
	s.server.AddHandler("GET", "/*", httpzserverstatic.StaticHandler(s.Pack))
}

// Addr returns the network address the server is listening on, or nil if not listening.
func (s *Site) Addr() net.Addr {
	if s.server == nil {
		return nil
	}
	return s.server.Addr()
}

// Base returns the base URL of the running site, using the listening address.
func (s *Site) Base() *urlz.Url {
	if s.server == nil {
		return nil
	}
	return s.server.Base()
}

func (s *Site) Close() error {
	if s.server == nil {
		return nil
	}
	return s.server.Close()
}

func (s *Site) Start() {
	s.server.Start()
}

func (s *Site) Serve() {
	s.server.Serve()
}

// Cmd returns a parent cobra command with 'serve' and 'extract' subcommands for the site.
func (s *Site) Cmd(use ...string) *cobra.Command {
	name := "site"
	if len(use) > 0 && use[0] != "" {
		name = use[0]
	}

	cmd := &cobra.Command{
		Use:   name,
		Short: "Site commands",
	}

	var address string
	var users []string
	defaultAddr := s.LocalAddress
	if defaultAddr == "" {
		defaultAddr = "localhost:8080"
	}

	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Serve the static site",
		RunE: func(cmd *cobra.Command, args []string) error {
			if address != "" {
				s.LocalAddress = address
			}
			if len(users) > 0 {
				if s.BasicUsers == nil {
					s.BasicUsers = make(map[string]string)
				}
				for _, u := range users {
					for _, item := range strings.Split(u, ",") {
						item = strings.TrimSpace(item)
						if item == "" {
							continue
						}
						user, pass, ok := strings.Cut(item, ":")
						if !ok {
							return fmt.Errorf("invalid user format %q, expected user:password", item)
						}
						s.BasicUsers[user] = pass
					}
				}
			}
			s.Context = cmd.Context()
			go func() {
				<-cmd.Context().Done()
				_ = s.Close()
			}()
			s.Listen()
			s.AddHandlers()
			s.Serve()
			return nil
		},
	}
	serveCmd.Flags().StringVar(&address, "address", defaultAddr, "Address to listen on")
	serveCmd.Flags().StringArrayVar(&users, "user", nil, "Basic auth user in user:password format")

	var out string
	extractCmd := &cobra.Command{
		Use:   "extract",
		Short: "Extract site assets to directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			if out == "" {
				return fmt.Errorf("--out directory is required")
			}
			dest := pathz.MustParse(out)
			err := s.ExtractTo(cmd.Context(), dest)
			if err != nil {
				return fmt.Errorf("failed to extract: %w", err)
			}
			return nil
		},
	}
	extractCmd.Flags().StringVar(&out, "out", "", "Destination directory to extract files into")
	_ = extractCmd.MarkFlagRequired("out")

	cmd.AddCommand(serveCmd, extractCmd)
	return cmd
}
