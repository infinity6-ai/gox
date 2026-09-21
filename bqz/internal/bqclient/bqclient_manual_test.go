package bqclient_test

import (
	"context"
	"testing"

	"github.com/infinity6-ai/gox/bqz/bqz"
	"github.com/infinity6-ai/gox/bqz/internal/bqclient"
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

	c := bqclient.New(ctx, bqclient.ClientOptions{
		Project: "i6-rs-contint",
	})

	err := c.CreateExternalTable(ctx, bqz.ExternalTable{
		Dataset:   "testds",
		Table:     "mytable",
		Uri:       "gs://i6-rs-contint-tmp/testds/mytable",
		HiveParts: []string{"a", "b"},
		Schema:    &SalesHistory{},
	})
	require.Nil(t, err)

}
