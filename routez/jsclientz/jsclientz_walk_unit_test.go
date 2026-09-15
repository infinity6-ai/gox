package jsclientz_test

import (
	"testing"

	"github.com/infinity6-ai/gox/commonz/slicez"
	"github.com/infinity6-ai/gox/routez/jsclientz"
	"github.com/infinity6-ai/gox/schemaz/schemaz"
	"github.com/stretchr/testify/require"
)

// Estes testes recriam, de forma independente, os casos de schemaz_test.go (nested object,
// array de objetos com nil, array de primitivos, Strs escalar) — já que aquelas fixtures são
// locais à função e não dá pra importar. FractionApi não usa nenhum desses formatos hoje.

func TestUnitIntrospectNestedObject(t *testing.T) {
	type Address struct {
		Street string `json:"street"`
	}
	type Person struct {
		MainAddress *Address
	}
	var person Person

	s := &schemaz.Schema{
		Object: func(read bool) map[string]*schemaz.Schema {
			return map[string]*schemaz.Schema{
				"main_address": {
					Object: func(read bool) map[string]*schemaz.Schema {
						if !read {
							person.MainAddress = &Address{}
						}
						if read && person.MainAddress == nil {
							return nil
						}
						return map[string]*schemaz.Schema{
							"street": {Raw: func() any { return &person.MainAddress.Street }},
						}
					},
				},
			}
		},
	}

	n := jsclientz.Introspect(s)
	require.Equal(t, jsclientz.KindObject, n.Kind)
	main := n.Fields["main_address"]
	require.NotNil(t, main)
	require.Equal(t, jsclientz.KindObject, main.Kind)
	street := main.Fields["street"]
	require.NotNil(t, street)
	require.Equal(t, jsclientz.KindString, street.Kind)
}

func TestUnitIntrospectArrayOfObjectsWithNilElements(t *testing.T) {
	type Address struct {
		Street string `json:"street"`
	}
	var addrs []*Address

	s := &schemaz.Schema{
		Object: func(read bool) map[string]*schemaz.Schema {
			return map[string]*schemaz.Schema{
				"company_addresses": {
					Array: func() *schemaz.Array {
						return &schemaz.Array{
							Len: len(addrs),
							Get: func(idx int, read bool) *schemaz.Schema {
								addrs = slicez.GrowLenTo(addrs, idx+1)
								return &schemaz.Schema{
									Object: func(read bool) map[string]*schemaz.Schema {
										if !read {
											addrs[idx] = &Address{}
										}
										if read && addrs[idx] == nil {
											return nil
										}
										return map[string]*schemaz.Schema{
											"street": {Raw: func() any { return &addrs[idx].Street }},
										}
									},
								}
							},
						}
					},
				},
			}
		},
	}

	n := jsclientz.Introspect(s)
	arr := n.Fields["company_addresses"]
	require.NotNil(t, arr)
	require.Equal(t, jsclientz.KindArray, arr.Kind)
	require.NotNil(t, arr.Element)
	require.Equal(t, jsclientz.KindObject, arr.Element.Kind)
	require.Contains(t, arr.Element.FieldOrder, "street")
}

func TestUnitIntrospectArrayOfPrimitives(t *testing.T) {
	var numbers []int

	s := &schemaz.Schema{
		Object: func(read bool) map[string]*schemaz.Schema {
			return map[string]*schemaz.Schema{
				"numbers": {
					Array: func() *schemaz.Array {
						return &schemaz.Array{
							Len: len(numbers),
							Get: func(idx int, read bool) *schemaz.Schema {
								numbers = slicez.GrowLenTo(numbers, idx+1)
								return &schemaz.Schema{Raw: func() any { return &numbers[idx] }}
							},
						}
					},
				},
			}
		},
	}

	n := jsclientz.Introspect(s)
	arr := n.Fields["numbers"]
	require.Equal(t, jsclientz.KindArray, arr.Kind)
	require.Equal(t, jsclientz.KindNumber, arr.Element.Kind)
}

func TestUnitIntrospectStrsScalarIsNotArray(t *testing.T) {
	var c float64

	s := &schemaz.Schema{
		Object: func(read bool) map[string]*schemaz.Schema {
			return map[string]*schemaz.Schema{
				"c": {
					Strs: func() *schemaz.Parser[[]string] {
						return &schemaz.Parser[[]string]{
							Parse: func(v []string) {},
							Format: func() ([]string, any) {
								return []string{"30.1"}, c
							},
						}
					},
				},
			}
		},
	}

	n := jsclientz.Introspect(s)
	cNode := n.Fields["c"]
	require.NotNil(t, cNode)
	require.True(t, cNode.MultiValue)
	require.True(t, cNode.StringArrayWire)
	require.Equal(t, jsclientz.KindNumber, cNode.Kind)
}

func TestUnitIntrospectStrsSliceIsArray(t *testing.T) {
	var b []float64

	s := &schemaz.Schema{
		Object: func(read bool) map[string]*schemaz.Schema {
			return map[string]*schemaz.Schema{
				"b": {
					Strs: func() *schemaz.Parser[[]string] {
						return &schemaz.Parser[[]string]{
							Parse: func(v []string) {},
							Format: func() ([]string, any) {
								return []string{}, b
							},
						}
					},
				},
			}
		},
	}

	n := jsclientz.Introspect(s)
	bNode := n.Fields["b"]
	require.True(t, bNode.MultiValue)
	require.Equal(t, jsclientz.KindArray, bNode.Kind)
	require.Equal(t, jsclientz.KindNumber, bNode.Element.Kind)
}

func TestUnitIntrospectRawNilIsUnknownNotPanic(t *testing.T) {
	s := &schemaz.Schema{
		Object: func(read bool) map[string]*schemaz.Schema {
			return map[string]*schemaz.Schema{
				"mystery": {Raw: func() any { return nil }},
			}
		},
	}

	n := jsclientz.Introspect(s)
	require.Equal(t, jsclientz.KindUnknown, n.Fields["mystery"].Kind)
}

func TestUnitIntrospectFieldOrderIsSorted(t *testing.T) {
	s := &schemaz.Schema{
		Object: func(read bool) map[string]*schemaz.Schema {
			var z, a, m int
			return map[string]*schemaz.Schema{
				"zebra": {Raw: func() any { return &z }},
				"apple": {Raw: func() any { return &a }},
				"mango": {Raw: func() any { return &m }},
			}
		},
	}

	n := jsclientz.Introspect(s)
	require.Equal(t, []string{"apple", "mango", "zebra"}, n.FieldOrder)
}
