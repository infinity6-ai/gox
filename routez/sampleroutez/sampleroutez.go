package sampleroutez

import (
	"github.com/infinity6-ai/gox/routez/apiz"
	"github.com/infinity6-ai/gox/routez/sampleroutez/routezsamplefraction"
)

// Entry é um grupo de Services cujo SDK JS deve ser gerado no mesmo diretório de saída.
type Entry struct {
	Services []apiz.Service
	// OutDir é relativo ao diretório do pacote sampleroutez (cwd de "go test" aqui).
	OutDir string
	// ExampleParams, se não-nil, gera um example.js chamando o SDK com esses valores.
	ExampleParams map[string]any
}

// Entries lista todas as Apis conhecidas para geração de SDK JS. Cada nova apiz.Api real
// entra aqui como uma nova Entry — não existe descoberta automática (Go não tem plugin
// discovery em runtime), então esta lista é o único ponto de registro.
func Entries() []Entry {
	return []Entry{
		{
			Services: routezsamplefraction.Services(),
			OutDir:   "routezsamplefraction/gen",
			ExampleParams: map[string]any{
				"numerator":     10,
				"denominator":   3,
				"precision":     3,
				"x_i6_trace_id": "xx",
				"reason":        "myreason",
			},
		},
	}
}
