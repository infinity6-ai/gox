package logzlast

import (
	"context"
	"slices"
	"sync"

	"github.com/infinity6-ai/gox/commonz/logz/logzspec"
)

func New(l logzspec.Logger) logzspec.Logger {
	return &LoggerEvent{
		logger: l,
	}
}

type Level string

const (
	LevelDebug Level = "debug"
	LevelInfo  Level = "info"
	LevelError Level = "error"
)

type Event struct {
	Level  Level
	Op     string
	Params map[string]any
	Errs   []error
}

type LoggerEvent struct {
	logger logzspec.Logger
}

var events []*Event
var mu sync.RWMutex

func LastEvents() []*Event {
	mu.RLock()
	defer mu.RUnlock()
	return slices.Clone(events)
}

func (l *LoggerEvent) add(event Event) {
	mu.Lock()
	defer mu.Unlock()
	if len(events) >= 5 {
		events = events[1:]
	}
	events = append(events, &event)
}

func (l *LoggerEvent) Appender() string {
	return l.logger.Appender()
}

func (l *LoggerEvent) Debug(ctx context.Context, op string, params map[string]any, errs ...error) {
	l.add(Event{
		Level:  LevelDebug,
		Op:     op,
		Params: params,
		Errs:   errs,
	})
}

func (l *LoggerEvent) Error(ctx context.Context, op string, params map[string]any, errs ...error) {
	l.add(Event{
		Level:  LevelError,
		Op:     op,
		Params: params,
		Errs:   errs,
	})
}

func (l *LoggerEvent) Info(ctx context.Context, op string, params map[string]any, errs ...error) {
	l.add(Event{
		Level:  LevelInfo,
		Op:     op,
		Params: params,
		Errs:   errs,
	})
}
