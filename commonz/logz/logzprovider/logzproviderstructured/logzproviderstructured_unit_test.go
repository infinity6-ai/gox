package logzproviderstructured

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/infinity6-ai/gox/commonz/logz/logzaudit"
	"github.com/infinity6-ai/gox/commonz/logz/logzspec"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"
)

func TestUnitParseSeverity(t *testing.T) {
	require.Equal(t, SeverityDebug, parseSeverity(logzspec.DEBUG))
	require.Equal(t, SeverityInfo, parseSeverity(logzspec.INFO))
	require.Equal(t, SeverityError, parseSeverity(logzspec.ERROR))
	require.PanicsWithValue(t, "Unknown level: FATAL", func() {
		parseSeverity(logzspec.Level("FATAL"))
	})
}

func TestUnitParseSourceLocation(t *testing.T) {
	t.Run("empty origin", func(t *testing.T) {
		entry := &logzspec.Entry{Origin: ""}
		require.Nil(t, parseSourceLocation(entry))
	})

	t.Run("no colon in origin", func(t *testing.T) {
		entry := &logzspec.Entry{Origin: "file_without_colon.go"}
		loc := parseSourceLocation(entry)
		require.NotNil(t, loc)
		require.Equal(t, "file_without_colon.go", loc.File)
		require.Equal(t, int64(0), loc.Line)
	})

	t.Run("valid file and line", func(t *testing.T) {
		entry := &logzspec.Entry{Origin: "test_file.go:42"}
		loc := parseSourceLocation(entry)
		require.NotNil(t, loc)
		require.Equal(t, "test_file.go", loc.File)
		require.Equal(t, int64(42), loc.Line)
	})

	t.Run("invalid line panics", func(t *testing.T) {
		entry := &logzspec.Entry{Origin: "test_file.go:abc"}
		require.Panics(t, func() {
			parseSourceLocation(entry)
		})
	})
}

func TestUnitConvertTimestamp(t *testing.T) {
	// Let's create a known time.
	// 2023-11-14 22:13:20.123 UTC -> Epoch millis = 1700000000123
	expected := "2023-11-14T22:13:20.123Z"
	got := convertTimestamp(1700000000123)
	require.Equal(t, expected, got)
}

func TestUnitGetTraceId(t *testing.T) {
	t.Run("empty context", func(t *testing.T) {
		ctx := context.Background()
		traceId, spanId := getTraceId(ctx)
		require.Empty(t, traceId)
		require.Empty(t, spanId)
	})

	t.Run("with trace id and span id", func(t *testing.T) {
		tid, err := trace.TraceIDFromHex("4bf92f3577b34da6a3ce929d0e0e4736")
		require.NoError(t, err)
		sid, err := trace.SpanIDFromHex("00f067aa0ba902b7")
		require.NoError(t, err)

		sc := trace.NewSpanContext(trace.SpanContextConfig{
			TraceID: tid,
			SpanID:  sid,
		})
		ctx := trace.ContextWithSpanContext(context.Background(), sc)
		traceId, spanId := getTraceId(ctx)
		require.Equal(t, "4bf92f3577b34da6a3ce929d0e0e4736", traceId)
		require.Equal(t, "00f067aa0ba902b7", spanId)
	})
}

func TestUnitGetTraceFullId(t *testing.T) {
	t.Run("empty traceId", func(t *testing.T) {
		require.Equal(t, "", getTraceFullId("", map[string]string{}))
	})

	t.Run("no project name", func(t *testing.T) {
		require.Equal(t, "12345", getTraceFullId("12345", map[string]string{}))
		require.Equal(t, "12345", getTraceFullId("12345", nil))
	})

	t.Run("with project name", func(t *testing.T) {
		ctx := map[string]string{
			CtxKeyGaeProjectName: "my-project",
		}
		require.Equal(t, "projects/my-project/traces/12345", getTraceFullId("12345", ctx))
	})
}

func TestUnitPost(t *testing.T) {
	// Helper to capture the JSON logged
	captureJSON := func(t *testing.T, ctx context.Context, entry *logzspec.Entry) StructuredEntry {
		t.Helper()
		var buf bytes.Buffer
		orig := mlogger.Writer()
		mlogger.SetOutput(&buf)
		defer mlogger.SetOutput(orig)

		Post(ctx, entry)

		output := buf.String()
		require.NotEmpty(t, output)

		var se StructuredEntry
		err := json.Unmarshal([]byte(output), &se)
		require.NoError(t, err)
		return se
	}

	t.Run("basic info log", func(t *testing.T) {
		ctx := context.Background()
		nowMillis := time.Now().UnixMilli()
		entry := &logzspec.Entry{
			Id:        "log-1",
			CreatedAt: nowMillis,
			Appender:  "test-appender",
			Origin:    "service.go:42",
			Level:     logzspec.INFO,
			Operation: "do-work",
			Params:    map[string]string{"foo": "bar"},
			Context:   map[string]string{"user": "alice"},
		}

		se := captureJSON(t, ctx, entry)

		require.Equal(t, SeverityInfo, se.Severity)
		require.Equal(t, "test-appender", se.Appender)
		require.Equal(t, "service.go", se.SourceLocation.File)
		require.Equal(t, int64(42), se.SourceLocation.Line)
		require.Equal(t, map[string]string{"foo": "bar"}, se.Params)
		require.Equal(t, map[string]string{"user": "alice"}, se.Context)
		require.Equal(t, "do-work", se.Message)
		require.Equal(t, "log-1", se.LogzInfo.Id)
		require.Equal(t, "do-work", se.LogzInfo.Operation)
		require.Empty(t, se.LogzInfo.Error)
		require.Empty(t, se.LogzInfo.Stack)
		require.Nil(t, se.Audit)
		require.Empty(t, se.Labels)
		require.Empty(t, se.TraceId)
		require.Empty(t, se.SpanId)

		expectedTimes := convertTimestamp(nowMillis)
		require.Equal(t, expectedTimes, se.Times)
	})

	t.Run("error log with stack and traces", func(t *testing.T) {
		tid, _ := trace.TraceIDFromHex("4bf92f3577b34da6a3ce929d0e0e4736")
		sid, _ := trace.SpanIDFromHex("00f067aa0ba902b7")
		sc := trace.NewSpanContext(trace.SpanContextConfig{
			TraceID: tid,
			SpanID:  sid,
		})
		ctx := trace.ContextWithSpanContext(context.Background(), sc)

		entry := &logzspec.Entry{
			Id:        "err-2",
			Level:     logzspec.ERROR,
			Operation: "failing-op",
			Error:     "database timeout",
			Stack:     "goroutine 1...",
			Context:   map[string]string{CtxKeyGaeProjectName: "cloud-proj"},
		}

		se := captureJSON(t, ctx, entry)

		require.Equal(t, SeverityError, se.Severity)
		require.Equal(t, "failing-op\ndatabase timeout\ngoroutine 1...", se.Message)
		require.Equal(t, "err-2", se.LogzInfo.Id)
		require.Equal(t, "failing-op", se.LogzInfo.Operation)
		require.Equal(t, "database timeout", se.LogzInfo.Error)
		require.Equal(t, "goroutine 1...", se.LogzInfo.Stack)
		require.Equal(t, "projects/cloud-proj/traces/4bf92f3577b34da6a3ce929d0e0e4736", se.TraceId)
		require.Equal(t, "00f067aa0ba902b7", se.SpanId)
	})

	t.Run("audit log", func(t *testing.T) {
		ctx := context.Background()
		audit := &logzaudit.Audit{
			ReqResp: &logzaudit.AuditEntryReqResp{
				Id: "audit-123",
			},
		}
		entry := &logzspec.Entry{
			Level:     logzspec.INFO,
			Operation: "audit-op",
			Audit:     audit,
		}

		se := captureJSON(t, ctx, entry)

		require.NotNil(t, se.Audit)
		require.Equal(t, "audit-123", se.Audit.ReqResp.Id)
		require.NotNil(t, se.Labels)
		require.Equal(t, "true", se.Labels["logzaudit"])
	})
}
