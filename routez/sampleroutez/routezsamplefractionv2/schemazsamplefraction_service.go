package routezsamplefractionv2

import (
	"context"
	"fmt"
	"strconv"

	"github.com/infinity6-ai/gox/routez/apizv2"
)

type FractionService struct {
	api *FractionApi
}

func (f *FractionService) New() apizv2.Service {
	return &FractionService{api: Api()}
}

func (f *FractionService) Api() apizv2.Api {
	return f.api
}

func (f *FractionService) Handler(ctx context.Context) (int, error) {
	f.api.Resp.TraceMessage = "reason: " + f.api.Req.Reason + ", trace: " + f.api.Req.TraceId
	f.api.Resp.Result.Display = fmt.Sprintf(fmt.Sprintf("%%.%df/%%.%df", int(f.api.Req.Precision), int(f.api.Req.Precision)), f.api.Req.Numerator, f.api.Req.Denominator)
	f.api.Resp.Result.Result = strconv.FormatFloat(f.api.Req.Numerator/f.api.Req.Denominator, 'f', f.api.Req.Precision, 64)
	return 201, nil
}

func Service() *FractionService {
	return &FractionService{}
}

func Services() []apizv2.Service {
	return []apizv2.Service{
		any(Service()).(apizv2.Service),
	}
}
