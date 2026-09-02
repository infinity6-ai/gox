package apidesczsamplearea

import (
	"github.com/infinity6-ai/gox/apidescz/apidescz"
)

// POST /api/encode/base64/{alf}?width=80

type Sample struct {
	Id    string `json:"id"`
	Value int    `json:"value"`
}

type Input struct {
	Alf     string  `json:"alf"`
	Width   int     `json:"width"`
	HashAlg string  `json:"hash_alg"`
	Sample  *Sample `json:"input_data"`
}

type OutputSample struct {
	Sample *Sample `json:"sample"`
	Data   string  `json:"data"`
}

type Output struct {
	Hash        string        `json:"hash"`
	OuputSample *OutputSample `json:"output_sample"`
}

func Bla() {
	sampleApi := &apidescz.Api[*Input, *Output]{
		Name:   "base64",
		Short:  "encode Sample in base64",
		Guide:  "# base64\n\nguide\n\n",
		Path:   "/api/encode/base64/{alf}",
		Input:  &Input{},
		Output: &Output{},
	}
	print(sampleApi)
}
