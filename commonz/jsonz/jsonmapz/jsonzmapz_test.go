package jsonmapz_test

import (
	"testing"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/jsonz/jsonmapz"
	"github.com/stretchr/testify/assert"
)

func TestUnitHowTo(t *testing.T) {

	out := map[string]any{
		"a": "",
		"b": []string{},
		"c": float64(0),
		"d": []float64{},
	}

	unparsed := map[string][]string{
		"a": {"a1", "a2"},
		"b": {"b1", "b2"},
		"c": {"10.2", "11.2"},
		"d": {"20.2", "21.2"},
		"g": {"ignored"},
	}

	expected := map[string]any{
		"a": "a1",
		"b": []string{"b1", "b2"},
		"c": 10.2,
		"d": []float64{20.2, 21.2},
	}

	err := jsonmapz.ParseMap(unparsed, out)
	errorz.Check(err)

	assert.Equal(t, expected, out)

}
