package cryptz25519service_test

import (
	"context"
	"strings"

	"github.com/infinity6-ai/gox/cryptz/cryptz25519"
	"github.com/infinity6-ai/gox/cryptz/cryptz25519/cryptz25519service"
)

func buildName(services ...cryptz25519service.Service25519) string {
	var ret strings.Builder
	for i, service := range services {
		ret.WriteString(service.Name())
		if i < len(services)-1 {
			ret.WriteString("_")
		}
	}
	return ret.String()
}

func getImplementations(ctx context.Context) []cryptz25519service.Service25519 {
	return []cryptz25519service.Service25519{
		cryptz25519.NewService(),
		cryptz25519.NewServiceOpenssl(ctx),
	}
}
