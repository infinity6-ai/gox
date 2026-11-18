package httpzsite

import (
	"fmt"
	"strings"

	"github.com/infinity6-ai/gox/commonz/pathz"
	"github.com/spf13/cobra"
)

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
