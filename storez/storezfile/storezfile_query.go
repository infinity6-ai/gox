package storezfile

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/infinity6-ai/gox/commonz/cmpz"
	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/gobz"
	"github.com/infinity6-ai/gox/cryptz/cryptzb64"
	"github.com/infinity6-ai/gox/storez/storez"
	"golang.org/x/exp/constraints"
)

const MAX = 1000000

func queryCompare[T constraints.Ordered](filter storez.Filter, value T) bool {
	expected := filter.Value.(T)
	switch filter.Op {
	case "=":
		if value != expected {
			return false
		}
	case ">=":
		if value < expected {
			return false
		}
	case "<=":
		if value > expected {
			return false
		}
	case ">":
		if value <= expected {
			return false
		}
	case "<":
		if value >= expected {
			return false
		}
	default:
		panic(fmt.Sprintf("Unknown op %s: %s", filter.Field, filter.Op))
	}
	return true
}

func queryFilterValue(filter storez.Filter, value any) bool {
	if value == nil {
		return false
	}
	if s, ok := value.(int64); ok {
		return queryCompare(filter, s)
	}
	return queryCompare(filter, value.(string))
}

func queryFilter(query *storez.Query, row map[string]*storez.Value) bool {
	for _, filter := range query.Filters {
		value := row[filter.Field]
		if value == nil {
			return false
		} else if s, ok := value.Value.([]any); ok {
			sRet := false
			for _, element := range s {
				if queryFilterValue(filter, element) {
					sRet = true
					break
				}
			}
			if !sRet {
				return false
			}
		} else if !queryFilterValue(filter, value.Value) {
			return false
		}
	}
	return true
}

type CursorEntry struct {
	OrderBys storez.OrderBy `json:"o"`
	Value    any            `json:"v"`
}

type Cursor struct {
	Entries []*CursorEntry `json:"e"`
}

func (c *Cursor) Compare(row map[string]*storez.Value) int {
	for _, entry := range c.Entries {
		column := entry.OrderBys.Field
		rowValue := row[column]
		ret := 0
		if rowValue == nil {
			ret = 1
		} else {
			ret = cmpz.Compare(entry.Value, rowValue.Value)
		}
		if ret != 0 {
			if entry.OrderBys.Desc {
				ret = ret * -1
			}
			return ret
		}
	}
	return 0
}

func (c *Cursor) Format() string {
	g := gobz.MustFormat(c)
	ret := cryptzb64.UrlEncode(g).String()
	return ret
}

func CursorParse(str string) *Cursor {
	if str == "" {
		return nil
	}
	b := errorz.Check2(cryptzb64.UrlDecode(str)).Bytes()
	ret := gobz.MustParse(b, &Cursor{})
	return ret
}

func NewCursor(orderBys []storez.OrderBy, row map[string]*storez.Value) *Cursor {
	ret := &Cursor{Entries: []*CursorEntry{}}
	for _, orderBy := range orderBys {
		ret.Entries = append(ret.Entries, &CursorEntry{
			OrderBys: orderBy,
			Value:    row[orderBy.Field].Value,
		})
	}
	return ret
}

func (me *StorezStrategyFile) Query(ctx context.Context, query *storez.Query) (string, []map[string]*storez.Value) {
	db := me.getOrCreateDB(query.Table)

	query = gobz.MustClone(query, &storez.Query{})
	query.OrderBys = append([]storez.OrderBy{}, query.OrderBys...)
	query.OrderBys = append(query.OrderBys, storez.OrderBy{Field: "id"})

	ret := []map[string]*storez.Value{}
	db.Walk(func(row map[string]*storez.Value) {
		if !queryFilter(query, row) {
			return
		}
		ret = append(ret, row)
	})

	sortByColumns(ret, query.OrderBys)

	offset := 0
	startCursor := CursorParse(query.StartCursor)
	if startCursor != nil {
		for _, row := range ret {
			if startCursor.Compare(row) < 0 {
				break
			}
			offset++
		}
	}

	if offset >= len(ret) {
		ret = []map[string]*storez.Value{}
		return query.StartCursor, ret
	}

	ret = ret[offset:]

	if len(ret) > query.Limit {
		ret = ret[:query.Limit]
	}

	endCursor := NewCursor(query.OrderBys, ret[len(ret)-1])

	return endCursor.Format(), ret
}

func sortByColum(i map[string]*storez.Value, j map[string]*storez.Value, column string) int {
	vi, viOk := i[column]
	vj, vjOk := j[column]

	if !viOk && !vjOk {
		return 0
	}
	if !viOk || !vjOk {
		if vjOk {
			return -1
		} else {
			return 1
		}
	}
	if si, ok := vi.Value.(int64); ok {
		sj := vj.Value.(int64)
		ret := si - sj
		if ret < 0 {
			return -1
		} else if ret > 0 {
			return 1
		} else {
			return 0
		}
	}
	si := vi.Value.(string)
	sj := vj.Value.(string)
	return strings.Compare(si, sj)
}

func sortByColumns(data []map[string]*storez.Value, columns []storez.OrderBy) {
	sort.Slice(data, func(i, j int) bool {
		rowi := data[i]
		rowj := data[j]
		ret := 0
		for _, column := range columns {
			ret = sortByColum(rowi, rowj, column.Field)
			if column.Desc {
				ret = ret * -1
			}
			if ret != 0 {
				return ret < 0
			}
		}
		return false
	})
}
