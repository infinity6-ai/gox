package bqzimpl_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/infinity6-ai/gox/bqz/bqz"
	"github.com/infinity6-ai/gox/bqz/bqzdataset"
	"github.com/infinity6-ai/gox/bqz/bqzjob"
	"github.com/infinity6-ai/gox/bqz/bqztable"
	"github.com/infinity6-ai/gox/bqz/internal/bqzimpl"
	"github.com/infinity6-ai/gox/commonz/urlz"
	"github.com/infinity6-ai/gox/fsz/fsz"
	"github.com/infinity6-ai/gox/fsz/fszjson"
	"github.com/stretchr/testify/require"
)

func TestRemoteExternalTable(t *testing.T) {
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

	u := urlz.MustParse("gs://i6-rs-contint-tmp/testds/mytable/a=1/b=x/part.json.gz")
	fsz.MustDelete(ctx, u)

	fszjson.MustUploadSlice(ctx, []SalesHistory{
		{ID: "a", ItemId: "item_a"},
		{ID: "b", ItemId: "item_b"},
	}, fszjson.UploadOptions{
		Url:  u,
		Gzip: true,
	})

	type testScenario struct {
		dataset       string
		table         string
		uri           string
		hiveParts     []string
		schema        any
		query         string
		binds         map[string]any
		expectedValue int
	}

	c := bqzimpl.New(ctx, bqz.ClientOptions{
		Project: "i6-rs-contint",
	})

	check := func(t *testing.T, s testScenario) {
		t.Helper()

		if s.uri != "" {
			err := c.CreateExternalTable(ctx, &bqz.ExternalTable{
				Dataset:   bqzdataset.New(s.dataset),
				Table:     bqztable.New(s.table),
				Uri:       s.uri,
				HiveParts: s.hiveParts,
				Schema:    s.schema,
			})
			require.NoError(t, err)
		}

		if s.query != "" {
			jobId := bqzjob.New(fmt.Sprintf("test_job_%d", time.Now().UnixNano()))
			q := &bqz.Query{
				Job:   jobId,
				Query: s.query,
				Binds: s.binds,
			}
			err := c.Dispatch(ctx, q)
			require.NoError(t, err)

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

			ok, err := it(ctx, &row)
			require.NoError(t, err)
			require.True(t, ok)
			require.Equal(t, s.expectedValue, row.Num)

			ok, err = it(ctx, &row)
			require.NoError(t, err)
			require.False(t, ok)

			ok, err = it(ctx, &row)
			require.NoError(t, err)
			require.False(t, ok)
		}
	}

	t.Run("Create external table", func(t *testing.T) {
		check(t, testScenario{
			dataset:       "testds",
			table:         "mytable",
			uri:           "gs://i6-rs-contint-tmp/testds/mytable/*",
			hiveParts:     []string{"a", "b"},
			schema:        &SalesHistory{},
			query:         "select count(0) AS num from testds.mytable",
			expectedValue: 2,
		})
	})

	t.Run("Dispatch query job and read results", func(t *testing.T) {
		check(t, testScenario{
			query:         "SELECT 42 AS num",
			expectedValue: 42,
		})
	})

	t.Run("Dispatch query job with binds and read results", func(t *testing.T) {
		check(t, testScenario{
			query:         "SELECT @val AS num",
			binds:         map[string]any{"val": 99},
			expectedValue: 99,
		})
	})
}
