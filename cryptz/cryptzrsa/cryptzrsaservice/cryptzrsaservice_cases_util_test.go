package cryptzrsaservice_test

import (
	"context"
	"strings"

	"github.com/infinity6-ai/gox/cryptz/cryptzrsa"
	"github.com/infinity6-ai/gox/cryptz/cryptzrsa/cryptzrsaservice"
)

func buildName(services ...cryptzrsaservice.RSAService) string {
	var ret strings.Builder
	for i, service := range services {
		ret.WriteString(service.Name())
		if i < len(services)-1 {
			ret.WriteString("_")
		}
	}
	return ret.String()
}

func getImplementations(ctx context.Context) []cryptzrsaservice.RSAService {
	return []cryptzrsaservice.RSAService{
		cryptzrsa.NewService(),
		cryptzrsa.NewServiceOpenssl(ctx),
	}
}
