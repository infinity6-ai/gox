package storezdatastore

import (
	"context"

	"cloud.google.com/go/datastore"
	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/storez/storez"
)

type StorezStrategyDatastoreTX struct {
	storez.StorezStrategy

	tx                *datastore.Transaction
	datastoreStrategy *StorezStrategyDatastore
}

func (me *StorezStrategyDatastoreTX) Close() error {
	return nil
}

func (me *StorezStrategyDatastoreTX) Transaction(ctx context.Context, callback func(client storez.StorezStrategy), opts ...storez.TransactionOption) {
	callback(me)
}

func (me *StorezStrategyDatastoreTX) GetAll(ctx context.Context, table string, ids []string) []map[string]*storez.Value {
	keys := Keys2DS(table, ids)
	rows := make([]datastore.PropertyList, len(ids))
	err := me.tx.GetMulti(keys, rows)
	ret := HandleMultiError(err, keys, rows)
	return ret
}

func (me *StorezStrategyDatastoreTX) PutAll(ctx context.Context, table string, entities []map[string]*storez.Value) {
	keys, rows := Rows2DS(table, entities)
	_, err := me.tx.PutMulti(keys, rows)
	errorz.Check(err)
}

func (me *StorezStrategyDatastoreTX) DeleteAll(ctx context.Context, table string, ids []string) {
	keys := Keys2DS(table, ids)
	err := me.tx.DeleteMulti(keys)
	errorz.Check(err)
}
