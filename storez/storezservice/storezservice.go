package storezservice

import (
	"context"
	"fmt"
	"io"

	"github.com/infinity6-ai/gox/commonz/ioz"
	"github.com/infinity6-ai/gox/storez/storez"
	"github.com/infinity6-ai/gox/storez/storezdatastore"
	"github.com/infinity6-ai/gox/storez/storezfile"
)

type NewStrategy func(ctx context.Context, projectId string, db string, schema *storez.StorezSchema) storez.StorezStrategy

type StorezService struct {
	New NewStrategy
}

var storezServices = map[string]*StorezService{
	"datastore": {
		New: storezdatastore.New,
	},
	"file": {
		New: storezfile.New,
	},
}

func Get(strategy string) *StorezService {
	return storezServices[strategy]
}

func New(ctx context.Context, strategy string, projectId string, db string, schema *storez.StorezSchema) (storez.StorezStrategy, error) {
	ret := Get(strategy)
	if ret == nil {
		return nil, fmt.Errorf("unknown strategy %s", strategy)
	}
	return ret.New(ctx, projectId, db, schema), nil
}

func Register(name string, service *StorezService) io.Closer {
	old := storezServices[name]
	closer := func() {
		storezServices[name] = old
	}
	storezServices[name] = service
	return ioz.CloserV(closer)
}

func Open(ctx context.Context, projectId string, db string, schema *storez.StorezSchema) *storez.StorezClient {
	strategyName := storez.I6StorezStrategyEncoded.ReqDecoded(ctx)
	if strategyName == "" {
		strategyName = "datastore"
	}
	factory, found := storezServices[strategyName]
	if !found {
		panic(fmt.Sprintf("unknown storez strategy %s", strategyName))
	}
	strategy := factory.New(ctx, projectId, db, schema)
	return storez.Strategy(schema, strategy)
}
