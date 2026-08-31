package cryptzrsaservice_test

import (
	"context"
	"testing"

	"github.com/infinity6-ai/gox/cryptz/cryptzrsa/cryptzrsaservice"
	"github.com/stretchr/testify/assert"
)

func checkKeyCreationAndExport(t *testing.T, service cryptzrsaservice.RSAService) {
	priv := service.PrivKeyCreate(2048)
	pub := &priv.PublicKey
	assert.Contains(t, service.PrivKeyExport(priv), "BEGIN PRIVATE KEY")
	assert.Contains(t, service.PubKeyExport(pub), "BEGIN PUBLIC KEY")
}

func TestUnitKeyCreationAndExport(t *testing.T) {
	services := getImplementations(context.Background())
	for _, s1 := range services {
		t.Run(buildName(s1), func(t *testing.T) {
			checkKeyCreationAndExport(t, s1)
		})
	}
}
