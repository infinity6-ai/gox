package apiclientz

import (
	"github.com/infinity6-ai/gox/httpz/client/httpzclient"
	"github.com/infinity6-ai/gox/routez/apiz"
)

func Get[P any, Q any, IH any, IB any, OH any, OB any](client *httpzclient.Client, api *apiz.Api[P, Q, IH, IB, OH, OB]) apiz.Handler[P, Q, IH, IB, OH, OB] {
	panic("implement me")
}
