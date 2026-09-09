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

type StorezService struct {
	New func(ctx context.Context, opts storez.StorezOpenOptions) storez.StorezStrategy
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

func Register(name string, service *StorezService) io.Closer {
	old := storezServices[name]
	closer := func() {
		storezServices[name] = old
	}
	storezServices[name] = service
	return ioz.CloserV(closer)
}

func Open(ctx context.Context, strategy string, opts storez.StorezOpenOptions) *storez.StorezClient {
	factory, found := storezServices[strategy]
	if !found {
		panic(fmt.Sprintf("unknown storez strategy %s", strategy))
	}
	return storez.Strategy(opts.Schema, factory.New(ctx, opts))
}
