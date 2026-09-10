package mapperz_test

import (
	"testing"

	"github.com/infinity6-ai/gox/commonz/jsonz"
	"github.com/infinity6-ai/gox/schemaz/mapperz"
	"github.com/stretchr/testify/require"
)

func TestUnitBasic(t *testing.T) {

	type Person struct {
		Name string
		Age  int
	}

	expected := Person{
		Name: "John",
		Age:  30,
	}

	var person Person

	mapper := &mapperz.Mapper{
		Target: map[string]any{
			"name": &mapperz.Mapper{
				Name:   "name",
				Target: &person.Name,
			},
			"age": &mapperz.Mapper{
				Name:   "age",
				Target: &person.Age,
			},
		},
	}

	str := jsonz.MustFormat(expected)
	jsonz.MustParse(str.Bytes(), mapper)

	require.Equal(t, expected, person)
}
