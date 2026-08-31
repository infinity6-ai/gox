package cryptzrsaservice_test

import (
	"context"
	"testing"

	"github.com/infinity6-ai/gox/cryptz/cryptzrsa/cryptzrsaservice"
	"github.com/stretchr/testify/assert"
)

func checkSignAndVerify(t *testing.T, sSign, sVerify cryptzrsaservice.RSAService) {
	priv := sSign.PrivKeyCreate(2048)
	pub := &priv.PublicKey

	sign := sSign.Sign(priv, []byte("mytext"))
	assert.Len(t, sign, 256)

	assert.True(t, sVerify.Verify(pub, []byte("mytext"), sign))
	assert.False(t, sVerify.Verify(pub, []byte("any"), sign))
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
