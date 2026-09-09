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
	Denumerator float64 `json:"denumerator"`
	Precision   int     `json:"precision"`
	TraceId     string  `json:"x-i6-trace-id"`
	Reason      string  `json:"reason"`
}

type FractionResp struct {
	TraceMessage string  `json:"x-i6-trace-message"`
	Result       *Result `json:"result"`
}

type FractionReqResp struct {
	Req  *FractionReq
	Resp *FractionResp
}

func Api() *apiz.Api[*FractionReqResp] {
	return &apiz.Api[*FractionReqResp]{
		Schema: Schema(),
		Handler: func(ctx context.Context, reqResp *FractionReqResp) (int, error) {

			reqResp.Resp.TraceMessage = "reason: " + reqResp.Req.Reason + ", trace: " + reqResp.Req.TraceId
			reqResp.Resp.Result.Display = fmt.Sprintf(fmt.Sprintf("%%.%df/%%.%df", int(reqResp.Req.Precision), int(reqResp.Req.Precision)), reqResp.Req.Numerator, reqResp.Req.Denumerator)
			reqResp.Resp.Result.Result = strconv.FormatFloat(reqResp.Req.Numerator/reqResp.Req.Denumerator, 'f', reqResp.Req.Precision, 64)

			return 201, nil
		},
	}
}
