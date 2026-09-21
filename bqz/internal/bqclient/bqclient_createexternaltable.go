package bqclient

import (
	"context"
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/infinity6-ai/gox/bqz/bqz"
	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/slicez"
	"github.com/infinity6-ai/gox/commonz/validation"
)

func (c *Client) CreateExternalTable(ctx context.Context, opts bqz.ExternalTable) error {
	err := c.CreateDataset(ctx, DatasetOptions{
		Dataset: opts.Dataset,
		Metadata: &bigquery.DatasetMetadata{
			Labels: map[string]string{
				"i6ds": opts.Dataset,
			},
		},
	})
	if err != nil {
		return err
	}

	schema, ok := opts.Schema.(bigquery.Schema)
	if !ok {
		schema, err = bigquery.InferSchema(opts.Schema)
		errorz.Check(err)
	}

	extConfig := &bigquery.ExternalDataConfig{
		SourceFormat:        bigquery.JSON,
		SourceURIs:          []string{opts.Uri},
		Compression:         bigquery.Gzip,
		AutoDetect:          false,
		IgnoreUnknownValues: true,
	}

	if len(opts.HiveParts) > 0 {
		hiveFields := map[string]struct{}{}
		hiveParts := slicez.MustMap(opts.HiveParts, func(_ int, value string) (string, bool) {
			hiveFields[value] = struct{}{}
			return fmt.Sprintf("{%s:STRING}", value), true
		})
		schema = slicez.MustFilter(schema, func(_ int, value *bigquery.FieldSchema) bool {
			_, ok := hiveFields[value.Name]
			return !ok
		})
		idx := strings.LastIndex(opts.Uri, "/")
		validation.Greater(idx, 0, "idx")
		sourceUriPrefix := opts.Uri[:idx] + "/" + strings.Join(hiveParts, "/")
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
	}

	tableRef := c.client.Dataset(opts.Dataset).Table(opts.Table)
	if err := tableRef.Delete(ctx); err != nil && ParseErrorCode(err) != 404 {
		return fmt.Errorf("error deleting table before recreation: %w", err)
	}

	err = tableRef.Create(ctx, meta)
	if err != nil {
		return fmt.Errorf("error creating table: %#v, %w", opts, err)
	}
	return nil
}
