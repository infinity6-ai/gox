package logzlast

import (
	"context"
	"slices"
	"sync"
)

type Logger interface {
	Error(ctx context.Context, op string, params map[string]any, errs ...error)
	Info(ctx context.Context, op string, params map[string]any, errs ...error)
	Debug(ctx context.Context, op string, params map[string]any, errs ...error)
	Appender() string
}

func New(l Logger) Logger {
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
	logger Logger
	events []*Event
	mu     sync.RWMutex
}

var instance LoggerEvent

func GetInstance() *LoggerEvent {
	return &instance
}

func LastEvents() []*Event {
	return instance.Events()
}

func (l *LoggerEvent) Events() []*Event {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return slices.Clone(l.events)
}

func (l *LoggerEvent) add(event Event) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if len(l.events) >= 5 {
		l.events = l.events[1:]
	}
	l.events = append(l.events, &event)
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
