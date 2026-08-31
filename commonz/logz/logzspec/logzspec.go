package logzspec

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"time"

	"go.code.infinity6.ai/platform/util/jsoner"
)

type ProviderLogger interface {
	Error(ctx context.Context, entry *Entry)
	Info(ctx context.Context, entry *Entry)
	Debug(ctx context.Context, entry *Entry)
	Appender() string
}

type Level string

const DEBUG Level = "DEBUG"
const INFO Level = "INFO"
const ERROR Level = "ERROR"

type Entry struct {
	Id        string            `json:"lid,omitempty"`
	CreatedAt int64             `json:"cat,omitempty"`
	Appender  string            `json:"apd,omitempty"`
	Origin    string            `json:"org,omitempty"`
	Level     Level             `json:"lvl,omitempty"`
	Operation string            `json:"opr,omitempty"`
	Params    map[string]string `json:"prm,omitempty"`
	Error     string            `json:"err,omitempty"`
	Stack     string            `json:"stk,omitempty"`
}

func NewEntry(originSkil int, appender string, level Level, op string, params map[string]any, errs ...error) *Entry {
	entry := &Entry{
		Appender:  appender,
		Level:     level,
		Operation: op,
		Params:    parseParams(params),
	}
	if len(errs) > 0 {
		entry.Error = formatError(errors.Join(errs...))
	}
	entry.fill(originSkil)
	return entry
}

func (entry *Entry) fill(originSkip int) {
	// entry.Id = uuid.NewRandom() // FIXME: id
	entry.CreatedAt = time.Now().UTC().UnixMilli()
	entry.Origin = getCaller(originSkip)

	stack := debug.Stack()
	entry.Stack = string(stack)

	if entry.Params == nil {
		entry.Params = map[string]string{}
	}
}

func getCaller(skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "unknown:0"
	}
	file = filepath.Base(file)
	return fmt.Sprintf("%s:%d", file, line)
}

func parseParams(params map[string]any) map[string]string {
	result := make(map[string]string)
	for key, value := range params {
		if value == nil {
			continue
		}
		errValue, ok := value.(error)
		if ok {
			value = formatError(errValue)
		}
		result[key] = jsoner.FormatString(value)
	}

	return result
}

func formatError(err error) string {
	if err == nil {
		return ""
	}
	return fmt.Sprintf("Error: %s", err.Error())
}
