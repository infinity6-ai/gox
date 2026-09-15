package jsclientz

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/infinity6-ai/gox/commonz/filez"
)

// sdkPackageJson marca o diretório gerado como módulo ES, para Node e a IDE resolverem
// import/export corretamente; não depende do schema, por isso é fixo.
const sdkPackageJson = `{
  "type": "module",
  "private": true,
  "description": "SDK gerado a partir de schemaz.Schema/apiz.Api — não editar à mão."
}
`

// exampleJsTemplate embrulha a chamada gerada num exemplo executável, ao lado do SDK.
const exampleJsTemplate = `// @ts-check
import { %[1]s } from './%[1]s.js';

const baseUrl = 'http://localhost:8080';

const result = await %[2]s;

console.log(result.status);
console.log(JSON.stringify(result.body));
`

// WriteSdk grava <dir>/<fnName>.js e <dir>/package.json, criando os diretórios pai se preciso.
func WriteSdk(dir, fnName, code string) error {
	if err := filez.WriteFile(filepath.Join(dir, fnName+".js"), []byte(code)); err != nil {
		return err
	}
	return filez.WriteFile(filepath.Join(dir, "package.json"), []byte(sdkPackageJson))
}

// jsCallExpr renderiza uma chamada JS a fnName, ex.: `samplefraction(baseUrl, {"numerator":10,...})`.
func jsCallExpr(fnName, baseUrlExpr string, params map[string]any) (string, error) {
	paramsJson, err := json.Marshal(params)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s(%s, %s)", fnName, baseUrlExpr, paramsJson), nil
}

// WriteExample grava <dir>/example.js chamando fnName com os params de amostra.
func WriteExample(dir, fnName string, params map[string]any) error {
	call, err := jsCallExpr(fnName, "baseUrl", params)
	if err != nil {
		return err
	}
	example := fmt.Sprintf(exampleJsTemplate, fnName, call)
	return filez.WriteFile(filepath.Join(dir, "example.js"), []byte(example))
}
