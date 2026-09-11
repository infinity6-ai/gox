package routezsamplefraction

import (
	"context"
	"fmt"
	"strconv"

	"github.com/infinity6-ai/gox/routez/apiz"
)

type FractionService struct {
	api *FractionApi
}

func (f *FractionService) New() apiz.Service {
	return &FractionService{api: Api()}
}

func (f *FractionService) Api() apiz.Api {
	return f.api
}

func (f *FractionService) Handler(ctx context.Context) (int, error) {
	f.api.Resp.TraceMessage = "reason: " + f.api.Req.Reason + ", trace: " + f.api.Req.TraceId
	f.api.Resp.Result.Display = fmt.Sprintf(fmt.Sprintf("%%.%df/%%.%df", int(f.api.Req.Precision), int(f.api.Req.Precision)), f.api.Req.Numerator, f.api.Req.Denominator)
	f.api.Resp.Result.Result = strconv.FormatFloat(f.api.Req.Numerator/f.api.Req.Denominator, 'f', f.api.Req.Precision, 64)
	return 201, nil
}

func Services() []apiz.Service {
	return []apiz.Service{
		&FractionService{},
	}
}
