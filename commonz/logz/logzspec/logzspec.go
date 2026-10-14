package logzspec

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/google/uuid"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/jsonz"
	"github.com/infinity6-ai/gox/commonz/logz/logzaudit"
)

type ProviderLogger func(ctx context.Context, entry *Entry)

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
	Audit     *logzaudit.Audit  `json:"prm,omitempty"`
	Params    map[string]string `json:"ctx,omitempty"`
	Context   map[string]string `json:"err,omitempty"`
	Error     string            `json:"stk,omitempty"`
	Stack     string            `json:"aud,omitempty"`
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
	id, err := uuid.NewV7()
	errorz.Check(err)
	entry.Id = base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(id[:])
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
		result[key] = jsonz.MustFormat(value).String()
	}

	return result
}

func formatError(err error) string {
	if err == nil {
		return ""
	}
	return fmt.Sprintf("Error: %s", err.Error())
}
