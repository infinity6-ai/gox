package cmpz_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/infinity6-ai/gox/commonz/cmpz"
	"github.com/stretchr/testify/assert"
)

type myComparable struct {
	val int
}

func (m myComparable) Compare(other cmpz.Comparable) int {
	o := other.(myComparable)
	if m.val < o.val {
		return -1
	}
	if m.val > o.val {
		return 1
	}
	return 0
}

func TestUnitCompareJsonNumber(t *testing.T) {
	assert.Equal(t, -1, cmpz.Compare(json.Number("1"), json.Number("2")))
	assert.Equal(t, 0, cmpz.Compare(json.Number("2"), json.Number("2")))
	assert.Equal(t, 1, cmpz.Compare(json.Number("3"), json.Number("2")))

	assert.Equal(t, -1, cmpz.Compare(json.Number("1.1"), json.Number("2")))
	assert.Equal(t, 0, cmpz.Compare(json.Number("2.1"), json.Number("2.1")))
	assert.Equal(t, 1, cmpz.Compare(json.Number("3"), json.Number("2.2")))

	assert.Panics(t, func() {
		cmpz.Compare(json.Number("b"), json.Number("A"))
	})
}

func TestUnitCompareInt(t *testing.T) {
	assert.Equal(t, -1, cmpz.Compare(1, 2))
	assert.Equal(t, 0, cmpz.Compare(2, 2))
	assert.Equal(t, 1, cmpz.Compare(3, 2))
}

func TestUnitCompareInt8(t *testing.T) {
	assert.Equal(t, -1, cmpz.Compare(int8(1), int8(2)))
	assert.Equal(t, 0, cmpz.Compare(int8(2), int8(2)))
	assert.Equal(t, 1, cmpz.Compare(int8(3), int8(2)))
}

func TestUnitCompareInt16(t *testing.T) {
	assert.Equal(t, -1, cmpz.Compare(int16(1), int16(2)))
	assert.Equal(t, 0, cmpz.Compare(int16(2), int16(2)))
	assert.Equal(t, 1, cmpz.Compare(int16(3), int16(2)))
}

func TestUnitCompareInt32(t *testing.T) {
	assert.Equal(t, -1, cmpz.Compare(int32(1), int32(2)))
	assert.Equal(t, 0, cmpz.Compare(int32(2), int32(2)))
	assert.Equal(t, 1, cmpz.Compare(int32(3), int32(2)))
}

func TestUnitCompareInt64(t *testing.T) {
	assert.Equal(t, -1, cmpz.Compare(int64(1), int64(2)))
	assert.Equal(t, 0, cmpz.Compare(int64(2), int64(2)))
	assert.Equal(t, 1, cmpz.Compare(int64(3), int64(2)))
}

func TestUnitCompareUint(t *testing.T) {
	assert.Equal(t, -1, cmpz.Compare(uint(1), uint(2)))
	assert.Equal(t, 0, cmpz.Compare(uint(2), uint(2)))
	assert.Equal(t, 1, cmpz.Compare(uint(3), uint(2)))
}

func TestUnitCompareUint8(t *testing.T) {
	assert.Equal(t, -1, cmpz.Compare(uint8(1), uint8(2)))
	assert.Equal(t, 0, cmpz.Compare(uint8(2), uint8(2)))
	assert.Equal(t, 1, cmpz.Compare(uint8(3), uint8(2)))
}

func TestUnitCompareUint16(t *testing.T) {
	assert.Equal(t, -1, cmpz.Compare(uint16(1), uint16(2)))
	assert.Equal(t, 0, cmpz.Compare(uint16(2), uint16(2)))
	assert.Equal(t, 1, cmpz.Compare(uint16(3), uint16(2)))
}

func TestUnitCompareUint32(t *testing.T) {
	assert.Equal(t, -1, cmpz.Compare(uint32(1), uint32(2)))
	assert.Equal(t, 0, cmpz.Compare(uint32(2), uint32(2)))
	assert.Equal(t, 1, cmpz.Compare(uint32(3), uint32(2)))
}

func TestUnitCompareUint64(t *testing.T) {
	assert.Equal(t, -1, cmpz.Compare(uint64(1), uint64(2)))
	assert.Equal(t, 0, cmpz.Compare(uint64(2), uint64(2)))
	assert.Equal(t, 1, cmpz.Compare(uint64(3), uint64(2)))
}

func TestUnitCompareUintptr(t *testing.T) {
	assert.Equal(t, -1, cmpz.Compare(uintptr(1), uintptr(2)))
	assert.Equal(t, 0, cmpz.Compare(uintptr(2), uintptr(2)))
	assert.Equal(t, 1, cmpz.Compare(uintptr(3), uintptr(2)))
}

func TestUnitCompareFloat32(t *testing.T) {
	assert.Equal(t, -1, cmpz.Compare(float32(1.1), float32(1.2)))
	assert.Equal(t, 0, cmpz.Compare(float32(1.2), float32(1.2)))
	assert.Equal(t, 1, cmpz.Compare(float32(1.3), float32(1.2)))
}

func TestUnitCompareFloat64(t *testing.T) {
	assert.Equal(t, -1, cmpz.Compare(1.1, 1.2))
	assert.Equal(t, 0, cmpz.Compare(1.2, 1.2))
	assert.Equal(t, 1, cmpz.Compare(1.3, 1.2))
}

func TestUnitCompareString(t *testing.T) {
	assert.Equal(t, -1, cmpz.Compare("a", "b"))
	assert.Equal(t, 0, cmpz.Compare("b", "b"))
	assert.Equal(t, 1, cmpz.Compare("c", "b"))
}

func TestUnitCompareTime(t *testing.T) {
	t1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)

	// Comparing time.Time values
	assert.Equal(t, -1, cmpz.Compare(t1, t2))
	assert.Equal(t, 0, cmpz.Compare(t1, t1))
	assert.Equal(t, 1, cmpz.Compare(t2, t1))

	// Comparing *time.Time values
	assert.Equal(t, -1, cmpz.Compare(&t1, &t2))
	assert.Equal(t, 0, cmpz.Compare(&t1, &t1))
	assert.Equal(t, 1, cmpz.Compare(&t2, &t1))

	// Comparing time.Time and *time.Time
	assert.Equal(t, 0, cmpz.Compare(any(t1), any(&t1)))
	assert.Equal(t, 0, cmpz.Compare(any(&t1), any(t1)))

	// Comparing with nil pointers
	assert.Equal(t, -1, cmpz.Compare((*time.Time)(nil), &t1))
	assert.Equal(t, 1, cmpz.Compare(&t1, (*time.Time)(nil)))
	assert.Equal(t, 0, cmpz.Compare((*time.Time)(nil), (*time.Time)(nil)))
}

func TestUnitCompareComparable(t *testing.T) {
	assert.Equal(t, -1, cmpz.Compare(myComparable{1}, myComparable{2}))
	assert.Equal(t, 0, cmpz.Compare(myComparable{2}, myComparable{2}))
	assert.Equal(t, 1, cmpz.Compare(myComparable{3}, myComparable{2}))
}

func TestUnitComparePanicOnUnsupportedType(t *testing.T) {
	type unsupported struct{}
	assert.Panics(t, func() { cmpz.Compare(unsupported{}, unsupported{}) })
}

func TestUnitSort(t *testing.T) {
	intDataAsc := []int{5, 2, 8, 1, 9}
	cmpz.Sort(intDataAsc, false)
	assert.Equal(t, []int{1, 2, 5, 8, 9}, intDataAsc)

	intDataDesc := []int{5, 2, 8, 1, 9}
	cmpz.Sort(intDataDesc, true)
	assert.Equal(t, []int{9, 8, 5, 2, 1}, intDataDesc)

	stringDataAsc := []string{"banana", "apple", "cherry"}
	cmpz.Sort(stringDataAsc, false)
	assert.Equal(t, []string{"apple", "banana", "cherry"}, stringDataAsc)

	stringDataDesc := []string{"banana", "apple", "cherry"}
	cmpz.Sort(stringDataDesc, true)
	assert.Equal(t, []string{"cherry", "banana", "apple"}, stringDataDesc)

	comparableData := []myComparable{{3}, {1}, {2}}
	cmpz.Sort(comparableData, false)
	assert.Equal(t, []myComparable{{1}, {2}, {3}}, comparableData)
}
