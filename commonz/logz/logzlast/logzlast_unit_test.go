package logzlast_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/infinity6-ai/gox/commonz/logz/logzlast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockLogger implements the logzlast.Logger interface for testing purposes.
type MockLogger struct {
	AppenderName string
	// You can add fields here to record calls if needed for more complex tests
	ErrorsCalled []struct {
		Ctx    context.Context
		Op     string
		Params map[string]any
		Errs   []error
	}
	InfosCalled []struct {
		Ctx    context.Context
		Op     string
		Params map[string]any
		Errs   []error
	}
	DebugsCalled []struct {
		Ctx    context.Context
		Op     string
		Params map[string]any
		Errs   []error
	}
}

func (m *MockLogger) Error(ctx context.Context, op string, params map[string]any, errs ...error) {
	m.ErrorsCalled = append(m.ErrorsCalled, struct {
		Ctx    context.Context
		Op     string
		Params map[string]any
		Errs   []error
	}{Ctx: ctx, Op: op, Params: params, Errs: errs})
}

func (m *MockLogger) Info(ctx context.Context, op string, params map[string]any, errs ...error) {
	m.InfosCalled = append(m.InfosCalled, struct {
		Ctx    context.Context
		Op     string
		Params map[string]any
		Errs   []error
	}{Ctx: ctx, Op: op, Params: params, Errs: errs})
}

func (m *MockLogger) Debug(ctx context.Context, op string, params map[string]any, errs ...error) {
	m.DebugsCalled = append(m.DebugsCalled, struct {
		Ctx    context.Context
		Op     string
		Params map[string]any
		Errs   []error
	}{Ctx: ctx, Op: op, Params: params, Errs: errs})
}

func (m *MockLogger) Appender() string {
	return m.AppenderName
}

func TestUnitLoggerEventNew(t *testing.T) {
	t.Parallel()
	type testScenario struct {
		name         string
		appenderName string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		mockLogger := &MockLogger{AppenderName: s.appenderName}
		loggerEvent := logzlast.New(mockLogger)

		require.NotNil(t, loggerEvent)
		assert.Equal(t, s.appenderName, loggerEvent.Appender())

		// Verify that the underlying logger is not directly exposed
		// and that Appender() method comes from the wrapper
		_, ok := loggerEvent.(*logzlast.LoggerEvent)
		require.True(t, ok, "logzlast.New should return a *logzlast.LoggerEvent")
	}

	t.Run("Valid Appender Name", func(t *testing.T) {
		check(t, testScenario{
			name:         "valid appender",
			appenderName: "test-appender",
		})
	})

	t.Run("Empty Appender Name", func(t *testing.T) {
		check(t, testScenario{
			name:         "empty appender",
			appenderName: "",
		})
	})
}

func TestUnitLoggerEventRecording(t *testing.T) {
	t.Parallel()
	type testScenario struct {
		name        string
		logFunc     func(logger logzlast.Logger, ctx context.Context, op string, params map[string]any, errs ...error)
		expectedLevel logzlast.Level
		op          string
		params      map[string]any
		errs        []error
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		mockLogger := &MockLogger{AppenderName: "test-appender"}
		loggerEvent := logzlast.New(mockLogger)

		ctx := context.Background()
		s.logFunc(loggerEvent, ctx, s.op, s.params, s.errs...)

		events := loggerEvent.(*logzlast.LoggerEvent).Events()
		require.Len(t, events, 1)

		event := events[0]
		assert.Equal(t, s.expectedLevel, event.Level)
		assert.Equal(t, s.op, event.Op)
		assert.Equal(t, s.params, event.Params)
		assert.Equal(t, s.errs, event.Errs)
	}

	t.Run("Debug event is recorded", func(t *testing.T) {
		check(t, testScenario{
			name:        "debug",
			logFunc:     func(logger logzlast.Logger, ctx context.Context, op string, params map[string]any, errs ...error) { logger.Debug(ctx, op, params, errs...) },
			expectedLevel: logzlast.LevelDebug,
			op:          "debug_op",
			params:      map[string]any{"key": "value"},
			errs:        []error{errors.New("debug error")},
		})
	})

	t.Run("Info event is recorded", func(t *testing.T) {
		check(t, testScenario{
			name:        "info",
			logFunc:     func(logger logzlast.Logger, ctx context.Context, op string, params map[string]any, errs ...error) { logger.Info(ctx, op, params, errs...) },
			expectedLevel: logzlast.LevelInfo,
			op:          "info_op",
			params:      map[string]any{"user_id": 123},
			errs:        nil,
		})
	})

	t.Run("Error event is recorded", func(t *testing.T) {
		check(t, testScenario{
			name:        "error",
			logFunc:     func(logger logzlast.Logger, ctx context.Context, op string, params map[string]any, errs ...error) { logger.Error(ctx, op, params, errs...) },
			expectedLevel: logzlast.LevelError,
			op:          "error_op",
			params:      map[string]any{"status": "failed"},
			errs:        []error{errors.New("critical error"), fmt.Errorf("wrapped error: %w", errors.New("original"))},
		})
	})

	t.Run("Event with no errors", func(t *testing.T) {
		check(t, testScenario{
			name:        "no errors",
			logFunc:     func(logger logzlast.Logger, ctx context.Context, op string, params map[string]any, errs ...error) { logger.Info(ctx, op, params, errs...) },
			expectedLevel: logzlast.LevelInfo,
			op:          "no_error_op",
			params:      map[string]any{"data": "some data"},
			errs:        nil,
		})
	})

	t.Run("Event with empty params", func(t *testing.T) {
		check(t, testScenario{
			name:        "empty params",
			logFunc:     func(logger logzlast.Logger, ctx context.Context, op string, params map[string]any, errs ...error) { logger.Debug(ctx, op, params, errs...) },
			expectedLevel: logzlast.LevelDebug,
			op:          "empty_params_op",
			params:      map[string]any{},
			errs:        nil,
		})
	})
}

func TestUnitLoggerEventCapacity(t *testing.T) {
	t.Parallel()
	mockLogger := &MockLogger{AppenderName: "test-appender"}
	loggerEvent := logzlast.New(mockLogger).(*logzlast.LoggerEvent)
	ctx := context.Background()

	// Add 5 events
	for i := 0; i < 5; i++ {
		loggerEvent.Info(ctx, fmt.Sprintf("op_%d", i), nil)
	}

	events := loggerEvent.Events()
	require.Len(t, events, 5, "Expected 5 events after adding 5")
	assert.Equal(t, "op_0", events[0].Op, "Expected first event to be op_0")
	assert.Equal(t, "op_4", events[4].Op, "Expected last event to be op_4")

	// Add a 6th event, expecting op_0 to be removed
	loggerEvent.Debug(ctx, "op_5", nil)
	events = loggerEvent.Events()
	require.Len(t, events, 5, "Expected 5 events after adding 6th event")
	assert.Equal(t, "op_1", events[0].Op, "Expected first event to be op_1 (op_0 should be gone)")
	assert.Equal(t, "op_5", events[4].Op, "Expected last event to be op_5")

	// Add a 7th event, expecting op_1 to be removed
	loggerEvent.Error(ctx, "op_6", nil, errors.New("test"))
	events = loggerEvent.Events()
	require.Len(t, events, 5, "Expected 5 events after adding 7th event")
	assert.Equal(t, "op_2", events[0].Op, "Expected first event to be op_2 (op_1 should be gone)")
	assert.Equal(t, "op_6", events[4].Op, "Expected last event to be op_6")


}
