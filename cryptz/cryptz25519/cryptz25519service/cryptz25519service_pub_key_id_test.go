package cryptz25519service_test

import (
	"context"
	"crypto/ed25519"
	"testing"

	"github.com/infinity6-ai/gox/cryptz/cryptz25519/cryptz25519service"
	"github.com/stretchr/testify/require"
)

func checkPubKeyId(t *testing.T, service cryptz25519service.Service25519) {
	priv := service.PrivKeyCreate()
	pub := priv.Public().(ed25519.PublicKey)
	require.Regexp(t, "^[A-Z0-9]{16}$", service.PubKeyId(pub))
}

func TestUnitPubKeyId(t *testing.T) {
	services := getImplementations(context.Background())
	for _, s1 := range services {
		t.Run(buildName(s1), func(t *testing.T) {
			checkPubKeyId(t, s1)
		})
	}
}
