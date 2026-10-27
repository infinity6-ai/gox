package cryptz25519service_test

import (
	"context"
	"crypto/ed25519"
	"testing"

	"github.com/infinity6-ai/gox/cryptz/cryptz25519/cryptz25519service"
	"github.com/stretchr/testify/require"
)

func checkSignAndVerify(t *testing.T, sSign, sVerify cryptz25519service.Service25519) {
	priv := sSign.PrivKeyCreate()
	pub := priv.Public().(ed25519.PublicKey)

	sign := sSign.Sign(priv, []byte("mytext"))
	require.Len(t, sign, 64)

	require.True(t, sVerify.Verify(pub, []byte("mytext"), sign))
	require.False(t, sVerify.Verify(pub, []byte("any"), sign))
}

func TestUnitSignAndVerify(t *testing.T) {
	services := getImplementations(context.Background())
	for _, s1 := range services {
		for _, s2 := range services {
			t.Run(buildName(s1, s2), func(t *testing.T) {
				checkSignAndVerify(t, s1, s2)
			})
		}
	}
}
