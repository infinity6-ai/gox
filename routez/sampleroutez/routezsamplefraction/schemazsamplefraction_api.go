package routezsamplefraction

import (
	"context"
	"fmt"
	"strconv"

	"github.com/infinity6-ai/gox/routez/apiz"
)

type Result struct {
	Display string `json:"display"`
	Result  string `json:"result"`
}

type FractionReq struct {
	Numerator   float64 `json:"numerator"`
	Denominator float64 `json:"denominator"`
	Precision   int     `json:"precision"`
	TraceId     string  `json:"x-i6-trace-id"`
	Reason      string  `json:"reason"`
}

type FractionResp struct {
	TraceMessage string  `json:"x-i6-trace-message"`
	Result       *Result `json:"result"`
}

type FractionApi struct {
	Req  *FractionReq
	Resp *FractionResp
}

func Spec() *apiz.Spec[*FractionApi] {
	return &apiz.Spec[*FractionApi]{
		Id:     "samplefraction",
		Desc:   nil,
		Method: "POST",
		Path:   "/api/gox/routez/sample/fraction/{numerator}/{denominator}",
		Spec:   &FractionApi{},
	}
}

func Service() *apiz.Service[*FractionApi] {
	return &apiz.Service[*FractionApi]{
		Spec: Spec(),
		Handler: func(ctx context.Context, reqResp *FractionApi) (int, error) {
			reqResp.Resp.TraceMessage = "reason: " + reqResp.Req.Reason + ", trace: " + reqResp.Req.TraceId
			reqResp.Resp.Result.Display = fmt.Sprintf(fmt.Sprintf("%%.%df/%%.%df", int(reqResp.Req.Precision), int(reqResp.Req.Precision)), reqResp.Req.Numerator, reqResp.Req.Denominator)
			reqResp.Resp.Result.Result = strconv.FormatFloat(reqResp.Req.Numerator/reqResp.Req.Denominator, 'f', reqResp.Req.Precision, 64)
			return 201, nil
		},
	}
}
