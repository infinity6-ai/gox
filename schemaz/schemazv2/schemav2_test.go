package schemazv2_test

import (
	"testing"

	"github.com/infinity6-ai/gox/commonz/jsonz"
	"github.com/infinity6-ai/gox/commonz/strconvz"
	"github.com/infinity6-ai/gox/schemaz/mapperz"
	"github.com/infinity6-ai/gox/schemaz/schemazv2"
	"github.com/stretchr/testify/require"
)

func TestUnitBasic(t *testing.T) {

	type Address struct {
		Street string `json:"street"`
		City   string `json:"city"`
	}

	type Person struct {
		Name               string     `json:"name"`
		Age                int        `json:"age"`
		MainAddress        *Address   `json:"main_address"`
		SecondaryAddresses *Address   `json:"secondary_addresses"`
		Addresses          []*Address `json:"addresses"`
		CompanyAddresses   []*Address `json:"company_addresses"`
		Numbers            []int      `json:"numbers"`
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
			{
				Street: "456 Elm St",
				City:   "Los Angeles",
			},
		},
		Numbers: []int{10, 20, 30, 40, 50},
	}

	var person Person

	mapper := &schemazv2.Schema{
		Object: func() map[string]*schemazv2.Schema {
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
				"secondary_addresses": {
					Object: func() map[string]*schemazv2.Schema {
						person.SecondaryAddresses = &Address{}
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
				"addresses": {
					Raw: func() any {
						return &person.Addresses
					},
				},
				"company_addresses": {
					Array: func() *schemazv2.Schema {
						return &schemazv2.Schema{
							Object: func() map[string]*schemazv2.Schema {
								a := &Address{}
								person.CompanyAddresses = append(person.CompanyAddresses, a)
								return map[string]*schemazv2.Schema{
									"street": {
										Raw: func() any { return &a.Street },
									},
									"city": {
										Raw: func() any { return &a.City },
									},
								}
							},
						}
					},
				},
				"numbers": {
					Array: func() *schemazv2.Schema {
						return &schemazv2.Schema{
							Raw: func() any {
								person.Numbers = append(person.Numbers, 0)
								return &mapperz.Mapper{
									Target: &person.Numbers[len(person.Numbers)-1],
								}
							},
						}
					},
				},
			}
		},
	}

	str := jsonz.MustFormat(expected)
	jsonz.MustParse(str.Bytes(), mapper)

	require.Equal(t, expected, person)
}

func TestUnitValues(t *testing.T) {
	str := `{"a":"10.1"}`
	type My struct {
		A float64
	}
	var my My
	mapper := &schemazv2.Schema{
		Object: func() map[string]*schemazv2.Schema {
			return map[string]*schemazv2.Schema{
				"a": {
					Values: func(unformatted []string) {
						var x string
						jsonz.MustParse(unformatted[0], &x)
						my.A = strconvz.MustParseNumber[float64](x)
					},
				},
			}
		},
	}
	jsonz.MustParse([]byte(str), mapper)
}
