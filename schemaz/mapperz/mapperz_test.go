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
		Name               string   `json:"name"`
		Age                int      `json:"age"`
		MainAddress        *Address `json:"main_address"`
		SecondaryAddresses *Address `json:"secondary_addresses"`
		// Addresses   []Address `json:"addresses"`
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
	}

	var person Person
	person.SecondaryAddresses = &Address{}

	mapper := &mapperz.Mapper{
		Target: map[string]*mapperz.Mapper{
			"name": {
				Name:   "name",
				Target: &person.Name,
			},
			"age": {
				Name:   "age",
				Target: &person.Age,
			},
			"main_address": {
				Name:   "main_address",
				Target: &person.MainAddress,
			},
			"secondary_addresses": {
				Name: "secondary_addresses",
				Target: &mapperz.Mapper{
					Target: map[string]*mapperz.Mapper{
						"street": {
							Name:   "street",
							Target: &person.SecondaryAddresses.Street,
						},
						"city": {
							Name:   "city",
							Target: &person.SecondaryAddresses.City,
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
