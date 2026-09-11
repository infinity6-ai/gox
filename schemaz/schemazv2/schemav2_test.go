package schemazv2_test

import (
	"strconv"
	"testing"

	"github.com/infinity6-ai/gox/commonz/jsonz"
	"github.com/infinity6-ai/gox/commonz/strconvz"
	"github.com/infinity6-ai/gox/schemaz/schemazv2"
	"github.com/stretchr/testify/require"
)

func TestUnitBasic(t *testing.T) {

	type Address struct {
		Street string `json:"street"`
		City   string `json:"city"`
	}

	type Person struct {
		Name                  string     `json:"name"`
		Age                   int        `json:"age"`
		MainAddress           *Address   `json:"main_address"`
		MainAddressNil        *Address   `json:"main_address_nil"`
		SecondaryAddresses    *Address   `json:"secondary_addresses"`
		SecondaryAddressesNil *Address   `json:"secondary_addresses_nil"`
		Addresses             []*Address `json:"addresses"`
		CompanyAddresses      []*Address `json:"company_addresses"`
		Numbers               []int      `json:"numbers"`
	}

	expected := Person{
		Name: "John",
		Age:  30,
		MainAddress: &Address{
			Street: "123 Main St",
			City:   "New York",
		},
		SecondaryAddresses: &Address{
			Street: "456 Elm St",
			City:   "Los Angeles",
		},
		Addresses: []*Address{
			{
				Street: "789 Oak St",
				City:   "Chicago",
			},
			{
				Street: "321 Pine St",
				City:   "Houston",
			},
		},
		CompanyAddresses: []*Address{
			{
				Street: "123 Main St",
				City:   "New York",
			},
			nil,
			{
				Street: "456 Elm St",
				City:   "Los Angeles",
			},
			nil,
		},
		Numbers: []int{10, 20, 30, 40, 50},
	}

	var person Person

	mapper := &schemazv2.Schema{
		Object: func(read bool) map[string]*schemazv2.Schema {
			return map[string]*schemazv2.Schema{
				"name": {
					Raw: func() any {
						return &person.Name
					},
				},
				"age": {
					Raw: func() any {
						return &person.Age
					},
				},
				"main_address": {
					Raw: func() any {
						return &person.MainAddress
					},
				},
				"main_address_nil": {
					Raw: func() any {
						return &person.MainAddressNil
					},
				},
				"secondary_addresses": {
					Object: func(read bool) map[string]*schemazv2.Schema {
						if !read {
							person.SecondaryAddresses = &Address{}
						}
						if read && person.SecondaryAddresses == nil {
							return nil
						}
						return map[string]*schemazv2.Schema{
							"street": {
								Raw: func() any { return &person.SecondaryAddresses.Street },
							},
							"city": {
								Raw: func() any { return &person.SecondaryAddresses.City },
							},
						}
					},
				},
				"secondary_addresses_nil": {
					Object: func(read bool) map[string]*schemazv2.Schema {
						if !read {
							person.SecondaryAddressesNil = &Address{}
						}
						if read && person.SecondaryAddressesNil == nil {
							return nil
						}
						return map[string]*schemazv2.Schema{
							"street": {
								Raw: func() any { return &person.SecondaryAddressesNil.Street },
							},
							"city": {
								Raw: func() any { return &person.SecondaryAddressesNil.City },
							},
						}
					},
				},
				"addresses": {
					Raw: func() any {
						return &person.Addresses
					},
				},
				"company_addresses": {
					Array: func() (length int, getElement func(idx int, read bool) *schemazv2.Schema) {
						return len(person.CompanyAddresses), func(idx int, read bool) *schemazv2.Schema {
							if idx >= len(person.CompanyAddresses) {
								var a *Address
								person.CompanyAddresses = append(person.CompanyAddresses, a)
							}
							return &schemazv2.Schema{
								Object: func(read bool) map[string]*schemazv2.Schema {
									if !read {
										person.CompanyAddresses[idx] = &Address{}
									}
									if read && person.CompanyAddresses[idx] == nil {
										return nil
									}
									return map[string]*schemazv2.Schema{
										"street": {
											Raw: func() any { return &person.CompanyAddresses[idx].Street },
										},
										"city": {
											Raw: func() any { return &person.CompanyAddresses[idx].City },
										},
									}
								},
							}
						}
					},
				},
				"numbers": {
					Array: func() (length int, getElement func(idx int, read bool) *schemazv2.Schema) {
						return len(person.Numbers), func(idx int, read bool) *schemazv2.Schema {
							if idx >= len(person.Numbers) {
								person.Numbers = append(person.Numbers, 0)
							}
							return &schemazv2.Schema{
								Raw: func() any {
									return &person.Numbers[idx]
								},
							}
						}
					},
				},
			}
		},
	}

	str := jsonz.MustFormat(expected)
	jsonz.MustParse(str.Bytes(), mapper)

	require.Equal(t, expected, person)

	mapperStr := jsonz.MustFormat(mapper)
	require.Equal(t, expected, *jsonz.MustParse(mapperStr.Bytes(), &Person{}))
}

func TestUnitValues(t *testing.T) {
	str := `{"a":["10.1"],"b":["20.1","20.2"],"c":["30.1","30.2"]}`
	type My struct {
		A float64
		B []float64
		C float64
	}
	var my My
	mapper := &schemazv2.Schema{
		Object: func(read bool) map[string]*schemazv2.Schema {
			return map[string]*schemazv2.Schema{
				"a": {
					Str: func() (func(v string), func() string) {
						return func(unformatted string) {
								my.A = strconvz.MustParseNumber[float64](unformatted)
							}, func() string {
								return strconv.FormatFloat(my.A, 'f', -1, 64)
							}
					},
				},
				"b": {
					Strs: func() (func(v []string), func() []string) {
						return func(unformatted []string) {
								my.B = make([]float64, len(unformatted))
								for i, v := range unformatted {
									my.B[i] = strconvz.MustParseNumber[float64](v)
								}
							}, func() []string {
								out := make([]string, len(my.B))
								for i, v := range my.B {
									out[i] = strconv.FormatFloat(v, 'f', -1, 64)
								}
								return out
							}
					},
				},
				"c": {
					Strs: func() (func(v []string), func() []string) {
						return func(unformatted []string) {
								my.C = strconvz.MustParseNumber[float64](unformatted[0])
							}, func() []string {
								return []string{strconv.FormatFloat(my.C, 'f', -1, 64)}
							}
					},
				},
			}
		},
	}

	expected := My{
		A: 10.1,
		B: []float64{20.1, 20.2},
		C: 30.1,
	}

	jsonz.MustParse([]byte(str), mapper)
	require.Equal(t, expected, my)

	mapperStr := jsonz.MustFormat(mapper).String()
	parsedMapperStr := jsonz.MustParse(mapperStr, new(map[string][]string{}))
	require.Equal(t, map[string][]string{
		"a": {"10.1"},
		"b": {"20.1", "20.2"},
		"c": {"30.1"},
	}, *parsedMapperStr)
}
