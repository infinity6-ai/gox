package storez_test

import (
	"context"
	"testing"

	"github.com/infinity6-ai/gox/storez/storez"
	"github.com/stretchr/testify/assert"
)

func CheckQuery(t *testing.T, client *storez.StorezClient) {
	CheckQueryClient(t, client)
	CheckQueryBasics(t, client)
	CheckQueryLimit(t, client)
	CheckQuerySlice(t, client)

	CheckQueryPaginate(t, client)
}

func CheckQueryBasics(t *testing.T, client *storez.StorezClient) {
	ctx := context.Background()
	storez.PutAll(ctx, client, "t", []*MyData{
		{Id: "a", Name: "na", Desc: "da", Scores: []int64{int64(1), int64(2)}},
		{Id: "b", Name: "nb", Desc: "db", Scores: []int64{int64(1), int64(2)}},
		{Id: "c", Name: "nc", Desc: "dc", Scores: []int64{int64(1)}},
	})

	result := storez.Run[*MyData](ctx, client, &storez.Query{
		Table:    "t",
		Filters:  []storez.Filter{{"name", ">=", "nb"}},
		OrderBys: []storez.OrderBy{{"name", true}},
		Limit:    10,
	})
	assert.Equal(t, &MyData{Id: "c", Name: "nc", Desc: "dc", Scores: []int64{int64(1)}}, result[0])
	assert.Equal(t, &MyData{Id: "b", Name: "nb", Desc: "db", Scores: []int64{int64(1), int64(2)}}, result[1])
	assert.Len(t, result, 2)
}

func CheckQuerySlice(t *testing.T, client *storez.StorezClient) {
	ctx := context.Background()
	storez.PutAll(ctx, client, "t", []*MyData{
		{Id: "a", Name: "na", Desc: "da", Scores: []int64{int64(1), int64(2)}},
		{Id: "b", Name: "nb", Desc: "db", Scores: []int64{int64(1), int64(2)}},
		{Id: "c", Name: "nc", Desc: "dc", Scores: []int64{int64(1)}},
	})

	result := storez.Run[*MyData](ctx, client, &storez.Query{
		Table: "t",
		Filters: []storez.Filter{
			{"scores", "=", int64(2)},
		},
		OrderBys: []storez.OrderBy{{"name", true}},
		Limit:    10,
	})
	assert.Equal(t, &MyData{Id: "b", Name: "nb", Desc: "db", Scores: []int64{int64(1), int64(2)}}, result[0])
	assert.Equal(t, &MyData{Id: "a", Name: "na", Desc: "da", Scores: []int64{int64(1), int64(2)}}, result[1])
	assert.Len(t, result, 2)
}

func CheckQueryLimit(t *testing.T, client *storez.StorezClient) {
	ctx := context.Background()
	storez.PutAll(ctx, client, "t", []*MyData{
		{Id: "a", Name: "na", Desc: "da", Scores: []int64{int64(1), int64(2)}},
		{Id: "b", Name: "nb", Desc: "db", Scores: []int64{int64(1), int64(2)}},
		{Id: "c", Name: "nc", Desc: "dc", Scores: []int64{int64(1)}},
	})

	result := storez.Run[*MyData](ctx, client, &storez.Query{
		Table: "t",
		Filters: []storez.Filter{
			{"id", ">=", "a"},
			{"name", ">=", "a"},
		},
		OrderBys: []storez.OrderBy{{"id", true}},
		Limit:    2,
	})
	assert.Equal(t, &MyData{Id: "c", Name: "nc", Desc: "dc", Scores: []int64{int64(1)}}, result[0])
	assert.Equal(t, &MyData{Id: "b", Name: "nb", Desc: "db", Scores: []int64{int64(1), int64(2)}}, result[1])
	assert.Len(t, result, 2)
}
