package sampleroutez_test

import (
	"testing"

	"github.com/infinity6-ai/gox/routez/jsclientz"
	"github.com/infinity6-ai/gox/routez/sampleroutez"
	"github.com/stretchr/testify/require"
)

// TestUnitGenerateJsSdk (re)gera o SDK JS de cada Entry registrada em Entries() — não existe
// mais um comando de CLI/go generate para isso, rodar os testes já regenera o gen/ de todo mundo.
func TestUnitGenerateJsSdk(t *testing.T) {
	for _, entry := range sampleroutez.Entries() {
		for _, service := range entry.Services {
			// New() é obrigatório: Services() devolve serviços com Api() nil, só New() aloca de verdade.
			svc := service.New()
			api := svc.Api()

			fnName, code, err := jsclientz.BuildJsFetch(api.ApiSpec(), api.GetDataRefs())
			require.NoError(t, err)
			require.NoError(t, jsclientz.WriteSdk(entry.OutDir, fnName, code))

			if entry.ExampleParams != nil {
				require.NoError(t, jsclientz.WriteExample(entry.OutDir, fnName, entry.ExampleParams))
			}
		}
	}
}
