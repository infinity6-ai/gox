package storezdatastore

import (
	"fmt"
	"reflect"

	"cloud.google.com/go/datastore"
	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/storez/storez"
)

func HandleMultiError(err error, keys []*datastore.Key, rows []datastore.PropertyList) []map[string]*storez.Value {
	ret := []map[string]*storez.Value{}
	if err != nil {
		multiErrors, ok := err.(datastore.MultiError)
		if !ok {
			errorz.Check(err)
		}
		for i, merr := range multiErrors {
			if merr == nil {
				ret = append(ret, Row2Map(keys[i].Name, rows[i]))
			} else if merr != datastore.ErrNoSuchEntity {
				errorz.Check(merr)
			}
		}
	} else {
		ret = Rows2Map(keys, rows)
	}
	return ret
}

func Rows2Map(keys []*datastore.Key, rows []datastore.PropertyList) []map[string]*storez.Value {
	ret := []map[string]*storez.Value{}
	for idx, row := range rows {
		key := keys[idx]
		ret = append(ret, Row2Map(key.Name, row))
	}
	return ret
}

func Row2Map(id string, row datastore.PropertyList) map[string]*storez.Value {
	ret := map[string]*storez.Value{}
	for _, p := range row {
		if p.Name == "id" {
			panic("id must not exists in PropertyList")
		}
		ret[p.Name] = &storez.Value{
			Value:   p.Value,
			Indexed: !p.NoIndex,
		}
	}
	ret["id"] = &storez.Value{
		Value:   id,
		Indexed: true,
	}
	return ret
}

func Keys2DS(table string, ids []string) []*datastore.Key {
	ret := []*datastore.Key{}
	for _, id := range ids {
		ret = append(ret, Key2DS(table, id))
	}
	return ret
}

func Key2DS(table string, id string) *datastore.Key {
	return datastore.NameKey(table, id, nil)
}

func Rows2DS(tableName string, rows []map[string]*storez.Value) ([]*datastore.Key, []datastore.PropertyList) {
	entities := []datastore.PropertyList{}
	keys := GetKeys(tableName, rows)
	for _, row := range rows {
		entities = append(entities, Row2DS(row))
	}
	return keys, entities
}

func Value2DS(v any) any {
	if v == nil {
		return nil
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Slice:
		if rv.IsNil() {
			return nil
		}
		if _, ok := v.([]any); ok {
			return v
		}
		converted := make([]any, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			converted[i] = rv.Index(i).Interface()
		}
		return converted
	case reflect.String, reflect.Int64:
		return v
	}
	panic(fmt.Errorf("unsupported type: %s", rv.Kind()))
}

func Row2DS(row map[string]*storez.Value) datastore.PropertyList {
	ret := datastore.PropertyList{}
	for k, v := range row {
		if k != "id" {
			nv := Value2DS(v.Value)
			ret = append(ret, datastore.Property{Name: k, Value: nv, NoIndex: !v.Indexed})
		}
	}
	return ret
}

func GetKeys(table string, entities []map[string]*storez.Value) []*datastore.Key {
	ret := []*datastore.Key{}
	for _, entity := range entities {
		id := entity["id"].Value.(string)
		if id == "" {
			panic("id is required found")
		}
		ret = append(ret, Key2DS(table, id))
	}
	return ret
}
