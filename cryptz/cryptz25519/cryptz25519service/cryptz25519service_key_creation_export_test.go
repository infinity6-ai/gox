package cryptz25519service_test

import (
	"context"
	"crypto/ed25519"
	"testing"

	"github.com/infinity6-ai/gox/cryptz/cryptz25519/cryptz25519service"
	"github.com/stretchr/testify/require"
)

func checkKeyCreationAndExport(t *testing.T, service cryptz25519service.Service25519) {
	priv := service.PrivKeyCreate()
	pub := priv.Public().(ed25519.PublicKey)
	require.Contains(t, service.PrivKeyExport(priv), "BEGIN PRIVATE KEY")
	require.Contains(t, service.PubKeyExport(pub), "BEGIN PUBLIC KEY")
}

func TestUnitKeyCreationAndExport(t *testing.T) {
	services := getImplementations(context.Background())
	for _, s1 := range services {
		t.Run(buildName(s1), func(t *testing.T) {
			checkKeyCreationAndExport(t, s1)
		})
	}
}
