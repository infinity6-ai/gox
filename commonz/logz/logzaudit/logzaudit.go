package logzaudit

import (
	"net/http"
	"net/url"
)

const LARGE_BODY_THREASHOLD = 64 * 1024

type AuditService struct {
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
}

type AuditBody struct {
	Size           int    `json:"size"`
	Large          bool   `json:"large"`
	CompressedSize int    `json:"compressedSize"`
	Tail           []byte `json:"tail"`
	Payload        []byte `json:"payload"`
}

type AuditResponse struct {
	CreatedAt int64       `json:"createdAt,omitempty"`
	Status    int         `json:"status,omitempty"`
	Headers   http.Header `json:"headers,omitempty"`
	Body      *AuditBody  `json:"body,omitempty"`
}

type AuditRequest struct {
	CreatedAt int64       `json:"createdAt,omitempty"`
	Host      string      `json:"host,omitempty"`
	Method    string      `json:"method,omitempty"`
	Path      string      `json:"path,omitempty"`
	Query     url.Values  `json:"query,omitempty"`
	Headers   http.Header `json:"headers,omitempty"`
	Body      *AuditBody  `json:"body,omitempty"`
}

type AuditEntryReqResp struct {
	Id      string         `json:"id,omitempty"`
	Service *AuditService  `json:"service,omitempty"`
	Req     *AuditRequest  `json:"req,omitempty"`
	Resp    *AuditResponse `json:"resp,omitempty"`
}

type Audit struct {
	ReqResp *AuditEntryReqResp `json:"reqResp,omitempty"`
}
