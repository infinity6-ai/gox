package logzproviderstructured

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"maps"
	"os"
	"strings"
	"time"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/logz/logzaudit"
	"github.com/infinity6-ai/gox/commonz/logz/logzspec"
	"github.com/infinity6-ai/gox/commonz/strconvz"
	"go.opentelemetry.io/otel/trace"
)

var mlogger *log.Logger

type Severity string

const CtxKeyGaeProjectName = "gae.project.name"

const SeverityDebug Severity = "DEBUG"
const SeverityInfo Severity = "INFO"
const SeverityNotice Severity = "NOTICE"
const SeverityError Severity = "ERROR"

type SourceLocationType struct {
	File string `json:"file,omitempty"`
	Line int64  `json:"line,omitempty"`
}

type StructuredLogzInfo struct {
	Id        string `json:"id,omitempty"`
	Operation string `json:"operation,omitempty"`
	Error     string `json:"error,omitempty"`
	Stack     string `json:"stack,omitempty"`
}

type StructuredEntry struct {
	Times          string              `json:"times,omitempty"`
	Severity       Severity            `json:"severity,omitempty"`
	Appender       string              `json:"appender,omitempty"`
	SourceLocation *SourceLocationType `json:"sourceLocation,omitempty"`
	Params         map[string]string   `json:"params,omitempty"`
	Context        map[string]string   `json:"context,omitempty"`
	Message        string              `json:"message,omitempty"`
	LogzInfo       *StructuredLogzInfo `json:"logzInfo,omitempty"`
	Audit          *logzaudit.Audit    `json:"logzAudit,omitempty"`
	Labels         map[string]string   `json:"logging.googleapis.com/labels,omitempty"`
	TraceId        string              `json:"logging.googleapis.com/trace,omitempty"`
	SpanId         string              `json:"logging.googleapis.com/spanId,omitempty"`
}

func init() {
	mlogger = log.New(os.Stderr, "", 0)
}

func parseSeverity(l logzspec.Level) Severity {
	switch l {
	case logzspec.DEBUG:
		return SeverityDebug
	case logzspec.INFO:
		return SeverityInfo
	case logzspec.ERROR:
		return SeverityError
	}
	panic(fmt.Sprintf("Unknown level: %s", l))
}

func parseSourceLocation(entry *logzspec.Entry) *SourceLocationType {
	if entry.Origin == "" {
		return nil
	}
	ret := &SourceLocationType{File: entry.Origin}
	idx := strings.LastIndex(entry.Origin, ":")
	if idx >= 0 {
		ret.File = entry.Origin[0:idx]
		lineStr := entry.Origin[idx+1 : len(entry.Origin)]
		ret.Line = strconvz.MustParseNumber[int64](lineStr)
	}
	return ret
}

func Post(ctx context.Context, entry *logzspec.Entry) {
	entryContext := maps.Clone(entry.Context)
	traceId, spanId := getTraceId(ctx)
	traceFullId := getTraceFullId(traceId, entryContext)

	labels := map[string]string{}
	if entry.Audit != nil {
		labels["logzaudit"] = "true"
	}
	if len(labels) == 0 {
		labels = nil
	}

	ne := &StructuredEntry{
		TraceId:        traceFullId,
		SpanId:         spanId,
		Severity:       parseSeverity(entry.Level),
		Times:          convertTimestamp(entry.CreatedAt),
		Appender:       entry.Appender,
		SourceLocation: parseSourceLocation(entry),
		LogzInfo: &StructuredLogzInfo{
			Id:        entry.Id,
			Operation: entry.Operation,
			Error:     entry.Error,
			Stack:     entry.Stack,
		},
		Labels:  labels,
		Params:  entry.Params,
		Context: entryContext,
		Audit:   entry.Audit,
	}
	msg := entry.Operation
	if entry.Error != "" {
		msg = fmt.Sprintf("%s\n%s\n%s", entry.Operation, entry.Error, entry.Stack)
	}
	ne.Message = msg
	ret, err := json.Marshal(ne)
	errorz.Check(err)
	mlogger.Printf("%s", ret)
}

func getTraceFullId(traceId string, entryContext map[string]string) string {
	if traceId == "" {
		return ""
	}
	projectName := entryContext[CtxKeyGaeProjectName]
	if projectName == "" {
		return traceId
	}
	return fmt.Sprintf("projects/%s/traces/%s", projectName, traceId)
}

func getTraceId(ctx context.Context) (string, string) {
	traceId := ""
	spanId := ""
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.HasTraceID() {
		traceId = spanCtx.TraceID().String()
	}
	if spanCtx.HasSpanID() {
		spanId = spanCtx.SpanID().String()
	}
	return traceId, spanId
}

func convertTimestamp(epochMillis int64) string {
	t := time.UnixMilli(epochMillis).UTC()
	return t.Format("2006-01-02T15:04:05.999999999Z")
}
