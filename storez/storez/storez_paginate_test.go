package storez_test

import (
	"context"
	"testing"

	"github.com/infinity6-ai/gox/storez/storez"
	"github.com/stretchr/testify/assert"
)

func CheckQueryPaginate(t *testing.T, client *storez.StorezClient) {
	ctx := context.Background()
	CleanTable(client, "t")

	rows := []*MyData{
		{Id: "b", Name: "n1", Desc: "db", Scores: []int64{int64(1), int64(2)}},
		{Id: "a", Name: "n2", Desc: "da", Scores: []int64{int64(1), int64(2)}},
		{Id: "c", Name: "n3", Desc: "dc", Scores: []int64{int64(1)}},
	}
	storez.PutAll(ctx, client, "t", rows)

	loadeds := make([]*MyData, 3)
	storez.Paginate(ctx, client, &storez.Query{
		Table:    "t",
		OrderBys: []storez.OrderBy{{"name", false}},
		Limit:    1,
	}, func(idx uint64, row *MyData) {
		loadeds[int(idx)] = row
		client.Delete(ctx, "t", row.Id)
	})
	assert.Equal(t, rows, loadeds)

	assert.Nil(t, client.Get(ctx, "t", "a"))
	assert.Nil(t, client.Get(ctx, "t", "b"))
	assert.Nil(t, client.Get(ctx, "t", "c"))
}
