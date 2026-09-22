package bqzimpl_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/infinity6-ai/gox/bqz/bqz"
	"github.com/infinity6-ai/gox/bqz/bqzdataset"
	"github.com/infinity6-ai/gox/bqz/bqzerr"
	"github.com/infinity6-ai/gox/bqz/bqzjob"
	"github.com/infinity6-ai/gox/bqz/bqztable"
	"github.com/infinity6-ai/gox/bqz/internal/bqzimpl"
	"github.com/infinity6-ai/gox/commonz/urlz"
	"github.com/infinity6-ai/gox/fsz/fsz"
	"github.com/infinity6-ai/gox/fsz/fszjson"
	"github.com/stretchr/testify/require"
	"cloud.google.com/go/bigquery"
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
		dataset        string
		table          string
		uri            string
		hiveParts      []string
		schema         any
		query          string
		binds          map[string]any
		runningTimeout time.Duration
		expectedValue  int
	}

	c := bqzimpl.New(ctx, bqz.ClientOptions{
		Project: "i6-rs-contint",
	})

	check := func(t *testing.T, s testScenario) {
		t.Helper()

		if s.uri != "" {
			ds := bqzdataset.New(s.dataset)
			tbl := bqztable.New(s.table)
			_ = c.DropTable(ctx, ds, tbl)

			err := c.CreateExternalTable(ctx, &bqz.ExternalTable{
				Dataset:   ds,
				Table:     tbl,
				Uri:       s.uri,
				HiveParts: s.hiveParts,
				Schema:    s.schema,
			})
			require.NoError(t, err)
		}

		if s.query != "" {
			jobId := bqzjob.New(fmt.Sprintf("test_job_%d", time.Now().UnixNano()))
			q := &bqz.Query{
				Job:            jobId,
				Query:          s.query,
				Binds:          s.binds,
				RunningTimeout: s.runningTimeout,
			}
			err := c.Dispatch(ctx, q)
			require.NoError(t, err)

			err = c.WaitFor(ctx, jobId)
			require.NoError(t, err)

			status, err := c.JobStatus(ctx, jobId)
			require.NoError(t, err)
			require.Equal(t, bqz.JobStatusDone, status)

			read := func() {
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

			read()
			read()
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

	t.Run("Create external table again returns ErrConflict", func(t *testing.T) {
		err := c.CreateExternalTable(ctx, &bqz.ExternalTable{
			Dataset:   bqzdataset.New("testds"),
			Table:     bqztable.New("mytable"),
			Uri:       "gs://i6-rs-contint-tmp/testds/mytable/*",
			HiveParts: []string{"a", "b"},
			Schema:    &SalesHistory{},
		})
		require.Error(t, err)
		require.True(t, errors.Is(err, bqzerr.ErrConflict))
	})

	t.Run("Dispatch query job and read results", func(t *testing.T) {
		check(t, testScenario{
			query:         "SELECT 41 AS num; SELECT 42 AS num",
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

	t.Run("Dispatch query job with RunningTimeout and read results", func(t *testing.T) {
		check(t, testScenario{
			query:          "SELECT 55 AS num",
			runningTimeout: 10 * time.Minute,
			expectedValue:  55,
		})
	})

	t.Run("Dispatch query job with RunningTimeout verifies JobTimeout configuration on server", func(t *testing.T) {
		jobId := bqzjob.New(fmt.Sprintf("test_job_%d", time.Now().UnixNano()))
		q := &bqz.Query{
			Job:            jobId,
			Query:          "SELECT 55 AS num",
			RunningTimeout: 10 * time.Minute,
		}
		err := c.Dispatch(ctx, q)
		require.NoError(t, err)

		// Verify on the actual BigQuery job configuration that JobTimeout was correctly mapped and applied.
		bqClient, err := bigquery.NewClient(ctx, "i6-rs-contint")
		require.NoError(t, err)
		defer bqClient.Close()

		job, err := bqClient.JobFromID(ctx, jobId.Get())
		require.NoError(t, err)

		config, err := job.Config()
		require.NoError(t, err)

		queryConfig, ok := config.(*bigquery.QueryConfig)
		require.True(t, ok)
		require.Equal(t, 10*time.Minute, queryConfig.JobTimeout)
	})

	t.Run("Dispatch query job with zero/unspecified RunningTimeout defaults to 3 minutes on server", func(t *testing.T) {
		jobId := bqzjob.New(fmt.Sprintf("test_job_%d", time.Now().UnixNano()))
		q := &bqz.Query{
			Job:   jobId,
			Query: "SELECT 55 AS num",
			// RunningTimeout is zero / unspecified
		}
		err := c.Dispatch(ctx, q)
		require.NoError(t, err)

		// Verify on the actual BigQuery job configuration that default JobTimeout of 3 minutes was applied.
		bqClient, err := bigquery.NewClient(ctx, "i6-rs-contint")
		require.NoError(t, err)
		defer bqClient.Close()

		job, err := bqClient.JobFromID(ctx, jobId.Get())
		require.NoError(t, err)

		config, err := job.Config()
		require.NoError(t, err)

		queryConfig, ok := config.(*bigquery.QueryConfig)
		require.True(t, ok)
		require.Equal(t, 3*time.Minute, queryConfig.JobTimeout)
	})

	t.Run("JobStatus returns status and error when job fails", func(t *testing.T) {
		jobId := bqzjob.New(fmt.Sprintf("test_job_%d", time.Now().UnixNano()))
		q := &bqz.Query{
			Job:   jobId,
			Query: "SELECT * FROM non_existent_dataset.non_existent_table_xyz_123",
		}
		err := c.Dispatch(ctx, q)
		require.NoError(t, err)

		// Wait for job to complete (which will fail)
		_ = c.WaitFor(ctx, jobId)

		status, err := c.JobStatus(ctx, jobId)
		require.Error(t, err)
		require.Equal(t, bqz.JobStatusDone, status)
	})
}

func TestRemoteTableExistenceAndLifecycle(t *testing.T) {
	ctx := context.Background()
	c := bqzimpl.New(ctx, bqz.ClientOptions{
		Project: "i6-rs-contint",
	})

	ds := bqzdataset.New("testds")
	tbl := bqztable.New("mytemp_test_table")

	// 1. Initial cleanup (Drop if exists, should return no error)
	err := c.DropTable(ctx, ds, tbl)
	require.NoError(t, err)

	// 2. Table should not exist initially
	exists, err := c.TableExists(ctx, ds, tbl)
	require.NoError(t, err)
	require.False(t, exists)

	// 3. Create the table
	type TempSchema struct {
		ID   string `json:"id" bigquery:"id"`
		Name string `json:"name" bigquery:"name"`
	}
	err = c.CreateExternalTable(ctx, &bqz.ExternalTable{
		Dataset: ds,
		Table:   tbl,
		Uri:     "gs://i6-rs-contint-tmp/testds/mytemp_test_table/*",
		Schema:  &TempSchema{},
	})
	require.NoError(t, err)

	// 4. Table should now exist
	exists, err = c.TableExists(ctx, ds, tbl)
	require.NoError(t, err)
	require.True(t, exists)

	// 5. Drop the table
	err = c.DropTable(ctx, ds, tbl)
	require.NoError(t, err)

	// 6. Table should not exist anymore
	exists, err = c.TableExists(ctx, ds, tbl)
	require.NoError(t, err)
	require.False(t, exists)
}
