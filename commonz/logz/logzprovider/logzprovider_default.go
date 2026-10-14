package logzprovider

import (
	"github.com/infinity6-ai/gox/commonz/logz/logzprovider/logzprovidersimple"
	"github.com/infinity6-ai/gox/commonz/logz/logzspec"
)

var defaultProvider logzspec.ProviderLogger = logzprovidersimple.SimpleProvider

func SetDefaultProvider(provider logzspec.ProviderLogger) {
	defaultProvider = provider
}

func GetDefaultProvider() logzspec.ProviderLogger {
	return defaultProvider
}
