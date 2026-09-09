package storez_test

import (
	"context"
	"testing"

	"github.com/infinity6-ai/gox/storez/storez"
	"github.com/infinity6-ai/gox/storez/storezservice"
	"github.com/stretchr/testify/assert"
)

type MyData struct {
	Id     string  `json:"id" datastore:"-"`
	Name   string  `json:"name" datastore:"name"`
	Desc   string  `json:"desc" datastore:"desc,noindex"`
	Scores []int64 `json:"scores" datastore:"scores"`
}

func Open(ctx context.Context, strategy string, projectId string, db string) *storez.StorezClient {
	schema := storez.CreateSchema()
	schema.Index("t", "name", "scores")
	return storezservice.Open(ctx, strategy, storez.StorezOpenOptions{
		ProjectId: projectId,
		Db:        db,
		Schema:    schema,
	})
}

func CheckGetTableNames(t *testing.T, client *storez.StorezClient) {
	client.Put(context.Background(), "t", map[string]any{"id": "a", "name": "na", "scores": []any{int64(1), int64(2)}})
	assert.Equal(t, []string{"t"}, client.GetTableNames(context.Background()))
}

func CheckClient(t *testing.T, client *storez.StorezClient) {
	ctx := context.Background()
	client.Delete(ctx, "t", "a")

	row := map[string]any{"id": "a", "name": "na", "scores": []any{int64(1), int64(2)}}
	assert.Nil(t, client.Get(ctx, "t", "a"))
	client.Put(ctx, "t", row)
	assert.Equal(t, row, client.Get(ctx, "t", "a"))

	assert.Equal(t, []map[string]any{row}, client.GetAll(ctx, "t", []string{"notfound", "a"}))

	client.Delete(ctx, "t", "a")
	assert.Nil(t, client.Get(ctx, "t", "a"))
}

func CheckClientTransaction(t *testing.T, client *storez.StorezClient) {
	ctx := context.Background()
	client.Delete(ctx, "t", "a")

	row := map[string]any{"id": "a", "name": "na"}
	assert.Nil(t, client.Get(ctx, "t", "a"))
	storez.Transaction(ctx, client, func(client *storez.StorezClient) {
		client.Put(ctx, "t", row)
	})
	assert.Equal(t, row, client.Get(ctx, "t", "a"))

	assert.Equal(t, []map[string]any{row}, client.GetAll(ctx, "t", []string{"notfound", "a"}))

	storez.Transaction(ctx, client, func(client *storez.StorezClient) {
		client.Delete(ctx, "t", "a")
	})
	assert.Nil(t, client.Get(ctx, "t", "a"))
}

func CheckQueryClient(t *testing.T, client *storez.StorezClient) {
	ctx := context.Background()
	storez.PutAll(ctx, client, "t", []*MyData{
		{Id: "a", Name: "na", Desc: "da", Scores: []int64{int64(1), int64(2)}},
		{Id: "b", Name: "nb", Desc: "db", Scores: []int64{int64(1), int64(2)}},
		{Id: "c", Name: "nc", Desc: "dc", Scores: []int64{int64(1)}},
	})

	result := client.Run(ctx, &storez.Query{
		Table: "t",
		Filters: []storez.Filter{
			{"scores", ">=", int64(2)},
			{"name", ">=", "nb"},
		},
		OrderBys: []storez.OrderBy{
			{"name", true},
		},
		Limit: 10,
	})
	assert.Equal(t, map[string]any{"id": "b", "name": "nb", "desc": "db", "scores": []any{int64(1), int64(2)}}, result[0])
	assert.Len(t, result, 1)
}

func CheckBasic(t *testing.T, client *storez.StorezClient) {
	ctx := context.Background()
	storez.Delete(ctx, client, "t", "a")

	assert.Nil(t, storez.Get(ctx, client, "t", &MyData{Id: "a"}))
	storez.Put(ctx, client, "t", &MyData{Id: "a", Name: "na", Desc: "na"})
	a := storez.Get(ctx, client, "t", &MyData{Id: "a"})
	assert.Equal(t, &MyData{Id: "a", Name: "na", Desc: "na"}, a)
	storez.Delete(ctx, client, "t", "a")
	assert.Nil(t, storez.Get(ctx, client, "t", &MyData{Id: "a"}))

	storez.Put(ctx, client, "t", &MyData{Id: "a", Name: "na", Desc: "na"})
	assert.NotNil(t, storez.Get(ctx, client, "t", &MyData{Id: "a"}))

	rows := client.Run(ctx, &storez.Query{
		Table: "t",
		Limit: 10,
		Filters: []storez.Filter{
			{"id", "=", "a"},
		},
	})
	assert.Equal(t, map[string]interface{}{"id": "a", "desc": "na", "name": "na", "scores": nil}, rows[0])
	assert.Len(t, rows, 1)
}

func CheckTransaction(t *testing.T, client *storez.StorezClient) {
	ctx := context.Background()
	storez.Delete(ctx, client, "t", "a")

	assert.Nil(t, storez.Get(ctx, client, "t", &MyData{Id: "a"}))
	storez.Transaction(ctx, client, func(client *storez.StorezClient) {
		storez.Put(ctx, client, "t", &MyData{Id: "a", Name: "na", Desc: "na"})
	})
	a := storez.Get(ctx, client, "t", &MyData{Id: "a"})
	assert.Equal(t, &MyData{Id: "a", Name: "na", Desc: "na"}, a)
	storez.Transaction(ctx, client, func(client *storez.StorezClient) {
		storez.Delete(ctx, client, "t", "a")
	})
	assert.Nil(t, storez.Get(ctx, client, "t", &MyData{Id: "a"}))
}

func CleanTable(client *storez.StorezClient, table string) {
	ctx := context.Background()
	for {
		rows := client.Run(ctx, &storez.Query{Table: table, Projections: []string{"id"}, Limit: 1000})
		if len(rows) == 0 {
			return
		}
		ids := []string{}
		for _, row := range rows {
			ids = append(ids, row["id"].(string))
		}
		client.DeleteAll(ctx, table, ids)
	}
}
