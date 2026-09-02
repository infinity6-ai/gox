package internal

type SampleApiParams struct {
	Numerator   float64
	Denominator float64
}

type SampleApiQuery struct {
	Precision int
}

type ReqMeta struct {
	TraceId string
}

type RespMeta struct {
	Hash string
}
