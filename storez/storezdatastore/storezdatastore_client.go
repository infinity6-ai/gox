package storezdatastore

import (
	"context"
	"fmt"
	"io"
	"os"

	"cloud.google.com/go/datastore"
	"github.com/infinity6-ai/gox/commonz/configz"
	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/idgen"
	"github.com/infinity6-ai/gox/commonz/ioz"
	"github.com/infinity6-ai/gox/commonz/logz"
	"github.com/infinity6-ai/gox/commonz/strconvz"
	"github.com/infinity6-ai/gox/storez/internal/storezoption"
	"github.com/infinity6-ai/gox/storez/storez"
)

type tlogger logz.Type

var logger = logz.Create(tlogger(true))

var I6StorezTxMaxAttempts = configz.Create("I6_STOREZ_TX_MAX_ATTEMPTS", "")

const ENV_EMULATOR_HOST = "DATASTORE_EMULATOR_HOST"

type StorezStrategyDatastore struct {
	client *datastore.Client
}

func Init(ctx context.Context) io.Closer {
	return errorz.Check2(storez.I6StorezStrategyEncoded.SetEncoded(ctx, "datastore"))
}

func InitEmulator(ctx context.Context, projectId string) (string, io.Closer) {
	if projectId == "" {
		projectId = fmt.Sprintf("demo-%s", idgen.Hex())
	}
	host := "localhost:7005"
	oldenv := os.Getenv(ENV_EMULATOR_HOST)
	os.Setenv(ENV_EMULATOR_HOST, host)
	logger.Info(ctx, "env changed", map[string]any{"key": ENV_EMULATOR_HOST, "value": host})
	init := Init(ctx)
	return projectId, ioz.CloserV(func() {
		init.Close()
		os.Setenv(ENV_EMULATOR_HOST, oldenv)
		logger.Info(ctx, "env reverted", map[string]any{"key": ENV_EMULATOR_HOST, "value": oldenv})
	})
}

func New(ctx context.Context, projectId string, db string, schema *storez.StorezSchema) storez.StorezStrategy {
	return open(ctx, projectId, db)
}

func open(ctx context.Context, projectId string, db string) *StorezStrategyDatastore {
	client, err := datastore.NewClientWithDatabase(ctx, projectId, db)
	errorz.Check(err)
	return &StorezStrategyDatastore{client: client}
}

func (me *StorezStrategyDatastore) Close() error {
	if me.client != nil {
		return me.client.Close()
	}
	return nil
}

func (me *StorezStrategyDatastore) Transaction(ctx context.Context, callback func(storez.StorezStrategy), opts ...storez.TransactionOption) {
	dsOpts := make([]datastore.TransactionOption, len(opts)+1, len(opts)+2)
	maxAttemptsSet := false
	for i, opt := range opts {
		maxAttempts, ok := opt.(*storezoption.TransactionOptionMaxAttempts)
		if ok {
			dsOpts[i] = datastore.MaxAttempts(maxAttempts.MaxAttempts)
			maxAttemptsSet = true
		}
	}
	dsOpts[len(opts)] = datastore.BeginLater
	if !maxAttemptsSet {
		maxAttempts, err := I6StorezTxMaxAttempts.Get(ctx)
		errorz.Check(err)
		if maxAttempts != "" {
			dsOpts = append(dsOpts, datastore.MaxAttempts(strconvz.MustParseNumber[int](maxAttempts)))
		}
	}

	_, err := me.client.RunInTransaction(ctx, func(tx *datastore.Transaction) error {
		strategy := &StorezStrategyDatastoreTX{datastoreStrategy: me, tx: tx}
		defer strategy.Close()
		callback(strategy)
		return nil
	}, dsOpts...)
	errorz.Check(err)
}

func (me *StorezStrategyDatastore) GetAll(ctx context.Context, table string, ids []string) []map[string]*storez.Value {
	keys := Keys2DS(table, ids)
	rows := make([]datastore.PropertyList, len(ids))
	err := me.client.GetMulti(ctx, keys, rows)
	ret := HandleMultiError(err, keys, rows)
	return ret
}

func (me *StorezStrategyDatastore) PutAll(ctx context.Context, table string, entities []map[string]*storez.Value) {
	keys, rows := Rows2DS(table, entities)
	_, err := me.client.PutMulti(ctx, keys, rows)
	errorz.Check(err)
}

func (me *StorezStrategyDatastore) DeleteAll(ctx context.Context, table string, ids []string) {
	keys := Keys2DS(table, ids)
	err := me.client.DeleteMulti(ctx, keys)
	errorz.Check(err)
}

func (me *StorezStrategyDatastore) GetTableNames(ctx context.Context) []string {
	max := 1001
	_, rows := me.Query(ctx, &storez.Query{
		Table:       "__kind__",
		Filters:     []storez.Filter{{Field: "id", Op: ">", Value: "__\ufffd"}},
		OrderBys:    []storez.OrderBy{{Field: "id", Desc: false}},
		Limit:       max,
		Projections: []string{"id"},
	})
	if len(rows) >= max {
		panic(fmt.Errorf("too many tables: %d", len(rows)))
	}
	ret := []string{}
	for _, row := range rows {
		ret = append(ret, row["id"].Value.(string))
	}
	return ret
}
