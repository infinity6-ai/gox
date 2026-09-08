package storezutil

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/jsonz"
	"github.com/infinity6-ai/gox/storez/storezutil/structz"
)

func SetKey(entity any, key string) {
	value := reflect.ValueOf(entity).Elem()

	if value.Kind() != reflect.Struct {
		panic(fmt.Sprintf("expected a struct, got %v", value.Kind()))
	}

	for i := 0; i < value.NumField(); i++ {
		field := value.Type().Field(i)
		if field.Name == "Id" {
			value.Field(i).SetString(key)
		}
	}
}

func GetKeys[T any](entities []T) []string {

	rows := structz.StructsToMapList(entities)
	keys := make([]string, len(entities))

	for i, row := range rows {
		keys[i] = row["id"].(string)
	}

	return keys
}

func GetKey(entity any) string {
	value := reflect.ValueOf(entity).Elem()

	if value.Kind() != reflect.Struct {
		panic(fmt.Sprintf("expected a struct, got %v", value.Kind()))
	}

	for i := 0; i < value.NumField(); i++ {
		field := value.Type().Field(i)
		if field.Name == "Id" {
			return value.Field(i).String()
		}
	}
	panic(fmt.Sprintf("Id field is required: %#v", entity))
}

func ParseRow(row string) map[string]any {
	ret := map[string]any{}
	jsonz.MustParse(row, &ret)
	FixRow(ret)
	return ret
}

func ParseRows(rows []string) []map[string]any {
	ret := []map[string]any{}
	for _, row := range rows {
		ret = append(ret, ParseRow(row))
	}
	return ret

}

func FixRow(ret map[string]any) {
	for k, v := range ret {
		if elements, ok := v.([]any); ok {
			for i, element := range elements {
				if num, ok := element.(json.Number); ok {
					intVal, err := num.Int64()
					errorz.Check(err)
					elements[i] = intVal
				}
			}
		}
		if num, ok := v.(json.Number); ok {
			intVal, err := num.Int64()
			errorz.Check(err)
			ret[k] = intVal
		}
	}
}
