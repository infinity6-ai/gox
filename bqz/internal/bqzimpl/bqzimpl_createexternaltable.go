package bqzimpl

import (
	"context"
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/infinity6-ai/gox/bqz/bqz"
	"github.com/infinity6-ai/gox/commonz/slicez"
	"github.com/infinity6-ai/gox/commonz/validation"
)

// CreateExternalTable implements [bqz.Service].
func (b *BqzServiceImpl) CreateExternalTable(ctx context.Context, table *bqz.ExternalTable) error {
	err := b.CreateDataset(ctx, &bqz.Dataset{
		Name: table.Dataset.Get(),
		Labels: map[string]string{
			"i6ds": table.Dataset.Get(),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to ensure dataset %s exists: %w", table.Dataset.Get(), err)
	}

	schema, ok := table.Schema.(bigquery.Schema)
	if !ok {
		schema, err = bigquery.InferSchema(table.Schema)
		if err != nil {
			return fmt.Errorf("failed to infer schema for external table %s: %w", table.Table.Get(), err)
		}
	}

	extConfig := &bigquery.ExternalDataConfig{
		SourceFormat:        bigquery.JSON,
		SourceURIs:          []string{table.Uri},
		Compression:         bigquery.Gzip,
		AutoDetect:          false,
		IgnoreUnknownValues: true,
	}

	if len(table.HiveParts) > 0 {
		hiveFields := map[string]struct{}{}
		hiveParts := slicez.MustMap(table.HiveParts, func(_ int, value string) (string, bool) {
			hiveFields[value] = struct{}{}
			return fmt.Sprintf("{%s:STRING}", value), true
		})
		schema = slicez.MustFilter(schema, func(_ int, value *bigquery.FieldSchema) bool {
			_, ok := hiveFields[value.Name]
			return !ok
		})
		idx := strings.LastIndex(table.Uri, "/")
		validation.Greater(idx, 0, "idx")
		sourceUriPrefix := table.Uri[:idx] + "/" + strings.Join(hiveParts, "/")
		extConfig.HivePartitioningOptions = &bigquery.HivePartitioningOptions{
			Mode:                   bigquery.CustomHivePartitioningMode,
			RequirePartitionFilter: false,
			SourceURIPrefix:        sourceUriPrefix,
		}
	}

	meta := &bigquery.TableMetadata{
		Schema:             schema,
		ExternalDataConfig: extConfig,
		ExpirationTime:     time.Now().UTC().Add(48 * time.Hour),
		Labels:             table.Labels,
	}

	tableRef := b.c.Dataset(table.Dataset.Get()).Table(table.Table.Get())
	if err := tableRef.Delete(ctx); err != nil && ParseErrorCode(err) != 404 {
		return fmt.Errorf("failed to delete external table %s before recreation: %w", table.Table.Get(), err)
	}

	err = tableRef.Create(ctx, meta)
	if err != nil {
		return fmt.Errorf("failed to create external table %s: %w", table.Table.Get(), err)
	}
	return nil
}
