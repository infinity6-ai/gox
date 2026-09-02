package apidesczsamplearea

import "github.com/infinity6-ai/gox/commonz/errorz"

// POST /api/encode/[alg]?w=80

type AreaInput struct {
	Alg       string `json:"alg"`
	Precision int    `json:"precision"`
}

func Bla() {
	errorz.Check(nil)
}
