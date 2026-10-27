package cryptz25519

import (
	"context"

	"github.com/infinity6-ai/gox/cryptz/cryptz25519/cryptz25519impl"
	"github.com/infinity6-ai/gox/cryptz/cryptz25519/cryptz25519openssl"
	"github.com/infinity6-ai/gox/cryptz/cryptz25519/cryptz25519service"
)

// NewService returns a new Service25519 that uses the native Go crypto library.
func NewService() cryptz25519service.Service25519 {
	return cryptz25519impl.New()
}

// NewServiceOpenssl returns a new Service25519 that uses openssl commands.
func NewServiceOpenssl(ctx context.Context) cryptz25519service.Service25519 {
	return cryptz25519openssl.New(ctx)
}
