package jsonmapz_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/jsonz/jsonmapz"
	"github.com/infinity6-ai/gox/commonz/pathz"
	"github.com/stretchr/testify/assert"
)

func TestUnitHowTo(t *testing.T) {

	type Data struct {
		Id  string
		Age int
	}

	t1, err := time.Parse(time.RFC3339, "2023-01-01T12:00:00Z")
	errorz.Check(err)
	t2, err := time.Parse(time.RFC3339, "2023-02-01T12:00:00Z")
	errorz.Check(err)
	t3, err := time.Parse(time.RFC3339, "2023-03-01T12:00:00Z")
	errorz.Check(err)
	p1, err := pathz.Parse("/a/b")
	errorz.Check(err)
	p2, err := pathz.Parse("c/d")
	errorz.Check(err)
	p3, err := pathz.Parse("../e")
	errorz.Check(err)

	out := map[string]any{
		"a": "",
		"b": []string{},
		"c": float64(0),
		"d": []float64{},
		"e": &Data{},
		"f": []*Data{},
		"h": map[string]any{},
		"i": []map[string]any{},
		"j": time.Time{},
		"k": []time.Time{},
		"l": &pathz.Path{},
		"m": []*pathz.Path{},
	}

	unparsed := map[string][]string{
		"a": {"a1", "a2"},
		"b": {"b1", "b2"},
		"c": {"10.2", "11.2"},
		"d": {"20.2", "21.2"},
		"e": {`{"Id":"e1", "Age":10}`, `{"Id":"e2", "Age":11}`},
		"f": {`{"Id":"f1", "Age":20}`, `{"Id":"f2", "Age":21}`},
		"g": {"ignored"},
		"h": {`{"Id":"h1", "Age":30}`, `{"Id":"h2", "Age":31}`},
		"i": {`{"Id":"i1", "Age":40}`, `{"Id":"i2", "Age":41}`},
		"j": {`"2023-01-01T12:00:00Z"`},
		"k": {`"2023-02-01T12:00:00Z"`, `"2023-03-01T12:00:00Z"`},
		"l": {`"/a/b"`},
		"m": {`"c/d"`, `"../e"`},
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
		"j": t1,
		"k": []time.Time{t2, t3},
		"l": p1,
		"m": []*pathz.Path{p2, p3},
	}

	// for each unparsed entry value we need to:
	// - ignore if it is not in out map
	// - if out[key] is string (use first) or []string just copy it
	// - else if out[key] is not a slice, jsonz.Parse the unparsed[key][0]
	// - else (it is a slice) jsonz.Parse the unparsed[key] into out[key]

	err = jsonmapz.ParseMap(unparsed, out)
	errorz.Check(err)

	assert.Equal(t, expected, out)

}
