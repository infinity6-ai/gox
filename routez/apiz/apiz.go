package apiz

import (
	"context"
	"net/http"

	"github.com/infinity6-ai/gox/schemaz/schemaz"
)

type DataRefs struct {
	PathParams  *schemaz.Schema
	QueryParams *schemaz.Schema
	ReqHeaders  *schemaz.Schema
	ReqBody     *schemaz.Schema
	RespHeaders *schemaz.Schema
	RespBody    *schemaz.Schema
}

type ApiSpec struct {
	Id     string
	Desc   func() *schemaz.Desc
	Method string
	Path   string
}

type Api interface {
	ApiSpec() ApiSpec
	ApiDataRefs() *DataRefs
}

type Resp struct {
	Status  int
	Headers http.Header
}

type Service interface {
	New() Service
	Api() Api
	Handler(ctx context.Context) *Resp
}

type ServiceT[A Api] interface {
	Service
	SetApi(api A)
}

func Do[S ServiceT[A], A Api](ctx context.Context, service S, reqResp A) *Resp {
	service = service.New().(S)
	service.SetApi(reqResp)
	reqResp.ApiDataRefs()

	ret := service.Handler(ctx)
	if ret == nil {
		ret = &Resp{Status: 200}
	}
	if ret.Status == 0 {
		ret.Status = 200
	}
	return ret
}
