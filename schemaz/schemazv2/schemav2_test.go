package schemazv2_test

import (
	"testing"

	"github.com/infinity6-ai/gox/commonz/jsonz"
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
		// Addresses: []*Address{
		// 	{
		// 		Street: "789 Oak St",
		// 		City:   "Chicago",
		// 	},
		// 	{
		// 		Street: "321 Pine St",
		// 		City:   "Houston",
		// 	},
		// },
		// CompanyAddresses: []*Address{
		// 	{
		// 		Street: "123 Main St",
		// 		City:   "New York",
		// 	},
		// 	{
		// 		Street: "456 Elm St",
		// 		City:   "Los Angeles",
		// 	},
		// },
		// Numbers: []int{10, 20, 30, 40, 50},
	}

	var person Person
	person.SecondaryAddresses = &Address{}

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
				// "addresses": {
				// 	Raw: func() any {
				// 		return &person.Addresses
				// 	},
				// },
				// "company_addresses": {
				// 	Raw: func() any {
				// 		return &person.CompanyAddresses
				// 	},
				// },
				// "numbers": {
				// 	Raw: func() any {
				// 		return &person.Numbers
				// 	},
				// },
			}
		},
		// Target: map[string]*schemazv2.Schema{
		// 	"name": {
		// 		Target: &person.Name,
		// 	},
		// 	"age": {
		// 		Target: &person.Age,
		// 	},
		// 	"main_address": {
		// 		Target: &person.MainAddress,
		// 	},
		// 	"secondary_addresses": {
		// 		Target: &schemazv2.Schema{
		// 			Target: map[string]*schemazv2.Schema{
		// 				"street": {
		// 					Target: &person.SecondaryAddresses.Street,
		// 				},
		// 				"city": {
		// 					Target: &person.SecondaryAddresses.City,
		// 				},
		// 			},
		// 		},
		// 	},
		// 	"addresses": {
		// 		Target: &person.Addresses,
		// 	},
		// 	"company_addresses": {
		// 		Target: &schemazv2.Schema{
		// 			Target: &mapperz.Array{
		// 				Element: func() *schemazv2.Schema {
		// 					y := &Address{}
		// 					person.CompanyAddresses = append(person.CompanyAddresses, y)
		// 					return &schemazv2.Schema{
		// 						Target: map[string]*schemazv2.Schema{
		// 							"street": {
		// 								Target: &y.Street,
		// 							},
		// 							"city": {
		// 								Target: &y.City,
		// 							},
		// 						},
		// 					}
		// 				},
		// 			},
		// 		},
		// 	},
		// 	"numbers": {
		// 		Target: &schemazv2.Schema{
		// 			Target: &mapperz.Array{
		// 				Element: func() *schemazv2.Schema {
		// 					person.Numbers = append(person.Numbers, 0)
		// 					return &schemazv2.Schema{
		// 						Target: &person.Numbers[len(person.Numbers)-1],
		// 					}
		// 				},
		// 			},
		// 		},
		// 	},
		// },
	}

	str := jsonz.MustFormat(expected)
	jsonz.MustParse(str.Bytes(), mapper)

	require.Equal(t, expected, person)
}
