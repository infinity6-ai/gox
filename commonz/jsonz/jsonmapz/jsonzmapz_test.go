package jsonmapz_test

import (
	"encoding/json"
	"testing"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/jsonz/jsonmapz"
	"github.com/stretchr/testify/assert"
)

func TestUnitHowTo(t *testing.T) {

	type Data struct {
		Id  string
		Age int
	}

	out := map[string]any{
		"a": "",         // string type gets the first string as is
		"b": []string{}, // []string gets a copy []string
		"c": float64(0), // every other type should use jsonz.Parse
		"d": []float64{},
		"e": &Data{},
		"f": []*Data{},
		"h": map[string]any{},
		"i": []map[string]any{},
	}

	unparsed := map[string][]string{
		"a": {"a1", "a2"},
		"b": {"b1", "b2"},
		"c": {"10.2", "11.2"},
		"d": {"20.2", "21.2"},
		"e": {"{\"Id\":\"e1\", \"Age\":10}", "{\"Id\":\"e2\", \"Age\":11}"},
		"f": {"{\"Id\":\"f1\", \"Age\":20}", "{\"Id\":\"f2\", \"Age\":21}"},
		"g": {"ignored"},
		"h": {"{\"Id\":\"h1\", \"Age\":30}", "{\"Id\":\"h2\", \"Age\":31}"},
		"i": {"{\"Id\":\"i1\", \"Age\":40}", "{\"Id\":\"i2\", \"Age\":41}"},
	}

	expected := map[string]any{
		"a": "a1",
		"b": []string{"b1", "b2"},
		"c": 10.2,
		"d": []float64{20.2, 21.2},
		"e": &Data{Id: "e1", Age: 10},
		"f": []*Data{
			{Id: "f1", Age: 20},
			{Id: "f2", Age: 21},
		},
		"h": map[string]any{"Id": "h1", "Age": json.Number("30")},
		"i": []map[string]any{
			{"Id": "i1", "Age": json.Number("40")},
			{"Id": "i2", "Age": json.Number("41")},
		},
	}

	err := jsonmapz.ParseMap(unparsed, out)
	errorz.Check(err)

	assert.Equal(t, expected, out)

}
