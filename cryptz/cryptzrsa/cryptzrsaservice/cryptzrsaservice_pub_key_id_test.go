package cryptzrsaservice_test

import (
	"context"
	"testing"

	"github.com/infinity6-ai/gox/cryptz/cryptzrsa/cryptzrsaservice"
	"github.com/stretchr/testify/assert"
)

func checkPubKeyId(t *testing.T, service cryptzrsaservice.RSAService) {
	priv := service.PrivKeyCreate(2048)
	pub := &priv.PublicKey
	assert.Regexp(t, "^[A-Z0-9]{16}$", service.PubKeyId(pub))
}

func TestUnitPubKeyId(t *testing.T) {
	services := getImplementations(context.Background())
	for _, s1 := range services {
		t.Run(buildName(s1), func(t *testing.T) {
			checkPubKeyId(t, s1)
		})
	}
}
