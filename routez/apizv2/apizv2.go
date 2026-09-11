package apizv2

import (
	"context"

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
	GetDataRefs() *DataRefs
}

type Handler[T Api] func(ctx context.Context, reqResp T) (int, error)

type Service interface {
	New() Service
	Api() Api
	Handler(ctx context.Context) (int, error)
}
