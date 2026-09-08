package storez

import (
	"encoding/json"

	"github.com/infinity6-ai/gox/commonz/errorz"
)

func row2Value(schema *StorezSchemaTable, row map[string]any) map[string]*Value {
	ret := map[string]*Value{}
	for k, v := range row {
		columnSchema := schema.GetColumn(k)
		ret[k] = &Value{
			Value:   v,
			Indexed: columnSchema.Indexed,
		}
	}
	return ret
}

func rows2Value(schema *StorezSchemaTable, rows []map[string]any) []map[string]*Value {
	ret := []map[string]*Value{}
	for _, row := range rows {
		rowValue := row2Value(schema, row)
		ret = append(ret, rowValue)
	}
	return ret
}

func rowFromValue(row map[string]*Value) map[string]any {
	ret := map[string]any{}
	for k, v := range row {
		ret[k] = v.Value
	}
	return ret
}

func rowsFromValue(rows []map[string]*Value) []map[string]any {
	ret := []map[string]any{}
	for _, row := range rows {
		rowValue := rowFromValue(row)
		ret = append(ret, rowValue)
	}
	return ret
}

func FixRow(ret map[string]*Value) {
	for k, v := range ret {
		if elements, ok := v.Value.([]any); ok {
			for i, element := range elements {
				if num, ok := element.(json.Number); ok {
					intVal, err := num.Int64()
					errorz.Check(err)
					elements[i] = intVal
				}
			}
		}
		if num, ok := v.Value.(json.Number); ok {
			intVal, err := num.Int64()
			errorz.Check(err)
			ret[k].Value = intVal
		}
	}
}
