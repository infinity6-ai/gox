package structz

import (
	"encoding/json"
	"reflect"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/jsonz"
	"github.com/infinity6-ai/gox/commonz/validation/checker"
)

func MapListToStructs[T any](data []map[string]any) []T {
	formatted := jsonz.MustFormat(data)
	ret := jsonz.MustParse(formatted.Bytes(), &[]T{})

	return *ret
}

func ToInt64(value interface{}) int64 {
	if jn, ok := value.(json.Number); ok {
		i, err := jn.Int64()
		errorz.Check(err)
		return i
	}
	v := reflect.ValueOf(value)

	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int()
	case reflect.Float32, reflect.Float64:
		return int64(v.Float())

	}
	checker.Fail("Unsupported type: %s", v.Kind())
	return 0
}

func StructsToMapList[T any](data []T) []map[string]any {
	formatted := jsonz.MustFormat(data)
	parsed := *jsonz.MustParse(formatted.Bytes(), &[]map[string]any{})

	ret := make([]map[string]any, len(parsed))
	for i, row := range parsed {
		ret[i] = map[string]any{}
		for k, v := range row {
			if jn, ok := v.(json.Number); ok {
				ret[i][k] = ToInt64(jn)
				continue
			}
			val := reflect.ValueOf(v)

			switch val.Kind() {
			case reflect.Int64, reflect.String:
				ret[i][k] = v
			case reflect.Int16, reflect.Int32, reflect.Float32, reflect.Float64:
				ret[i][k] = ToInt64(v)

			case reflect.Slice:

				result := make([]any, val.Len())

				for j := 0; j < val.Len(); j++ {
					switch val.Index(j).Interface().(type) {
					case string:
						result[j] = val.Index(j).Interface().(string)
					default:
						result[j] = ToInt64(val.Index(j).Interface())
					}

				}

				ret[i][k] = result

			default:
				checker.Nil(v, "Unsupported type: %s", val.Kind())
				ret[i][k] = nil

			}

		}
	}

	return ret
}
