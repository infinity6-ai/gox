package structz_test

import (
	"testing"

	"github.com/infinity6-ai/gox/storez/storezutil/structz"
	"github.com/stretchr/testify/assert"
)

type MyData struct {
	Id     int64 `json:"id"`
	Name   string
	Scores []int64
	Descs  []string
}

type UnsupportedMyData struct {
	Id MyData `json:"id"`
}

type UnsupportedMyDataArray struct {
	Id []MyData `json:"id"`
}

func TestUnitStructz(t *testing.T) {
	a := []MyData{{int64(1), "b", []int64{1, 2}, []string{"d1", "d2"}}}
	b := []map[string]any{{"id": int64(1), "Name": "b", "Scores": []any{int64(1), int64(2)}, "Descs": []any{"d1", "d2"}}}

	assert.Equal(t, a, structz.MapListToStructs[MyData](b))
	assert.Equal(t, b, structz.StructsToMapList(a))

	c := []*MyData{{int64(1), "b", []int64{1, 2}, []string{"d1", "d2"}}}
	d := []map[string]any{{"id": int64(1), "Name": "b", "Scores": []any{int64(1), int64(2)}, "Descs": []any{"d1", "d2"}}}

	assert.Equal(t, c, structz.MapListToStructs[*MyData](b))
	assert.Equal(t, d, structz.StructsToMapList(c))

	a = []MyData{{int64(1), "b", []int64{1, 2}, nil}}
	b = []map[string]any{{"id": int64(1), "Name": "b", "Scores": []any{int64(1), int64(2)}, "Descs": nil}}

	assert.Equal(t, b, structz.StructsToMapList(a))

	u := []UnsupportedMyData{{a[0]}}

	assert.Panics(t, func() { structz.StructsToMapList(u) })

	uarray := []UnsupportedMyDataArray{{a}}

	assert.Panics(t, func() { structz.StructsToMapList(uarray) })
}
