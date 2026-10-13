// Package logzaudit provides structures and constants for defining audit log entries,
// specifically focusing on HTTP request and response auditing. These structures are
// used to capture and format detailed information about service interactions for
// auditing and logging purposes.
package logzaudit

import (
	"net/http"
	"net/url"
)

// LargeBodyThreashold defines the maximum size in bytes (64KB) of the head
// portion of a request or response body. If the body exceeds this threshold,
// it is flagged as large.
const LargeBodyThreashold = 64 * 1024

// TailBodyThreashold defines the maximum size in bytes (1KB) of the tail
// snippet extracted from the bottom (end) of the non-head content when the body
// is large (exceeds LargeBodyThreashold).
const TailBodyThreashold = 1 * 1024

// AuditService represents the identity and version of the service
// involved in the audited interaction.
type AuditService struct {
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
}

// AuditBody encapsulates the details of a request or response body being audited.
type AuditBody struct {
	// Size is the total uncompressed size of the body in bytes.
	Size int `json:"size"`
	// Large indicates whether the total body size exceeds LargeBodyThreashold.
	Large bool `json:"large"`
	// CompressedSize is the size of the compressed body in bytes, if applicable.
	CompressedSize int `json:"compressedSize"`
	// Tail contains up to TailBodyThreashold bytes sliced from the bottom (end)
	// of the remaining non-head content when the body exceeds LargeBodyThreashold.
	// It is omitted if the body is not large.
	Tail []byte `json:"tail,omitempty"`
	// Payload is always the head of the body, limited to LargeBodyThreashold bytes.
	// It exists and is populated unless no content exists.
	Payload []byte `json:"payload,omitempty"`
}

func (a *AuditBody) SetPayload(head []byte) {
	if len(head) >= LargeBodyThreashold {
		a.Large = true
		a.Payload = head[:LargeBodyThreashold]
	} else {
		a.Large = false
		a.Payload = head
	}
}

// AuditResponse captures the details of an HTTP response for auditing,
// including its status code, headers, body information, and the time
// it was created.
type AuditResponse struct {
	CreatedAt int64       `json:"createdAt,omitempty"`
	Status    int         `json:"status,omitempty"`
	Headers   http.Header `json:"headers,omitempty"`
	Body      *AuditBody  `json:"body,omitempty"`
}

// AuditRequest captures the details of an HTTP request for auditing,
// including the host, method, path, query parameters, headers,
// body information, and the time it was received or created.
type AuditRequest struct {
	CreatedAt int64       `json:"createdAt,omitempty"`
	Host      string      `json:"host,omitempty"`
	Method    string      `json:"method,omitempty"`
	Path      string      `json:"path,omitempty"`
	Query     url.Values  `json:"query,omitempty"`
	Headers   http.Header `json:"headers,omitempty"`
	Body      *AuditBody  `json:"body,omitempty"`
}

// AuditEntryReqResp represents a complete audited interaction, tying
// together a unique identifier, the service information, the incoming
// request, and the outgoing response.
type AuditEntryReqResp struct {
	Id      string         `json:"id,omitempty"`
	Service *AuditService  `json:"service,omitempty"`
	Req     *AuditRequest  `json:"req,omitempty"`
	Resp    *AuditResponse `json:"resp,omitempty"`
}

// Audit serves as a top-level wrapper for different types of audit
// records. Currently, it supports request/response audit entries.
type Audit struct {
	ReqResp *AuditEntryReqResp `json:"reqResp,omitempty"`
}
