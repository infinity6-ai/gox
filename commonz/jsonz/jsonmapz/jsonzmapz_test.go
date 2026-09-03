package jsonmapz_test

import (
	"testing"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/jsonz/jsonmapz"
	"github.com/stretchr/testify/assert"
)

func TestPocHowTo(t *testing.T) {

	type Data struct {
		Id  string
		Age int
	}

	out := map[string]any{
		"a": "",
		"b": []string{},
		"c": float64(0),
		"d": []float64{},
		"e": &Data{},
		"f": []*Data{},
	}

	unparsed := map[string][]string{
		"a": {"a1", "a2"},
		"b": {"b1", "b2"},
		"c": {"10.2", "11.2"},
		"d": {"20.2", "21.2"},
		"e": {"{\"Id\":\"e1\", \"Age\":10}", "{\"Id\":\"e2\", \"Age\":11}"},
		"f": {"{\"Id\":\"f1\", \"Age\":20}", "{\"Id\":\"f2\", \"Age\":21}"},
	}

	expected := map[string]any{
		"a": "a1",
		"b": []string{"b1", "b2"},
		"c": 10.2,
		"d": []float64{20.2, 21.2},
		"e": &Data{Id: "e1", Age: 10},
		"F": []*Data{
			{Id: "f1", Age: 20},
			{Id: "f2", Age: 21},
		},
	}

	err := jsonmapz.ParseMap(unparsed, out)
	errorz.Check(err)

	assert.Equal(t, expected, out)

}
