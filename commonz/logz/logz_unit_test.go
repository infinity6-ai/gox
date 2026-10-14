package logz_test

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/infinity6-ai/gox/commonz/logz"
	"github.com/infinity6-ai/gox/commonz/logz/logzaudit"
	"github.com/infinity6-ai/gox/commonz/logz/logzctx"
	"github.com/infinity6-ai/gox/commonz/logz/logzprovider"
	"github.com/infinity6-ai/gox/commonz/logz/logzprovider/logzproviderstructured"
	"github.com/infinity6-ai/gox/commonz/logz/logzspec"
)

type tlogger logz.Type

func TestUnitLoggerCreate(t *testing.T) {
	type invalidLogger bool

	type testScenario struct {
		name        string
		input       any
		expectPanic bool
		panicMsg    string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		if s.expectPanic {
			require.PanicsWithValue(t, s.panicMsg, func() {
				logz.Create(s.input)
			})
			return
		}

		logger := logz.Create(s.input)
		require.NotNil(t, logger, "Logger instance should not be nil")
		require.Equal(t, "github.com/infinity6-ai/gox/commonz/logz_test", logger.Appender())
	}

	t.Run("Valid tlogger creates logger with correct appender", func(t *testing.T) {
		check(t, testScenario{
			name:        "Valid tlogger",
			input:       tlogger(true),
			expectPanic: false,
		})
	})

	t.Run("Invalid type name panics with type error", func(t *testing.T) {
		check(t, testScenario{
			name:        "Invalid type name",
			input:       invalidLogger(true),
			expectPanic: true,
			panicMsg:    "appender type must be tlogger, but was: invalidLogger",
		})
	})

	t.Run("Logger boolean value false panics", func(t *testing.T) {
		check(t, testScenario{
			name:        "False logger value",
			input:       tlogger(false),
			expectPanic: true,
			panicMsg:    "logger must be true",
		})
	})
}

func TestUnitLoggerLevels(t *testing.T) {
	type testScenario struct {
		name          string
		logAction     func(ctx context.Context, logger logz.Logger)
		expectedLevel logzspec.Level
		expectedOp    string
		expectedErr   string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		ctx := logzctx.Prepare(t.Context())
		logger := logz.Create(tlogger(true))
		collector := logger.(logz.Collector).Collector()

		s.logAction(ctx, logger)

		entries := collector()
		require.Len(t, entries, 1, "Expected exactly one log entry")
		require.Equal(t, s.expectedLevel, entries[0].Level)
		require.Equal(t, s.expectedOp, entries[0].Operation)
		if s.expectedErr != "" {
			require.Contains(t, entries[0].Error, s.expectedErr)
		} else {
			require.Empty(t, entries[0].Error)
		}
	}

	t.Run("Info logs entry with INFO level and op", func(t *testing.T) {
		check(t, testScenario{
			name: "Info level",
			logAction: func(ctx context.Context, l logz.Logger) {
				l.Info(ctx, "info-operation", nil)
			},
			expectedLevel: logzspec.INFO,
			expectedOp:    "info-operation",
		})
	})

	t.Run("Debug logs entry with DEBUG level and op", func(t *testing.T) {
		check(t, testScenario{
			name: "Debug level",
			logAction: func(ctx context.Context, l logz.Logger) {
				l.Debug(ctx, "debug-operation", nil)
			},
			expectedLevel: logzspec.DEBUG,
			expectedOp:    "debug-operation",
		})
	})

	t.Run("Error logs entry with ERROR level and joined error messages", func(t *testing.T) {
		check(t, testScenario{
			name: "Error level",
			logAction: func(ctx context.Context, l logz.Logger) {
				l.Error(ctx, "error-operation", nil, errors.New("err one"), errors.New("err two"))
			},
			expectedLevel: logzspec.ERROR,
			expectedOp:    "error-operation",
			expectedErr:   "err one\nerr two",
		})
	})
}

func TestUnitLoggerParams(t *testing.T) {
	type testScenario struct {
		name           string
		params         map[string]any
		expectedParams map[string]string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		ctx := logzctx.Prepare(t.Context())
		logger := logz.Create(tlogger(true))
		collector := logger.(logz.Collector).Collector()

		logger.Info(ctx, "param-test-op", s.params)

		entries := collector()
		require.Len(t, entries, 1)
		require.Equal(t, s.expectedParams, entries[0].Params)
	}

	t.Run("Primitive types are JSON formatted as string values", func(t *testing.T) {
		check(t, testScenario{
			name: "Primitives",
			params: map[string]any{
				"str":  "hello",
				"num":  42,
				"flag": true,
			},
			expectedParams: map[string]string{
				"str":  `"hello"`,
				"num":  "42",
				"flag": "true",
			},
		})
	})

	t.Run("Nil params are omitted and error params are formatted with Error prefix", func(t *testing.T) {
		check(t, testScenario{
			name: "Nil and Error",
			params: map[string]any{
				"omitted": nil,
				"err":     errors.New("db timeout"),
			},
			expectedParams: map[string]string{
				"err": `"Error: db timeout"`,
			},
		})
	})
}

func TestUnitLoggerCollector(t *testing.T) {
	type testScenario struct {
		name string
	}

	t.Run("Entries logged before Collector initialization are not recorded", func(t *testing.T) {
		ctx := logzctx.Prepare(t.Context())
		logger := logz.Create(tlogger(true))

		logger.Info(ctx, "before-collector", nil)

		collector := logger.(logz.Collector).Collector()
		logger.Info(ctx, "after-collector", nil)

		entries := collector()
		require.Len(t, entries, 1)
		require.Equal(t, "after-collector", entries[0].Operation)
	})

	t.Run("Collector drains entries on retrieval", func(t *testing.T) {
		ctx := logzctx.Prepare(t.Context())
		logger := logz.Create(tlogger(true))
		collector := logger.(logz.Collector).Collector()

		logger.Info(ctx, "first", nil)
		logger.Info(ctx, "second", nil)

		entries := collector()
		require.Len(t, entries, 2)
		require.Equal(t, "first", entries[0].Operation)
		require.Equal(t, "second", entries[1].Operation)

		// Subsequent call should be drained and empty
		require.Empty(t, collector())
	})

	t.Run("Collector retains at most 5 entries in ring buffer", func(t *testing.T) {
		ctx := logzctx.Prepare(t.Context())
		logger := logz.Create(tlogger(true))
		collector := logger.(logz.Collector).Collector()

		logger.Info(ctx, "entry-1", nil)
		logger.Info(ctx, "entry-2", nil)
		logger.Info(ctx, "entry-3", nil)
		logger.Info(ctx, "entry-4", nil)
		logger.Info(ctx, "entry-5", nil)
		logger.Info(ctx, "entry-6", nil)
		logger.Info(ctx, "entry-7", nil)

		entries := collector()
		require.Len(t, entries, 5)
		require.Equal(t, "entry-3", entries[0].Operation)
		require.Equal(t, "entry-4", entries[1].Operation)
		require.Equal(t, "entry-5", entries[2].Operation)
		require.Equal(t, "entry-6", entries[3].Operation)
		require.Equal(t, "entry-7", entries[4].Operation)
	})
}

func TestUnitLogzCtx(t *testing.T) {
	type testScenario struct {
		name        string
		action      func(t *testing.T)
		expectPanic bool
		panicMsg    string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		if s.expectPanic {
			require.PanicsWithValue(t, s.panicMsg, func() {
				s.action(t)
			})
			return
		}
		s.action(t)
	}

	t.Run("Get on unprepared context returns nil", func(t *testing.T) {
		check(t, testScenario{
			name: "Get unprepared",
			action: func(t *testing.T) {
				t.Helper()
				require.Nil(t, logzctx.Get(context.Background()))
			},
			expectPanic: false,
		})
	})

	t.Run("Prepare initializes empty context attributes map", func(t *testing.T) {
		check(t, testScenario{
			name: "Prepare fresh context",
			action: func(t *testing.T) {
				t.Helper()
				ctx := logzctx.Prepare(context.Background())
				attrs := logzctx.Get(ctx)
				require.NotNil(t, attrs)
				require.Empty(t, attrs)
			},
			expectPanic: false,
		})
	})

	t.Run("Prepare copies existing attributes to new context without mutating original", func(t *testing.T) {
		check(t, testScenario{
			name: "Prepare copies and isolates attributes",
			action: func(t *testing.T) {
				t.Helper()
				parentCtx := logzctx.Prepare(context.Background())
				logz.SetAttr(parentCtx, "parentKey", "parentVal")

				childCtx := logzctx.Prepare(parentCtx)
				childAttrs := logzctx.Get(childCtx)
				require.NotNil(t, childAttrs)
				require.Equal(t, "parentVal", childAttrs["parentKey"])

				// Mutating child context does not affect parent context
				logz.SetAttr(childCtx, "childKey", "childVal")
				require.Equal(t, "childVal", logzctx.Get(childCtx)["childKey"])
				require.NotContains(t, logzctx.Get(parentCtx), "childKey")
			},
			expectPanic: false,
		})
	})

	t.Run("SetAttr adds a single key-value attribute", func(t *testing.T) {
		check(t, testScenario{
			name: "SetAttr single",
			action: func(t *testing.T) {
				t.Helper()
				ctx := logzctx.Prepare(context.Background())
				logz.SetAttr(ctx, "env", "production")

				attrs := logzctx.Get(ctx)
				require.Equal(t, "production", attrs["env"])
			},
			expectPanic: false,
		})
	})

	t.Run("SetAttrs adds multiple key-value attributes", func(t *testing.T) {
		check(t, testScenario{
			name: "SetAttrs map",
			action: func(t *testing.T) {
				t.Helper()
				ctx := logzctx.Prepare(context.Background())
				logz.SetAttrs(ctx, map[string]string{
					"service": "billing",
					"region":  "us-east-1",
				})

				attrs := logzctx.Get(ctx)
				require.Equal(t, "billing", attrs["service"])
				require.Equal(t, "us-east-1", attrs["region"])
			},
			expectPanic: false,
		})
	})

	t.Run("SetAttr panics when context is not prepared", func(t *testing.T) {
		check(t, testScenario{
			name: "SetAttr panic on unprepared context",
			action: func(t *testing.T) {
				t.Helper()
				logz.SetAttr(context.Background(), "key", "val")
			},
			expectPanic: true,
			panicMsg:    "logz not configured",
		})
	})

	t.Run("SetAttrs panics when context is not prepared", func(t *testing.T) {
		check(t, testScenario{
			name: "SetAttrs panic on unprepared context",
			action: func(t *testing.T) {
				t.Helper()
				logz.SetAttrs(context.Background(), map[string]string{"key": "val"})
			},
			expectPanic: true,
			panicMsg:    "logz not configured",
		})
	})
}

func TestUnitLogzCtxPropagation(t *testing.T) {
	type testScenario struct {
		name            string
		setupCtx        func() context.Context
		mutateCtx       func(ctx context.Context)
		expectedContext map[string]string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		ctx := s.setupCtx()
		logger := logz.Create(tlogger(true))
		collector := logger.(logz.Collector).Collector()

		if s.mutateCtx != nil {
			s.mutateCtx(ctx)
		}

		logger.Info(ctx, "ctx-propagation-op", nil)

		entries := collector()
		require.Len(t, entries, 1)
		require.Equal(t, s.expectedContext, entries[0].Context)
	}

	t.Run("Context attributes are propagated to log entries", func(t *testing.T) {
		check(t, testScenario{
			name: "Context attributes included in entry",
			setupCtx: func() context.Context {
				ctx := logzctx.Prepare(context.Background())
				logz.SetAttr(ctx, "reqId", "req-1234")
				return ctx
			},
			expectedContext: map[string]string{"reqId": "req-1234"},
		})
	})

	t.Run("Independent contexts maintain isolated attributes", func(t *testing.T) {
		ctx1 := logzctx.Prepare(context.Background())
		logz.SetAttr(ctx1, "step", "1")

		ctx2 := logzctx.Prepare(context.Background())
		logz.SetAttrs(ctx2, map[string]string{"step": "2", "userId": "user-42"})

		logger := logz.Create(tlogger(true))
		collector := logger.(logz.Collector).Collector()

		logger.Info(ctx1, "step-one", nil)
		logger.Info(ctx2, "step-two", nil)

		entries := collector()
		require.Len(t, entries, 2)
		require.Equal(t, map[string]string{"step": "1"}, entries[0].Context)
		require.Equal(t, map[string]string{"step": "2", "userId": "user-42"}, entries[1].Context)
	})

	t.Run("Child context prepared from parent preserves parent attributes and isolates mutations", func(t *testing.T) {
		parentCtx := logzctx.Prepare(context.Background())
		logz.SetAttr(parentCtx, "parentKey", "parentVal")

		childCtx := logzctx.Prepare(parentCtx)
		logz.SetAttr(childCtx, "childKey", "childVal")

		logger := logz.Create(tlogger(true))
		collector := logger.(logz.Collector).Collector()

		logger.Info(parentCtx, "parent-op", nil)
		logger.Info(childCtx, "child-op", nil)

		entries := collector()
		require.Len(t, entries, 2)
		require.Equal(t, map[string]string{"parentKey": "parentVal"}, entries[0].Context)
		require.Equal(t, map[string]string{"parentKey": "parentVal", "childKey": "childVal"}, entries[1].Context)
	})

	t.Run("Log entries are immutable snapshots and unaffected by subsequent SetAttr", func(t *testing.T) {
		ctx := logzctx.Prepare(context.Background())
		logger := logz.Create(tlogger(true))
		collector := logger.(logz.Collector).Collector()

		logz.SetAttr(ctx, "k", "v1")
		logger.Info(ctx, "op1", nil)

		logz.SetAttr(ctx, "k", "v2")
		logger.Info(ctx, "op2", nil)

		entries := collector()
		require.Len(t, entries, 2)
		require.Equal(t, "v1", entries[0].Context["k"])
		require.Equal(t, "v2", entries[1].Context["k"])
	})

	t.Run("Unprepared context logs entry with nil context", func(t *testing.T) {
		check(t, testScenario{
			name: "Unprepared context has nil context map",
			setupCtx: func() context.Context {
				return context.Background()
			},
			expectedContext: nil,
		})
	})
}

func TestUnitLogzAudit(t *testing.T) {
	type testScenario struct {
		name          string
		auditEntry    *logzaudit.AuditEntryReqResp
		setupContext  func(ctx context.Context)
		expectedOp    string
		expectedLevel logzspec.Level
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		ctx := logzctx.Prepare(t.Context())
		if s.setupContext != nil {
			s.setupContext(ctx)
		}

		logger := logz.Create(tlogger(true))
		auditLogger, ok := logger.(logz.AuditLogger)
		require.True(t, ok, "Logger should implement logz.AuditLogger")

		collector := logger.(logz.Collector).Collector()

		auditLogger.AuditReqResp(ctx, s.auditEntry)

		entries := collector()
		require.Len(t, entries, 1)
		entry := entries[0]

		require.Equal(t, s.expectedOp, entry.Operation)
		require.Equal(t, s.expectedLevel, entry.Level)
		require.Empty(t, entry.Stack, "Audit entries should have an empty stack")
		require.NotNil(t, entry.Audit, "Audit field should not be nil")
		require.Equal(t, s.auditEntry, entry.Audit.ReqResp)

		if s.setupContext != nil {
			require.Equal(t, logzctx.Get(ctx), entry.Context)
		}
	}

	t.Run("AuditReqResp logs request-only audit entry", func(t *testing.T) {
		auditReq := &logzaudit.AuditEntryReqResp{
			Id: "audit-req-only",
			Service: &logzaudit.AuditService{
				Name:    "order-service",
				Version: "1.0.0",
			},
			Req: &logzaudit.AuditRequest{
				CreatedAt: time.Now().UnixMilli(),
				Host:      "localhost:8080",
				Method:    http.MethodGet,
				Path:      "/orders",
				Query:     url.Values{"page": []string{"1"}},
				Headers:   http.Header{"Accept": []string{"application/json"}},
				Body: &logzaudit.AuditBody{
					Size: 0,
				},
			},
		}

		check(t, testScenario{
			name:          "Request only audit",
			auditEntry:    auditReq,
			expectedOp:    "audit-req-resp",
			expectedLevel: logzspec.INFO,
		})
	})

	t.Run("AuditReqResp logs full request and response with context attributes", func(t *testing.T) {
		auditReqResp := &logzaudit.AuditEntryReqResp{
			Id: "audit-full-cycle",
			Service: &logzaudit.AuditService{
				Name:    "auth-service",
				Version: "2.1.0",
			},
			Req: &logzaudit.AuditRequest{
				CreatedAt: time.Now().UnixMilli(),
				Host:      "auth.example.com",
				Method:    http.MethodPost,
				Path:      "/v1/tokens",
				Query:     url.Values{"grant_type": []string{"refresh_token"}},
				Headers:   http.Header{"Content-Type": []string{"application/json"}},
				Body: &logzaudit.AuditBody{
					Size:    20,
					Payload: []byte(`{"refresh":"abc"}`),
				},
			},
			Resp: &logzaudit.AuditResponse{
				CreatedAt: time.Now().UnixMilli(),
				Status:    http.StatusOK,
				Headers:   http.Header{"Content-Type": []string{"application/json"}},
				Body: &logzaudit.AuditBody{
					Size:    35,
					Payload: []byte(`{"token":"xyz","expires_in":3600}`),
				},
			},
		}

		check(t, testScenario{
			name:       "Full request and response audit with context",
			auditEntry: auditReqResp,
			setupContext: func(ctx context.Context) {
				logz.SetAttr(ctx, "session_id", "sess-999")
				logz.SetAttr(ctx, "client_ip", "192.168.1.1")
			},
			expectedOp:    "audit-req-resp",
			expectedLevel: logzspec.INFO,
		})
	})
}

func TestUnitLogzAuditBodySetPayload(t *testing.T) {
	type testScenario struct {
		name        string
		input       []byte
		wantLarge   bool
		wantLen     int
		wantPayload []byte
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		body := &logzaudit.AuditBody{}
		body.SetPayload(s.input)

		require.Equal(t, s.wantLarge, body.Large)
		require.Equal(t, s.wantLen, len(body.Payload))
		require.Equal(t, s.wantPayload, body.Payload)
	}

	t.Run("Small body below threshold is not marked large and kept whole", func(t *testing.T) {
		sample := []byte(`{"hello":"world"}`)
		check(t, testScenario{
			name:        "Small body",
			input:       sample,
			wantLarge:   false,
			wantLen:     len(sample),
			wantPayload: sample,
		})
	})

	t.Run("Empty body is not marked large", func(t *testing.T) {
		check(t, testScenario{
			name:        "Empty body",
			input:       []byte{},
			wantLarge:   false,
			wantLen:     0,
			wantPayload: []byte{},
		})
	})

	t.Run("Body exactly at LargeBodyThreashold is marked large", func(t *testing.T) {
		exact := bytes.Repeat([]byte("A"), logzaudit.LargeBodyThreashold)
		check(t, testScenario{
			name:        "Exact threshold body",
			input:       exact,
			wantLarge:   true,
			wantLen:     logzaudit.LargeBodyThreashold,
			wantPayload: exact,
		})
	})

	t.Run("Large body exceeding LargeBodyThreashold is marked large and truncated", func(t *testing.T) {
		large := bytes.Repeat([]byte("B"), logzaudit.LargeBodyThreashold+2048)
		check(t, testScenario{
			name:        "Large body truncated",
			input:       large,
			wantLarge:   true,
			wantLen:     logzaudit.LargeBodyThreashold,
			wantPayload: large[:logzaudit.LargeBodyThreashold],
		})
	})
}

func TestUnitLoggerStructuredProvider(t *testing.T) {
	origProvider := logzprovider.GetDefaultProvider()
	t.Cleanup(func() {
		logzprovider.SetDefaultProvider(origProvider)
	})

	// Set to structured provider
	logzprovider.SetDefaultProvider(logzproviderstructured.Post)

	// Since structured provider writes to stderr via its own logger,
	// we just ensure it doesn't panic when fully integrated with logz.Logger.
	logger := logz.Create(tlogger(true))
	ctx := logzctx.Prepare(context.Background())
	logz.SetAttr(ctx, "integration", "true")

	require.NotPanics(t, func() {
		logger.Info(ctx, "structured-integration", map[string]any{"key": "value"})
		logger.Error(ctx, "structured-error", nil, errors.New("some error"))
	})
}
