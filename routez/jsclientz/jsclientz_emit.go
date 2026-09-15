package jsclientz

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/infinity6-ai/gox/routez/apiz"
	"github.com/infinity6-ai/gox/routez/internal/converter"
	"github.com/infinity6-ai/gox/schemaz/schemaz"
)

// jsIdentRe valida que spec.Id pode virar o nome de uma função JS exportada.
var jsIdentRe = regexp.MustCompile(`^[A-Za-z_$][A-Za-z0-9_$]*$`)

// pathParamRe encontra placeholders {name} no ApiSpec.Path.
var pathParamRe = regexp.MustCompile(`\{([A-Za-z0-9_]+)\}`)

// field junta um nome de wire com seu Node, já na ordem de emissão.
type field struct {
	name string
	node *Node
}

// BuildJsFetch gera o JSDoc + função fetch de uma Api, a partir do ApiSpec e dos 6 schemas de
// GetDataRefs. É o ponto de entrada do pacote para quem só quer o código JS pronto.
func BuildJsFetch(spec apiz.ApiSpec, refs *apiz.DataRefs) (fnName string, code string, err error) {
	if !jsIdentRe.MatchString(spec.Id) {
		return "", "", fmt.Errorf("jsclientz: %q não é um identificador JS válido", spec.Id)
	}
	fnName = spec.Id

	pathFields := introspectFields(refs.PathParams)
	queryFields := introspectFields(refs.QueryParams)
	headerFields := introspectFields(refs.ReqHeaders)
	bodyFields := introspectFields(refs.ReqBody)

	if err := validateFieldNames(pathFields, queryFields, headerFields, bodyFields); err != nil {
		return "", "", err
	}
	if err := validatePathParams(spec.Path, pathFields); err != nil {
		return "", "", err
	}

	var respHeaderNode, respBodyNode *Node
	if refs.RespHeaders != nil {
		respHeaderNode = Introspect(refs.RespHeaders)
	}
	if refs.RespBody != nil {
		respBodyNode = Introspect(refs.RespBody)
	}

	allReqFields := concatFields(pathFields, queryFields, headerFields, bodyFields)

	var b strings.Builder
	b.WriteString(buildJsDoc(spec, allReqFields, respHeaderNode, respBodyNode))
	fmt.Fprintf(&b, "export async function %s(baseUrl, params) {\n", fnName)
	b.WriteString(buildJsRequest(spec, queryFields, headerFields, bodyFields))
	b.WriteString(buildJsResult(respHeaderNode, respBodyNode))
	b.WriteString("}\n")
	return fnName, b.String(), nil
}

// introspectFields lê um schema de nível superior (sempre um Object, uma seção do DataRefs) e
// devolve seus campos já ordenados; nil vira lista vazia (a seção não existe para essa Api).
func introspectFields(schema *schemaz.Schema) []field {
	if schema == nil {
		return nil
	}
	n := Introspect(schema)
	fields := make([]field, 0, len(n.FieldOrder))
	for _, name := range n.FieldOrder {
		fields = append(fields, field{name: name, node: n.Fields[name]})
	}
	return fields
}

// concatFields junta vários grupos de campos numa lista só, na ordem dos grupos.
func concatFields(groups ...[]field) []field {
	var out []field
	for _, g := range groups {
		out = append(out, g...)
	}
	return out
}

// validateFieldNames garante que todo nome de campo vira uma propriedade JS válida via
// notação de ponto (params.nome) — só spec.Id era validado antes, um nome com "-" gerava
// JS sintaticamente errado (ou pior, silenciosamente errado: "params.a-b" é subtração).
func validateFieldNames(groups ...[]field) error {
	for _, g := range groups {
		for _, f := range g {
			if !jsIdentRe.MatchString(f.name) {
				return fmt.Errorf("jsclientz: campo %q não é um identificador JS válido", f.name)
			}
		}
	}
	return nil
}

// validatePathParams garante que os placeholders {name} de spec.Path batem exatamente com os
// campos do schema de PathParams — sem isso, um drift entre a rota Go e o schema (renomear um
// só dos dois) gera um cliente que compila mas manda "undefined" na URL, sem erro nenhum.
func validatePathParams(path string, pathFields []field) error {
	placeholders := map[string]bool{}
	for _, m := range pathParamRe.FindAllStringSubmatch(path, -1) {
		placeholders[m[1]] = true
	}
	declared := map[string]bool{}
	for _, f := range pathFields {
		declared[f.name] = true
	}
	for name := range placeholders {
		if !declared[name] {
			return fmt.Errorf("jsclientz: path %q referencia {%s}, mas o schema de PathParams não tem esse campo", path, name)
		}
	}
	for name := range declared {
		if !placeholders[name] {
			return fmt.Errorf("jsclientz: schema de PathParams tem o campo %q, mas o path %q não tem {%s}", name, path, name)
		}
	}
	return nil
}

// jsTypeLiteral renderiza o tipo JSDoc de um Node — "*" cobre tanto nó ausente quanto KindUnknown.
func jsTypeLiteral(n *Node) string {
	if n == nil {
		return "*"
	}
	switch n.Kind {
	case KindString:
		return "string"
	case KindNumber:
		return "number"
	case KindBoolean:
		return "boolean"
	case KindArray:
		return jsTypeLiteral(n.Element) + "[]"
	case KindObject:
		if len(n.FieldOrder) == 0 {
			return "Object"
		}
		parts := make([]string, 0, len(n.FieldOrder))
		for _, name := range n.FieldOrder {
			parts = append(parts, name+": "+jsTypeLiteral(n.Fields[name]))
		}
		return "{" + strings.Join(parts, ", ") + "}"
	default:
		return "*"
	}
}

// descSummary extrai o resumo de um Desc, nil-safe.
func descSummary(d *schemaz.Desc) string {
	if d == nil {
		return ""
	}
	return d.Summary
}

// buildJsDoc monta o comentário JSDoc: descrição da Api (se houver), um @param por campo de
// request e o @returns com o formato do resultado.
func buildJsDoc(spec apiz.ApiSpec, reqFields []field, respHeaderNode, respBodyNode *Node) string {
	var b strings.Builder
	b.WriteString("/**\n")
	if spec.Desc != nil {
		if d := spec.Desc(); d != nil && d.Summary != "" {
			fmt.Fprintf(&b, " * %s\n", d.Summary)
		}
	}
	b.WriteString(" * @param {string} baseUrl\n")
	b.WriteString(" * @param {Object} params\n")
	for _, f := range reqFields {
		typ := jsTypeLiteral(f.node)
		if s := descSummary(nodeDesc(f.node)); s != "" {
			fmt.Fprintf(&b, " * @param {%s} params.%s - %s\n", typ, f.name, s)
		} else {
			fmt.Fprintf(&b, " * @param {%s} params.%s\n", typ, f.name)
		}
	}
	fmt.Fprintf(&b, " * @returns {Promise<%s>}\n", respTypeLiteral(respHeaderNode, respBodyNode))
	b.WriteString(" */\n")
	return b.String()
}

// nodeDesc devolve n.Desc protegendo contra n nil.
func nodeDesc(n *Node) *schemaz.Desc {
	if n == nil {
		return nil
	}
	return n.Desc
}

// respTypeLiteral monta o tipo JSDoc do valor resolvido pela Promise: {status, headers?, body?}.
func respTypeLiteral(respHeaderNode, respBodyNode *Node) string {
	parts := []string{"status: number"}
	if respHeaderNode != nil && len(respHeaderNode.FieldOrder) > 0 {
		headerParts := make([]string, 0, len(respHeaderNode.FieldOrder))
		for _, name := range respHeaderNode.FieldOrder {
			headerParts = append(headerParts, name+": "+jsTypeLiteral(respHeaderNode.Fields[name]))
		}
		parts = append(parts, "headers: {"+strings.Join(headerParts, ", ")+"}")
	}
	if respBodyNode != nil {
		parts = append(parts, "body: "+jsTypeLiteral(respBodyNode))
	}
	return "{" + strings.Join(parts, ", ") + "}"
}

// buildJsRequest monta a URL, query, headers e corpo da requisição, e dispara o fetch.
// Os path params são interpolados direto em spec.Path, não precisam de uma lista à parte.
func buildJsRequest(spec apiz.ApiSpec, queryFields, headerFields, bodyFields []field) string {
	var b strings.Builder

	jsPath := pathParamRe.ReplaceAllString(spec.Path, "${params.$1}")
	fmt.Fprintf(&b, "  const url = new URL(`${baseUrl}%s`);\n", jsPath)

	for _, f := range queryFields {
		writeParamAppend(&b, "url.searchParams", f)
	}

	b.WriteString("  const headers = {'Content-Type': 'application/json'};\n")
	for _, f := range headerFields {
		wireName := converter.Json2HeaderName(f.name)
		if f.node != nil && f.node.Kind == KindArray {
			fmt.Fprintf(&b, "  headers[%q] = params.%s.map(String).join(\", \");\n", wireName, f.name)
		} else {
			fmt.Fprintf(&b, "  headers[%q] = String(params.%s);\n", wireName, f.name)
		}
	}

	fmt.Fprintf(&b, "  const options = {method: %q, headers};\n", spec.Method)
	if len(bodyFields) > 0 {
		assignments := make([]string, 0, len(bodyFields))
		for _, f := range bodyFields {
			assignments = append(assignments, f.name+": "+bodyValueExpr(f))
		}
		fmt.Fprintf(&b, "  options.body = JSON.stringify({%s});\n", strings.Join(assignments, ", "))
	}

	b.WriteString("  const resp = await fetch(url, options);\n")
	return b.String()
}

// writeParamAppend emite .set para valor único, ou um loop .append por valor quando o campo é
// realmente uma lista (Kind==KindArray) — não usa MultiValue direto: MultiValue só diz que o
// campo veio de Strs, mas um Raw apontando pra []string também deve virar múltiplos .append,
// e um Strs com nativo escalar (caso "c" em schemaz_test.go) deve continuar sendo um valor só.
func writeParamAppend(b *strings.Builder, target string, f field) {
	if f.node != nil && f.node.Kind == KindArray {
		fmt.Fprintf(b, "  for (const v of params.%s) { %s.append(%q, String(v)); }\n", f.name, target, f.name)
	} else {
		fmt.Fprintf(b, "  %s.set(%q, String(params.%s));\n", target, f.name, f.name)
	}
}

// bodyValueExpr monta a expressão JS de um campo do corpo. Campos Str/Strs (StringArrayWire)
// viajam como array JSON de strings; campos Raw viajam com o valor nativo, sem embrulho. A
// forma do lado JS (array ou valor único) segue Kind==KindArray, não MultiValue (mesma razão
// de writeParamAppend).
func bodyValueExpr(f field) string {
	if f.node == nil || !f.node.StringArrayWire {
		return "params." + f.name
	}
	if f.node.Kind == KindArray {
		return "params." + f.name + ".map(String)"
	}
	return "[String(params." + f.name + ")]"
}

// buildJsResult monta o objeto {status, headers, body} devolvido pela função gerada.
func buildJsResult(respHeaderNode, respBodyNode *Node) string {
	var b strings.Builder
	b.WriteString("  const result = {status: resp.status};\n")
	if respHeaderNode != nil && len(respHeaderNode.FieldOrder) > 0 {
		b.WriteString("  result.headers = {};\n")
		for _, name := range respHeaderNode.FieldOrder {
			wireName := converter.Json2HeaderName(name)
			fmt.Fprintf(&b, "  result.headers[%q] = resp.headers.get(%q);\n", name, wireName)
		}
	}
	if respBodyNode != nil {
		b.WriteString("  result.body = await resp.json();\n")
	}
	b.WriteString("  return result;\n")
	return b.String()
}
