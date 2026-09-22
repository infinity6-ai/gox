package bqzimpl

import (
	"context"
	"fmt"

	"github.com/infinity6-ai/gox/bqz/bqzdataset"
	"github.com/infinity6-ai/gox/bqz/bqztable"
)

// TableExists implements [bqz.Service].
func (b *BqzServiceImpl) TableExists(ctx context.Context, dataset bqzdataset.Dataset, table bqztable.Table) (bool, error) {
	_, err := b.c.Dataset(dataset.Get()).Table(table.Get()).Metadata(ctx)
	if err != nil {
		if ParseErrorCode(err) == 404 {
			return false, nil
		}
		return false, fmt.Errorf("failed to check table existence for %s.%s: %w", dataset.Get(), table.Get(), err)
	}
	return true, nil
}
