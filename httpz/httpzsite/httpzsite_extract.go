package httpzsite

import (
	"context"
	"fmt"

	"github.com/infinity6-ai/gox/commonz/pathz"
	"github.com/infinity6-ai/gox/commonz/staticz"
)

// ExtractTo extracts the site's static pack assets to the specified destination path.
func (s *Site) ExtractTo(ctx context.Context, path *pathz.Path) error {
	err := staticz.ExtractTo(ctx, path, s.Pack)
	if err != nil {
		return fmt.Errorf("failed to extract site pack to %s: %w", path, err)
	}
	return nil
}
