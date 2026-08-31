package logzprovider

import "context"

type Logger interface {
	Error(ctx context.Context, op string, params map[string]any, errs ...error)
	Info(ctx context.Context, op string, params map[string]any, errs ...error)
	Debug(ctx context.Context, op string, params map[string]any, errs ...error)
	Appender() string
}

type Provider func(appender string) Logger

var defaultProvider Provider

func init() {
	defaultProvider = func(appender string) Logger {
		if appender == "" {
			panic("appender must not be empty")
		}
		return &SimpleProvider{
			appender: appender,
		}
	}
}

func SetDefaultProvider(provider Provider) {
	defaultProvider = provider
}

func GetDefaultProvider() Provider {
	return defaultProvider
}

type SimpleProvider struct {
	appender string
}

func (s *SimpleProvider) Appender() string {
	return s.appender
}

func (s *SimpleProvider) Debug(ctx context.Context, op string, params map[string]any, errs ...error) {
	panic("unimplemented")
}

func (s *SimpleProvider) Error(ctx context.Context, op string, params map[string]any, errs ...error) {
	panic("unimplemented")
}

func (s *SimpleProvider) Info(ctx context.Context, op string, params map[string]any, errs ...error) {
	panic("unimplemented")
}
