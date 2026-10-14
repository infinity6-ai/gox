package logzprovider

import (
	"os"

	"github.com/infinity6-ai/gox/commonz/logz/logzprovider/logzprovidersimple"
	"github.com/infinity6-ai/gox/commonz/logz/logzprovider/logzproviderstructured"
	"github.com/infinity6-ai/gox/commonz/logz/logzspec"
)

var defaultProvider logzspec.ProviderLogger

func init() {
	structured := os.Getenv("I6_LOGZ_STRUCTURED") == "true"
	if structured {
		defaultProvider = logzproviderstructured.Post
	} else {
		defaultProvider = logzprovidersimple.Post
	}
}

func SetDefaultProvider(provider logzspec.ProviderLogger) {
	defaultProvider = provider
}

func GetDefaultProvider() logzspec.ProviderLogger {
	return defaultProvider
}
