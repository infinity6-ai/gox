package jsclientz

import (
	"encoding"
	"encoding/json"
	"reflect"
	"sort"
	"strings"

	"github.com/infinity6-ai/gox/schemaz/schemaz"
)

// maxDepth evita recursão infinita ao percorrer schemas/tipos muito profundos ou cíclicos.
const maxDepth = 32

// jsonMarshalerType/textMarshalerType detectam tipos com serialização JSON customizada
// (time.Time, uuid.UUID, etc.) antes de tentar decompor os campos Go do struct, que dariam
// um resultado errado (ex.: time.Time só tem campos não exportados — viraria "{}" vazio,
// quando na real ele serializa como uma string).
var (
	jsonMarshalerType = reflect.TypeFor[json.Marshaler]()
	textMarshalerType = reflect.TypeFor[encoding.TextMarshaler]()
)

// implementsMarshaler checa t e *t (o método pode estar em qualquer um dos dois receivers).
func implementsMarshaler(t reflect.Type, iface reflect.Type) bool {
	return t.Implements(iface) || reflect.PointerTo(t).Implements(iface)
}

// Introspect percorre um *schemaz.Schema (de uma instância zero-valor de uma Api) e monta a IR.
func Introspect(schema *schemaz.Schema) *Node {
	return introspectSchema(schema, 0)
}

// introspectSchema decide qual dos 5 bindings (Object/Array/Raw/Str/Strs) está setado e delega.
func introspectSchema(s *schemaz.Schema, depth int) *Node {
	if s == nil {
		return &Node{Kind: KindUnknown, Nullable: true}
	}
	var n *Node
	if depth > maxDepth {
		n = &Node{Kind: KindUnknown}
	} else {
		switch {
		case s.Object != nil:
			n = introspectObject(s.Object, depth)
		case s.Array != nil:
			n = introspectArray(s.Array, depth)
		case s.Raw != nil:
			n = introspectRaw(s.Raw, depth)
		case s.Str != nil:
			n = introspectStr(s.Str)
		case s.Strs != nil:
			n = introspectStrs(s.Strs)
		default:
			n = &Node{Kind: KindUnknown}
		}
	}
	n.Desc = s.Desc
	return n
}

// introspectObject chama Object(false) (direção unmarshal, aloca ponteiros aninhados) e ordena
// as chaves, já que a ordem de um map em Go não é estável entre execuções.
func introspectObject(fn func(read bool) map[string]*schemaz.Schema, depth int) *Node {
	m := safeObject(fn)
	if m == nil {
		return &Node{Kind: KindObject, Nullable: true, Fields: map[string]*Node{}}
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	fields := make(map[string]*Node, len(m))
	for _, k := range keys {
		fields[k] = introspectSchema(m[k], depth+1)
	}
	return &Node{Kind: KindObject, Fields: fields, FieldOrder: keys}
}

// introspectArray usa Get(0, false) como sonda de forma — o mesmo padrão que os autores de
// schema usam (via slicez.GrowLenTo) para crescer o slice sob demanda no caminho de unmarshal.
func introspectArray(fn func() *schemaz.Array, depth int) *Node {
	arr := safeArray(fn)
	if arr == nil {
		return &Node{Kind: KindArray, Nullable: true, Element: &Node{Kind: KindUnknown}}
	}
	elemSchema := safeArrayGet(arr.Get, 0)
	if elemSchema == nil {
		return &Node{Kind: KindArray, Element: &Node{Kind: KindUnknown}}
	}
	return &Node{Kind: KindArray, Element: introspectSchema(elemSchema, depth+1)}
}

// introspectRaw obtém o ponteiro via Raw() e infere o tipo real pelo reflect.Type do ponteiro.
func introspectRaw(fn func() any, depth int) *Node {
	v := safeRaw(fn)
	if v == nil {
		return &Node{Kind: KindUnknown, Nullable: true}
	}
	return introspectType(reflect.TypeOf(v), map[reflect.Type]bool{}, depth)
}

// introspectStr usa o segundo retorno de Format() (valor nativo, antes de virar string) para
// descobrir o tipo real por trás de um Parser[string] — único canal de tipo disponível aqui.
func introspectStr(fn func() *schemaz.Parser[string]) *Node {
	native := safeStrFormat(fn)
	if native == nil {
		return &Node{Kind: KindUnknown, StringArrayWire: true}
	}
	n := introspectType(reflect.TypeOf(native), map[reflect.Type]bool{}, 0)
	n.StringArrayWire = true
	return n
}

// introspectStrs é igual a introspectStr, mas marca MultiValue — fato de transporte que não
// deve ser confundido com "é array": o campo "c" em schemaz_test.go é Strs com nativo escalar.
func introspectStrs(fn func() *schemaz.Parser[[]string]) *Node {
	native := safeStrsFormat(fn)
	if native == nil {
		return &Node{Kind: KindUnknown, MultiValue: true, StringArrayWire: true}
	}
	base := introspectType(reflect.TypeOf(native), map[reflect.Type]bool{}, 0)
	base.MultiValue = true
	base.StringArrayWire = true
	return base
}

// introspectType mapeia um reflect.Type Go para o Kind JS/JSDoc correspondente, recursivamente.
func introspectType(t reflect.Type, visited map[reflect.Type]bool, depth int) *Node {
	if t == nil {
		return &Node{Kind: KindUnknown}
	}
	if depth > maxDepth || visited[t] {
		return &Node{Kind: KindUnknown}
	}
	switch t.Kind() {
	case reflect.String:
		return &Node{Kind: KindString}
	case reflect.Bool:
		return &Node{Kind: KindBoolean}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return &Node{Kind: KindNumber}
	case reflect.Ptr:
		n := introspectType(t.Elem(), extendVisited(visited, t), depth+1)
		n.Nullable = true
		return n
	case reflect.Slice, reflect.Array:
		return &Node{Kind: KindArray, Element: introspectType(t.Elem(), extendVisited(visited, t), depth+1)}
	case reflect.Map:
		return &Node{Kind: KindObject, DynamicKeys: true, Element: introspectType(t.Elem(), extendVisited(visited, t), depth+1)}
	case reflect.Struct:
		// time.Time, uuid.UUID etc. têm serialização JSON própria — decompor os campos Go
		// daria errado (time.Time só tem campo não exportado, viraria "{}" vazio).
		if implementsMarshaler(t, jsonMarshalerType) {
			return &Node{Kind: KindUnknown} // MarshalJSON pode devolver qualquer forma, não dá pra saber qual
		}
		if implementsMarshaler(t, textMarshalerType) {
			return &Node{Kind: KindString} // encoding/json sempre embrulha TextMarshaler num JSON string
		}
		return introspectStructType(t, extendVisited(visited, t), depth+1)
	default:
		// Interface/any/chan/func/etc. — tratado como JSON opaco, não é erro.
		return &Node{Kind: KindUnknown}
	}
}

// introspectStructType decompõe um struct Go usando as tags `json:"..."` como nome de wire.
// Só é usado quando um Raw aponta direto para um struct em vez de passar por Object aninhado
// (não ocorre em FractionApi hoje, mas nada impede um autor futuro de fazer isso).
func introspectStructType(t reflect.Type, visited map[reflect.Type]bool, depth int) *Node {
	fields := map[string]*Node{}
	var order []string
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.PkgPath != "" {
			continue // campo não exportado
		}
		name, skip := jsonFieldName(f)
		if skip {
			continue
		}
		fields[name] = introspectType(f.Type, visited, depth+1)
		order = append(order, name)
	}
	sort.Strings(order)
	return &Node{Kind: KindObject, Fields: fields, FieldOrder: order}
}

// jsonFieldName lê a tag json de um campo de struct, com fallback pro nome do campo Go.
func jsonFieldName(f reflect.StructField) (name string, skip bool) {
	tag := f.Tag.Get("json")
	if tag == "-" {
		return "", true
	}
	if tag == "" {
		return f.Name, false
	}
	parts := strings.SplitN(tag, ",", 2)
	if parts[0] == "" {
		return f.Name, false
	}
	return parts[0], false
}

// extendVisited copia o set de tipos visitados no caminho atual (não é global — tipos irmãos
// repetidos, como dois campos *Address, continuam válidos; só ciclos reais são bloqueados).
func extendVisited(visited map[reflect.Type]bool, t reflect.Type) map[reflect.Type]bool {
	n := make(map[reflect.Type]bool, len(visited)+1)
	for k, v := range visited {
		n[k] = v
	}
	n[t] = true
	return n
}

// safeObject chama Object(false) protegendo contra panics de closures escritas à mão por
// autores de API — degrada para "não sei" em vez de derrubar a geração inteira.
func safeObject(fn func(read bool) map[string]*schemaz.Schema) (m map[string]*schemaz.Schema) {
	defer func() {
		if recover() != nil {
			m = nil
		}
	}()
	return fn(false)
}

// safeArray é o equivalente de safeObject para Array().
func safeArray(fn func() *schemaz.Array) (a *schemaz.Array) {
	defer func() {
		if recover() != nil {
			a = nil
		}
	}()
	return fn()
}

// safeArrayGet é o equivalente de safeObject para Array.Get(idx, false).
func safeArrayGet(fn func(idx int, read bool) *schemaz.Schema, idx int) (s *schemaz.Schema) {
	defer func() {
		if recover() != nil {
			s = nil
		}
	}()
	return fn(idx, false)
}

// safeRaw é o equivalente de safeObject para Raw().
func safeRaw(fn func() any) (v any) {
	defer func() {
		if recover() != nil {
			v = nil
		}
	}()
	return fn()
}

// safeStrFormat chama Str().Format() e devolve só o valor nativo (segundo retorno).
func safeStrFormat(fn func() *schemaz.Parser[string]) (native any) {
	defer func() {
		if recover() != nil {
			native = nil
		}
	}()
	p := fn()
	if p == nil || p.Format == nil {
		return nil
	}
	_, native = p.Format()
	return native
}

// safeStrsFormat é o equivalente de safeStrFormat para Strs().Format().
func safeStrsFormat(fn func() *schemaz.Parser[[]string]) (native any) {
	defer func() {
		if recover() != nil {
			native = nil
		}
	}()
	p := fn()
	if p == nil || p.Format == nil {
		return nil
	}
	_, native = p.Format()
	return native
}
