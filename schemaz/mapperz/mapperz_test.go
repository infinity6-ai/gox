package mapperz_test

import (
	"testing"

	"github.com/infinity6-ai/gox/commonz/jsonz"
	"github.com/infinity6-ai/gox/schemaz/mapperz"
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
	person.SecondaryAddresses = &Address{}

	mapper := &mapperz.Mapper{
		Target: map[string]*mapperz.Mapper{
			"name": {
				Target: &person.Name,
			},
			"age": {
				Target: &person.Age,
			},
			"main_address": {
				Target: &person.MainAddress,
			},
			"secondary_addresses": {
				Target: &mapperz.Mapper{
					Target: map[string]*mapperz.Mapper{
						"street": {
							Target: &person.SecondaryAddresses.Street,
						},
						"city": {
							Target: &person.SecondaryAddresses.City,
						},
					},
				},
			},
			"addresses": {
				Target: &person.Addresses,
			},
			"company_addresses": {
				Target: &mapperz.Mapper{
					Target: &mapperz.Array{
						Element: func() *mapperz.Mapper {
							y := &Address{}
							person.CompanyAddresses = append(person.CompanyAddresses, y)
							return &mapperz.Mapper{
								Target: map[string]*mapperz.Mapper{
									"street": {
										Target: &y.Street,
									},
									"city": {
										Target: &y.City,
									},
								},
							}
						},
					},
				},
			},
			"numbers": {
				Target: &mapperz.Mapper{
					Target: &mapperz.Array{
						Element: func() *mapperz.Mapper {
							person.Numbers = append(person.Numbers, 0)
							return &mapperz.Mapper{
								Target: &person.Numbers[len(person.Numbers)-1],
							}
						},
					},
				},
			},
		},
	}

	str := jsonz.MustFormat(expected)
	jsonz.MustParse(str.Bytes(), mapper)

	require.Equal(t, expected, person)
}
