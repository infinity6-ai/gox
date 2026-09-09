package storez

import (
	"context"

	"github.com/infinity6-ai/gox/storez/storezutil"
	"github.com/infinity6-ai/gox/storez/storezutil/structz"
)

type StorezOpenOptions struct {
	ProjectId string        `json:"project_id,omitempty"`
	Db        string        `json:"db,omitempty"`
	Schema    *StorezSchema `json:"schema,omitempty"`
}

func Strategy(schema *StorezSchema, strategy StorezStrategy) *StorezClient {
	return &StorezClient{schema: schema, strategy: strategy}
}

func Transaction(ctx context.Context, client *StorezClient, callback func(client *StorezClient), opts ...TransactionOption) {
	client.Transaction(ctx, func(client *StorezClient) error {
		callback(client)
		return nil
	}, opts...)
}

func PutAll[T any](ctx context.Context, client *StorezClient, table string, entities []T) {
	rows := structz.StructsToMapList(entities)
	client.PutAll(ctx, table, rows)
}

func DeleteAll(ctx context.Context, client *StorezClient, table string, ids []string) {
	client.DeleteAll(ctx, table, ids)
}

func GetAll[T any](ctx context.Context, client *StorezClient, table string, entities []T) []T {
	ids := storezutil.GetKeys(entities)
	rows := client.GetAll(ctx, table, ids)
	ret := structz.MapListToStructs[T](rows)
	return ret
}

func Put(ctx context.Context, client *StorezClient, table string, entity any) {
	PutAll(ctx, client, table, []any{entity})
}

func Delete(ctx context.Context, client *StorezClient, table string, id string) {
	DeleteAll(ctx, client, table, []string{id})
}

func Get[T any](ctx context.Context, client *StorezClient, table string, entity T) T {
	var ret T
	l := GetAll(ctx, client, table, []T{entity})
	if len(l) > 0 {
		ret = l[0]
	}
	return ret
}

func Run[T any](ctx context.Context, client *StorezClient, query *Query) []T {
	_, ret := RunWithCursor[T](ctx, client, query)
	return ret
}

func RunWithCursor[T any](ctx context.Context, client *StorezClient, query *Query) (string, []T) {
	cursor, mapRes := client.RunWithCursor(ctx, query)
	ret := structz.MapListToStructs[T](mapRes)
	return cursor, ret
}

func Paginate[T any](ctx context.Context, client *StorezClient, query *Query, callback func(idx uint64, row T)) {
	client.Paginate(ctx, query, func(idx uint64, row map[string]any) {
		ret := structz.MapListToStructs[T]([]map[string]any{row})[0]
		callback(idx, ret)
	})
}
