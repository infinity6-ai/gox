package storezservice

import (
	"context"
	"fmt"
	"regexp"

	"github.com/infinity6-ai/gox/commonz/validation/checker"
	"github.com/infinity6-ai/gox/storez/storez"
	"github.com/infinity6-ai/gox/storez/storezdatastore"
	"github.com/infinity6-ai/gox/storez/storezfile"
)

var pattern = regexp.MustCompile("^[a-z0-9]+$")

func Open(ctx context.Context, projectId string, db string, schema *storez.StorezSchema) *storez.StorezClient {
	checker.RegexMatch(pattern, db, "datastore database name")
	strategy := storez.I6StorezStrategyEncoded.ReqDecoded(ctx)
	var strategyImpl storez.StorezStrategy
	switch strategy {
	case "datastore":
		strategyImpl = storezdatastore.Open(ctx, projectId, db)
	case "file":
		strategyImpl = storezfile.Open(ctx, projectId, db)
	}
	if strategyImpl == nil {
		panic(fmt.Sprintf("Unknown I6_STOREZ_STRATEGY: %s", strategy))
	}
	return storez.Strategy(schema, strategyImpl)
}
