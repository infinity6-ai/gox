package storez

import (
	"context"
	"io"

	"github.com/infinity6-ai/gox/storez/internal/storezoption"
	"github.com/infinity6-ai/gox/storez/storezvalidation"
)

type Value struct {
	Value   any  `json:"value"`
	Indexed bool `json:"indexed,omitempty"`
}

type TransactionOption interface {
	storezoption.TransactionOption
}

func WithTransactionOptionMaxAttempts(n int) TransactionOption {
	return &storezoption.TransactionOptionMaxAttempts{MaxAttempts: n}
}

type StorezStrategy interface {
	io.Closer

	GetTableNames(ctx context.Context) []string

	GetAll(ctx context.Context, table string, ids []string) []map[string]*Value
	PutAll(ctx context.Context, table string, rows []map[string]*Value)
	DeleteAll(ctx context.Context, table string, ids []string)

	Query(ctx context.Context, query *Query) (string, []map[string]*Value)

	Transaction(ctx context.Context, callback func(StorezStrategy), opts ...TransactionOption)
}

type StorezClient struct {
	schema   *StorezSchema
	strategy StorezStrategy
}

func (me *StorezClient) Close() {
	me.strategy.Close()
}

func (me *StorezClient) Transaction(ctx context.Context, callback func(client *StorezClient) error, opts ...TransactionOption) {
	me.strategy.Transaction(ctx, func(strategy StorezStrategy) {
		txclient := Strategy(me.schema, strategy)
		callback(txclient)
	}, opts...)
}

func (me *StorezClient) GetAll(ctx context.Context, table string, ids []string) []map[string]any {
	storezvalidation.TableName(table)
	storezvalidation.Id(ids...)
	rows := me.strategy.GetAll(ctx, table, ids)
	ret := rowsFromValue(rows)
	return ret
}

func (me *StorezClient) GetValueAll(ctx context.Context, table string, ids []string) []map[string]*Value {
	return me.strategy.GetAll(ctx, table, ids)
}

func (me *StorezClient) Get(ctx context.Context, table string, id string) map[string]any {
	result := me.GetAll(ctx, table, []string{id})
	if len(result) == 0 {
		return nil
	}
	return result[0]
}

func (me *StorezClient) PutAll(ctx context.Context, table string, rows []map[string]any) {
	storezvalidation.TableName(table)
	storezvalidation.Row(rows...)
	tableSchema := me.schema.GetTable(table)
	rowValues := rows2Value(tableSchema, rows)
	me.strategy.PutAll(ctx, table, rowValues)
}

func (me *StorezClient) Put(ctx context.Context, table string, row map[string]any) {
	me.PutAll(ctx, table, []map[string]any{row})
}

func (me *StorezClient) DeleteAll(ctx context.Context, table string, ids []string) {
	storezvalidation.TableName(table)
	storezvalidation.Id(ids...)
	me.strategy.DeleteAll(ctx, table, ids)
}

func (me *StorezClient) Delete(ctx context.Context, table string, id string) {
	me.DeleteAll(ctx, table, []string{id})
}

func (me *StorezClient) RunValueWithCursor(ctx context.Context, query *Query) (string, []map[string]*Value) {
	if query.Limit == 0 {
		panic("limit must not be zero")
	}
	return me.strategy.Query(ctx, query)
}

func (me *StorezClient) PaginateValue(ctx context.Context, query *Query, callback func(idx uint64, row map[string]*Value)) {
	if query.Limit == 0 {
		panic("limit must not be zero")
	}
	idx := uint64(0)
	for {
		cursor, rows := me.RunValueWithCursor(ctx, query)
		if len(rows) == 0 {
			return
		}
		for _, row := range rows {
			callback(idx, row)
			idx++
		}
		query.StartCursor = cursor
	}
}

func (me *StorezClient) Run(ctx context.Context, query *Query) []map[string]any {
	_, rows := me.RunWithCursor(ctx, query)
	return rows
}

func convertRowValue(row map[string]*Value) map[string]any {
	retRow := map[string]any{}
	for k, v := range row {
		retRow[k] = v.Value
	}
	return retRow
}

func (me *StorezClient) RunWithCursor(ctx context.Context, query *Query) (string, []map[string]any) {
	cursor, rows := me.RunValueWithCursor(ctx, query)
	ret := []map[string]any{}
	for _, row := range rows {
		retRow := convertRowValue(row)
		ret = append(ret, retRow)
	}
	return cursor, ret
}

func (me *StorezClient) Paginate(ctx context.Context, query *Query, callback func(idx uint64, row map[string]any)) {
	me.PaginateValue(ctx, query, func(idx uint64, row map[string]*Value) {
		retRow := convertRowValue(row)
		callback(idx, retRow)
	})
}

func (me *StorezClient) GetTableNames(ctx context.Context) []string {
	return me.strategy.GetTableNames(ctx)
}
