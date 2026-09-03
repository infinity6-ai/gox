package structjsonz_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/jsonz/structjsonz"
	"github.com/infinity6-ai/gox/commonz/pathz"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Data struct {
	Id  string `json:"Id"`
	Age int    `json:"Age"`
}

type TestStruct struct {
	A string           `json:"a"`
	B []string         `json:"b"`
	C float64          `json:"c"`
	D []float64        `json:"d"`
	E *Data            `json:"e"`
	F []*Data          `json:"f"`
	G string           `json:"g,omitempty"`
	H map[string]any   `json:"h"`
	I []map[string]any `json:"i"`
	J time.Time        `json:"j"`
	K []time.Time      `json:"k"`
	L *pathz.Path      `json:"l"`
	M []*pathz.Path    `json:"m"`
}

func TestUnitParse(t *testing.T) {
	unparsed := map[string][]string{
		"a": {"a1", "a2"},
		"b": {"b1", "b2"},
		"c": {"10.2", "11.2"},
		"d": {"20.2", "21.2"},
		"e": {`{"Id":"e1","Age":10}`},
		"f": {`{"Id":"f1","Age":20}`, `{"Id":"f2","Age":21}`},
		"g": {"ignored"},
		"h": {`{"Id":"h1","Age":30}`},
		"i": {`{"Id":"i1","Age":40}`, `{"Id":"i2","Age":41}`},
		"j": {`"2023-01-01T12:00:00Z"`},
		"k": {`"2023-02-01T12:00:00Z"`, `"2023-03-01T12:00:00Z"`},
		"l": {`"/a/b"`},
		"m": {`"c/d"`, `"../e"`},
	}

	var got TestStruct
	err := structjsonz.Parse(unparsed, &got)
	require.NoError(t, err)

	t1, _ := time.Parse(time.RFC3339, "2023-01-01T12:00:00Z")
	t2, _ := time.Parse(time.RFC3339, "2023-02-01T12:00:00Z")
	t3, _ := time.Parse(time.RFC3339, "2023-03-01T12:00:00Z")
	p1, _ := pathz.Parse("/a/b")
	p2, _ := pathz.Parse("c/d")
	p3, _ := pathz.Parse("../e")

	expected := TestStruct{
		A: "a1",
		B: []string{"b1", "b2"},
		C: 10.2,
		D: []float64{20.2, 21.2},
		E: &Data{Id: "e1", Age: 10},
		F: []*Data{
			{Id: "f1", Age: 20},
			{Id: "f2", Age: 21},
		},
		H: map[string]any{"Id": "h1", "Age": float64(30)},
		I: []map[string]any{
			{"Id": "i1", "Age": float64(40)},
			{"Id": "i2", "Age": float64(41)},
		},
		J: t1,
		K: []time.Time{t2, t3},
		L: p1,
		M: []*pathz.Path{p2, p3},
	}

	assert.Equal(t, expected, got)
}

func TestUnitFormat(t *testing.T) {
	t1, _ := time.Parse(time.RFC3339, "2023-01-01T12:00:00Z")
	t2, _ := time.Parse(time.RFC3339, "2023-02-01T12:00:00Z")
	t3, _ := time.Parse(time.RFC3339, "2023-03-01T12:00:00Z")
	p1, _ := pathz.Parse("/a/b")
	p2, _ := pathz.Parse("c/d")
	p3, _ := pathz.Parse("../e")

	parsed := &TestStruct{
		A: "a1",
		B: []string{"b1", "b2"},
		C: 10.2,
		D: []float64{20.2, 21.2},
		E: &Data{Id: "e1", Age: 10},
		F: []*Data{
			{Id: "f1", Age: 20},
			{Id: "f2", Age: 21},
		},
		H: map[string]any{"Id": "h1", "Age": float64(30)},
		I: []map[string]any{
			{"Id": "i1", "Age": float64(40)},
			{"Id": "i2", "Age": float64(41)},
		},
		J: t1,
		K: []time.Time{t2, t3},
		L: p1,
		M: []*pathz.Path{p2, p3},
	}

	expected := map[string][]string{
		"a": {"a1"},
		"b": {"b1", "b2"},
		"c": {"10.2"},
		"d": {"20.2", "21.2"},
		"e": {`{"Id":"e1","Age":10}`},
		"f": {`{"Id":"f1","Age":20}`, `{"Id":"f2","Age":21}`},
		"h": {`{"Age":30,"Id":"h1"}`},
		"i": {`{"Age":40,"Id":"i1"}`, `{"Age":41,"Id":"i2"}`},
		"j": {`"2023-01-01T12:00:00Z"`},
		"k": {`"2023-02-01T12:00:00Z"`, `"2023-03-01T12:00:00Z"`},
		"l": {`"/a/b"`},
		"m": {`"c/d"`, `"../e"`},
	}

	out, err := structjsonz.Format(parsed)
	errorz.Check(err)

	// Custom comparison because of map key order randomness in JSON
	require.Equal(t, len(expected), len(out), "length of maps should be equal")

	for k, v := range expected {
		outV, ok := out[k]
		require.True(t, ok, "key %s not found in output", k)

		if k == "h" {
			var expectedMap, outMap map[string]any
			err := json.Unmarshal([]byte(v[0]), &expectedMap)
			require.NoError(t, err)
			err = json.Unmarshal([]byte(outV[0]), &outMap)
			require.NoError(t, err)
			assert.Equal(t, expectedMap, outMap)
		} else if k == "i" {
			var expectedSlice, outSlice []map[string]any
			for _, item := range v {
				var m map[string]any
				err := json.Unmarshal([]byte(item), &m)
				require.NoError(t, err)
				expectedSlice = append(expectedSlice, m)
			}
			for _, item := range outV {
				var m map[string]any
				err := json.Unmarshal([]byte(item), &m)
				require.NoError(t, err)
				outSlice = append(outSlice, m)
			}
			assert.ElementsMatch(t, expectedSlice, outSlice)
		} else {
			assert.ElementsMatch(t, v, outV)
		}
	}
}
