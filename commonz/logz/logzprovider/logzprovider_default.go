package logzprovider

import "github.com/infinity6-ai/gox/commonz/logz/logzspec"

type Provider func(appender string) logzspec.Logger

var defaultProvider Provider

func init() {
	defaultProvider = func(appender string) logzspec.Logger {
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
