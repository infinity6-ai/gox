package cmpz

import (
	"cmp"
	"encoding/json"
	"fmt"
	"slices"
	"time"

	"github.com/infinity6-ai/gox/commonz/errorz"
)

type Comparable interface {
	Compare(other Comparable) int
}

func Compare[T any](a, b T) int {
	switch v1 := any(a).(type) {
	case int:
		return cmp.Compare(v1, any(b).(int))
	case int8:
		return cmp.Compare(v1, any(b).(int8))
	case int16:
		return cmp.Compare(v1, any(b).(int16))
	case int32:
		return cmp.Compare(v1, any(b).(int32))
	case int64:
		return cmp.Compare(v1, any(b).(int64))
	case uint:
		return cmp.Compare(v1, any(b).(uint))
	case uint8:
		return cmp.Compare(v1, any(b).(uint8))
	case uint16:
		return cmp.Compare(v1, any(b).(uint16))
	case uint32:
		return cmp.Compare(v1, any(b).(uint32))
	case uint64:
		return cmp.Compare(v1, any(b).(uint64))
	case uintptr:
		return cmp.Compare(v1, any(b).(uintptr))
	case float32:
		return cmp.Compare(v1, any(b).(float32))
	case float64:
		return cmp.Compare(v1, any(b).(float64))
	case string:
		return cmp.Compare(v1, any(b).(string))
	case time.Time, *time.Time:
		return compareTime(a, b)
	case json.Number:
		n1 := any(a).(json.Number)
		n2 := any(b).(json.Number)
		return compareJsonNumber(n1, n2)
	case Comparable:
		return v1.Compare(any(b).(Comparable))
	default:
		panic(fmt.Errorf("type %T is not supported by cmp.Compare", v1))
	}
}

func compareTime(a, b any) int {
	n1 := parseTime(a)
	n2 := parseTime(b)
	if n1 == nil && n2 == nil {
		return 0
	}
	if n1 == nil && n2 != nil {
		return -1
	}
	if n1 != nil && n2 == nil {
		return 1
	}
	return n1.Compare(*n2)
}

func parseTime(a any) *time.Time {
	ret, ok := a.(*time.Time)
	if !ok {
		v := any(a).(time.Time)
		ret = &v
	}
	return ret
}

func compareJsonNumber(n1, n2 json.Number) int {
	i1, e1 := n1.Int64()
	i2, e2 := n2.Int64()
	if e1 == nil || e2 == nil {
		return Compare(i1, i2)
	}
	f1, e1 := n1.Float64()
	errorz.Check(e1)
	f2, e2 := n2.Float64()
	errorz.Check(e2)
	return Compare(f1, f2)
}

func Sort[T any](data []T, desc bool) {
	slices.SortFunc(data, func(a, b T) int {
		ret := Compare(a, b)
		if desc {
			return ret * -1
		}
		return ret
	})
}
