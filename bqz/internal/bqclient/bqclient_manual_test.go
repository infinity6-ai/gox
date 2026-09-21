package bqclient_test

import (
	"context"
	"testing"

	"github.com/infinity6-ai/gox/bqz/bqz"
	"github.com/infinity6-ai/gox/bqz/internal/bqclient"
	"github.com/infinity6-ai/gox/commonz/urlz"
	"github.com/infinity6-ai/gox/fsz/fsz"
	"github.com/stretchr/testify/require"
)

func TestManualExternalTable(t *testing.T) {
	ctx := context.Background()

	type SalesHistory struct {
		ID                 string `json:"id" bigquery:"id"`
		IngestedAt         string `json:"ingested_at" bigquery:"ingested_at"`
		AggrQt             string `json:"aggr_qt" bigquery:"aggr_qt"`
		GroupId            string `json:"group_id" bigquery:"group_id"`
		ItemDs             string `json:"item_ds" bigquery:"item_ds"`
		ItemId             string `json:"item_id" bigquery:"item_id"`
		XConsumerPrice     string `json:"x_consumer_price" bigquery:"x_consumer_price"`
		XPosPrice          string `json:"x_pos_price" bigquery:"x_pos_price"`
		XRegulatedMaxPrice string `json:"x_regulated_max_price" bigquery:"x_regulated_max_price"`
	}

	u := urlz.MustParse("gs://i6-rs-contint-tmp/testds/mytable/a=1/b=x/part.json")
	fsz.MustDelete(ctx, u)

	// fsz.MustUpload(ctx, u, nil, gzipz.)

	type testScenario struct {
		dataset       string
		table         string
		uri           string
		hiveParts     []string
		schema        any
		query         string
		expectedValue int
	}

	c := bqclient.New(ctx, bqclient.ClientOptions{
		Project: "i6-rs-contint",
	})

	check := func(t *testing.T, s testScenario) {
		t.Helper()

		if s.uri != "" {
			err := c.CreateExternalTable(ctx, bqz.ExternalTable{
				Dataset:   s.dataset,
				Table:     s.table,
				Uri:       s.uri,
				HiveParts: s.hiveParts,
				Schema:    s.schema,
			})
			require.NoError(t, err)
		}

		if s.query != "" {
			jobId, err := c.Dispatch(ctx, bqclient.QueryOptions{
				Query: s.query,
			})
			require.NoError(t, err)
			require.NotEmpty(t, jobId)

			err = c.WaitFor(ctx, jobId)
			require.NoError(t, err)

			done, err := c.IsDone(ctx, jobId)
			require.NoError(t, err)
			require.True(t, done)

			it, err := c.Read(ctx, jobId)
			require.NoError(t, err)

			type resultRow struct {
				Num int `bigquery:"num"`
			}
			var row resultRow
			err = it.Next(&row)
			require.NoError(t, err)
			require.Equal(t, s.expectedValue, row.Num)
		}
	}

	t.Run("Create external table", func(t *testing.T) {
		check(t, testScenario{
			dataset:       "testds",
			table:         "mytable",
			uri:           "gs://i6-rs-contint-tmp/testds/mytable",
			hiveParts:     []string{"a", "b"},
			schema:        &SalesHistory{},
			query:         "select count(0) AS num from testds.mytable",
			expectedValue: 1,
		})
	})

	t.Run("Dispatch query job and read results", func(t *testing.T) {
		check(t, testScenario{
			query:         "SELECT 42 AS num",
			expectedValue: 42,
		})
	})
}
