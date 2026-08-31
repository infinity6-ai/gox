package cryptzrsa

import (
	"context"

	"github.com/infinity6-ai/gox/cryptz/cryptzrsa/cryptzrsaimpl"
	"github.com/infinity6-ai/gox/cryptz/cryptzrsa/cryptzrsaopenssl"
	"github.com/infinity6-ai/gox/cryptz/cryptzrsa/cryptzrsaservice"
)

// NewService returns a new RSAService that uses the native Go crypto library.
func NewService() cryptzrsaservice.RSAService {
	return cryptzrsaimpl.New()
}

// NewServiceOpenssl returns a new RSAService that uses openssl commands.
func NewServiceOpenssl(ctx context.Context) cryptzrsaservice.RSAService {
	return cryptzrsaopenssl.New(ctx)
}
