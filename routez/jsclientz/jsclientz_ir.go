package jsclientz

import "github.com/infinity6-ai/gox/schemaz/schemaz"

// Kind é o tipo JS/JSDoc inferido para um nó do schema.
type Kind string

const (
	KindString  Kind = "string"
	KindNumber  Kind = "number"
	KindBoolean Kind = "boolean"
	KindObject  Kind = "object"
	KindArray   Kind = "array"
	// KindUnknown não é erro: o campo é tratado como JSON opaco, repassado como está.
	KindUnknown Kind = "unknown"
)

// Node é a representação intermediária (IR) de um nó do schemaz.Schema, produzida pelo walker
// e consumida pelo emissor de JS.
type Node struct {
	Kind Kind
	// Desc é opcional (a maioria dos schemas hoje não a preenche).
	Desc *schemaz.Desc
	// Nullable indica que o nó pode ser nil (Object/Array que retornam nil, ou ponteiro Go).
	Nullable bool
	// MultiValue é fato de transporte (Strs), independente do Kind (ver caso "c" em schemaz_test.go).
	MultiValue bool
	// StringArrayWire é true para nós vindos de Str ou Strs: ambos serializam como array JSON
	// de strings (Str = array de 1 elemento). Só importa para montar/ler ReqBody manualmente;
	// QueryParams/ReqHeaders usam URLSearchParams/Headers, que já abstraem isso.
	StringArrayWire bool
	// DynamicKeys indica Raw -> map[K]V (chaves não fixas; hoje não ocorre em uso real).
	DynamicKeys bool
	// Fields são os campos de um objeto, chaveados pelo nome de wire (chave usada em Object()).
	Fields map[string]*Node
	// FieldOrder é Fields ordenado (map não garante ordem estável entre execuções).
	FieldOrder []string
	// Element é o tipo do elemento de um array, ou o tipo de valor quando DynamicKeys.
	Element *Node
}
