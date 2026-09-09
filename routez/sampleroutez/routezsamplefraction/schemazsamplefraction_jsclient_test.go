package routezsamplefraction_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/infinity6-ai/gox/commonz/filez"
	"github.com/infinity6-ai/gox/httpz/httpzserver"
	"github.com/infinity6-ai/gox/routez/routez"
	"github.com/infinity6-ai/gox/routez/sampleroutez/routezsamplefraction"
	"github.com/infinity6-ai/gox/schemaz/schemaz"
	"github.com/stretchr/testify/require"
)

func jsFieldType(t schemaz.Type) string {
	switch t {
	case schemaz.TypeString:
		return "string"
	case schemaz.TypeNumber:
		return "number"
	case schemaz.TypeBoolean:
		return "boolean"
	case schemaz.TypeArray:
		return "Array"
	case schemaz.TypeObject:
		return "Object"
	default:
		return "*"
	}
}

// jsTypeLiteral renders a schemaz.Spec as an inline JSDoc object type, e.g. "{display: string, result: number}".
func jsTypeLiteral(spec schemaz.Spec) string {
	switch spec.Type {
	case schemaz.TypeObject:
		parts := make([]string, 0, len(spec.Fields))
		for _, f := range spec.Fields {
			parts = append(parts, f.Name+": "+jsTypeLiteral(f.Spec))
		}
		return "{" + strings.Join(parts, ", ") + "}"
	case schemaz.TypeArray:
		return jsFieldType(spec.ArrayType) + "[]"
	default:
		return jsFieldType(spec.Type)
	}
}

// headerName converts a schema field name (e.g. "trace_id") to its HTTP header form ("trace-id").
func headerName(fieldName string) string {
	return strings.ReplaceAll(fieldName, "_", "-")
}

// reqFields lists every field the client must supply to call the api, in schema order.
func reqFields(api *schemaz.Api) []schemaz.Field {
	fields := append([]schemaz.Field{}, api.ReqParams...)
	fields = append(fields, api.ReqQuery...)
	fields = append(fields, api.ReqHeaders...)
	if api.ReqBody != nil {
		fields = append(fields, api.ReqBody.Fields...)
	}
	return fields
}

// respTypeLiteral renders the JSDoc type of the value the generated function resolves to.
func respTypeLiteral(api *schemaz.Api) string {
	parts := []string{"status: number"}
	if len(api.RespHeaders) > 0 {
		headerParts := make([]string, 0, len(api.RespHeaders))
		for _, f := range api.RespHeaders {
			headerParts = append(headerParts, f.Name+": "+jsFieldType(f.Spec.Type))
		}
		parts = append(parts, "headers: {"+strings.Join(headerParts, ", ")+"}")
	}
	if api.RespBody != nil {
		parts = append(parts, "body: "+jsTypeLiteral(*api.RespBody))
	}
	return "{" + strings.Join(parts, ", ") + "}"
}

// buildJsDoc renders the JSDoc comment that gives SDK consumers autocomplete on params and on the resolved value.
func buildJsDoc(api *schemaz.Api) string {
	var b strings.Builder
	b.WriteString("/**\n")
	if api.Desc.Summary != "" {
		fmt.Fprintf(&b, " * %s\n", api.Desc.Summary)
	}
	b.WriteString(" * @param {string} baseUrl\n")
	b.WriteString(" * @param {Object} params\n")
	for _, f := range reqFields(api) {
		if f.Desc.Summary != "" {
			fmt.Fprintf(&b, " * @param {%s} params.%s - %s\n", jsFieldType(f.Spec.Type), f.Name, f.Desc.Summary)
		} else {
			fmt.Fprintf(&b, " * @param {%s} params.%s\n", jsFieldType(f.Spec.Type), f.Name)
		}
	}
	fmt.Fprintf(&b, " * @returns {Promise<%s>}\n", respTypeLiteral(api))
	b.WriteString(" */\n")
	return b.String()
}

// buildJsRequest renders the part of the function body that builds the URL/options and sends the request.
func buildJsRequest(api *schemaz.Api) string {
	var b strings.Builder

	path := api.Path
	for _, f := range api.ReqParams {
		path = strings.ReplaceAll(path, "{"+f.Name+"}", "${params."+f.Name+"}")
	}
	fmt.Fprintf(&b, "  const url = new URL(`${baseUrl}%s`);\n", path)
	for _, f := range api.ReqQuery {
		fmt.Fprintf(&b, "  url.searchParams.set(%q, String(params.%s));\n", f.Name, f.Name)
	}

	b.WriteString("  const headers = {'Content-Type': 'application/json'};\n")
	for _, f := range api.ReqHeaders {
		fmt.Fprintf(&b, "  headers[%q] = String(params.%s);\n", headerName(f.Name), f.Name)
	}
	fmt.Fprintf(&b, "  const options = {method: %q, headers};\n", api.Method)
	if api.ReqBody != nil {
		assignments := make([]string, 0, len(api.ReqBody.Fields))
		for _, f := range api.ReqBody.Fields {
			assignments = append(assignments, f.Name+": params."+f.Name)
		}
		fmt.Fprintf(&b, "  options.body = JSON.stringify({%s});\n", strings.Join(assignments, ", "))
	}

	b.WriteString("  const resp = await fetch(url, options);\n")
	return b.String()
}

// buildJsResult renders the part of the function body that shapes the value returned to the caller.
func buildJsResult(api *schemaz.Api) string {
	var b strings.Builder

	b.WriteString("  const result = {status: resp.status};\n")
	if len(api.RespHeaders) > 0 {
		b.WriteString("  result.headers = {};\n")
		for _, f := range api.RespHeaders {
			fmt.Fprintf(&b, "  result.headers[%q] = resp.headers.get(%q);\n", f.Name, headerName(f.Name))
		}
	}
	if api.RespBody != nil {
		b.WriteString("  result.body = await resp.json();\n")
	}
	b.WriteString("  return result;\n")
	return b.String()
}

// buildJsFetch generates JSDoc-annotated JS source for an async function that
// calls the given schemaz.Api with the fetch API, so SDK consumers get autocomplete.
func buildJsFetch(api *schemaz.Api) (fnName string, code string) {
	fnName = api.Id
	code = buildJsDoc(api) +
		fmt.Sprintf("export async function %s(baseUrl, params) {\n", fnName) +
		buildJsRequest(api) +
		buildJsResult(api) +
		"}\n"
	return fnName, code
}

// jsCallExpr renders a JS call to fnName, e.g. `samplefraction(baseUrl, {"numerator":10,...})`.
// baseUrlExpr is inserted verbatim so callers can pass either a quoted literal or a bare identifier.
func jsCallExpr(fnName, baseUrlExpr string, params map[string]any) (string, error) {
	paramsJson, err := json.Marshal(params)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s(%s, %s)", fnName, baseUrlExpr, paramsJson), nil
}

// sampleFractionParams are the request values shared by the live round-trip test and the
// generated gen/example.js, so the example always reflects what's actually verified to work.
var sampleFractionParams = map[string]any{
	"numerator":   10,
	"denumerator": 3,
	"precision":   3,
	"trace_id":    "xx",
	"reason":      "myreason",
}

// genPackageJson marks gen/ as an ES module so both Node and the IDE resolve
// the generated import/export syntax; it has nothing to do with the schema, so it's hardcoded.
const genPackageJson = `{
  "type": "module",
  "private": true,
  "description": "Example SDK output for routezsamplefraction, generated from schemaz.Api for IDE autocomplete demos."
}
`

// genExampleJsTemplate wraps the generated call expression in a runnable sample, kept alongside
// the generated SDK so opening gen/ in an IDE demonstrates the JSDoc-driven autocomplete live.
const genExampleJsTemplate = `// @ts-check
import { %[1]s } from './samplefraction.js';

const baseUrl = 'http://localhost:8080';

const result = await %[2]s;

console.log(result.status);
console.log(result.headers.req_id);
console.log(result.body.display, result.body.result);
`

// writeJsSdkExamples writes gen/: the SDK code generated from the schema plus a matching usage
// example, so it can be opened in an IDE to see the generated JSDoc driving autocomplete.
func writeJsSdkExamples(t *testing.T, fnName, code string, params map[string]any) {
	t.Helper()
	genDir := "gen"
	require.NoError(t, filez.WriteFile(filepath.Join(genDir, "samplefraction.js"), []byte(code)))
	require.NoError(t, filez.WriteFile(filepath.Join(genDir, "package.json"), []byte(genPackageJson)))

	call, err := jsCallExpr(fnName, "baseUrl", params)
	require.NoError(t, err)
	example := fmt.Sprintf(genExampleJsTemplate, fnName, call)
	require.NoError(t, filez.WriteFile(filepath.Join(genDir, "example.js"), []byte(example)))
}

// runJsFetch executes generated JS code with Node against a live server and returns the fetched result.
func runJsFetch(t *testing.T, fnName, code, baseUrl string, params map[string]any) map[string]any {
	t.Helper()
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	call, err := jsCallExpr(fnName, fmt.Sprintf("%q", baseUrl), params)
	require.NoError(t, err)

	driver := fmt.Sprintf("%s\n%s.then(r => console.log(JSON.stringify(r))).catch(e => { console.error(e); process.exit(1); });\n",
		code, call)
	scriptPath := filepath.Join(t.TempDir(), "client.mjs")
	require.NoError(t, filez.WriteFile(scriptPath, []byte(driver)))

	out, err := exec.Command("node", scriptPath).CombinedOutput()
	require.NoError(t, err, string(out))

	var result map[string]any
	require.NoError(t, json.Unmarshal(bytes.TrimSpace(out), &result))
	return result
}

func TestManualJsClient(t *testing.T) {
	ctx := t.Context()
	s := httpzserver.New(ctx, httpzserver.Options{})
	defer s.Close()
	s.Listen()
	s.Start()

	routez.Register(s, routezsamplefraction.Api())

	fnName, code := buildJsFetch(routezsamplefraction.Schema())

	require.Contains(t, code, "@param {number} params.numerator - numerator")
	require.Contains(t, code, "@param {string} params.reason - reason")
	require.Contains(t, code, "@returns {Promise<{status: number, headers: {req_id: string}, body: {display: string, result: string}}>}")

	writeJsSdkExamples(t, fnName, code, sampleFractionParams)

	result := runJsFetch(t, fnName, code, s.Base().String(), sampleFractionParams)

	require.InDelta(t, 201, result["status"], 0)
	require.Equal(t, "reason: myreason, trace: xx", result["headers"].(map[string]any)["req_id"])
	body := result["body"].(map[string]any)
	require.Equal(t, "10.000/3.000", body["display"])
	require.Equal(t, "3.333", body["result"])
}
