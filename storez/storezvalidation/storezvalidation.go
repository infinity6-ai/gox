package storezvalidation

import (
	"fmt"
	"reflect"
	"regexp"

	"github.com/infinity6-ai/gox/commonz/validation/checker"
)

var regexIdSimple = regexp.MustCompile(`^[A-Za-z0-9\-]+$`)
var regexId = regexp.MustCompile(`^[A-Za-z0-9=_\-\.]+$`)

func TableName(TableName string) {
	checker.RegexMatch(regexIdSimple, TableName, "storez-table-name: %s", TableName)
}

func Id(ids ...string) {
	for _, id := range ids {
		checker.RegexMatch(regexId, id, "storez-id")
	}
}

func ValueSimple(values ...any) {
	for _, v := range values {
		if v == nil {
			continue
		}
		switch t := v.(type) {
		case string, int64:
			continue
		case []any:
			for _, elem := range t {
				checker.True(reflect.TypeOf(elem).Kind() != reflect.Slice, "slices of slices")
				ValueSimple(elem)
			}
		case []string, []int64:
			continue
		default:
			panic(fmt.Errorf("Invalid type: %s", t))
		}
	}
}

func Value(values ...any) {
	for _, v := range values {
		ValueSimple(v)
	}
}

func Row(rows ...map[string]any) {
	for _, row := range rows {
		for k, v := range row {
			if k == "id" {
				Id(v.(string))
			} else {
				Value(v)
			}
		}
	}
}
