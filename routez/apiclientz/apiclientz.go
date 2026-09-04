package apiclientz

import (
	"context"

	"github.com/infinity6-ai/gox/httpz/client/httpzclient"
	"github.com/infinity6-ai/gox/routez/apiz"
)

func parseRequest[P any, Q any, IH any, IB any, OH any, OB any](api *apiz.Api[P, Q, IH, IB, OH, OB], req *apiz.Req[P, Q, IH, IB]) *httpzclient.Client {
	// structjsonz.MustParseSingle(params, &p)
	// structjsonz.MustParse(req.Query, &q)
	// structjsonz.MustParse(converter.Header2Json(req.Headers), &reqHeaders)

	// ret := httpzclient.NewReq(api.Schema.Method, api.Schema.Path)
	// ret. := structjsonz.MustFormatSingle(req.PathParams)

	// converter.Json2Header()
}

func Get[P any, Q any, IH any, IB any, OH any, OB any](client *httpzclient.Client, api *apiz.Api[P, Q, IH, IB, OH, OB]) apiz.Handler[P, Q, IH, IB, OH, OB] {
	return func(ctx context.Context, req *apiz.Req[P, Q, IH, IB]) (*apiz.Resp[OH, OB], error) {

		parseRequest(api, req)

		// client.Do()
		panic("X")

	}
}
