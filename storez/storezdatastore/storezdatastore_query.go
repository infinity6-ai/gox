package storezdatastore

import (
	"context"
	"fmt"

	"cloud.google.com/go/datastore"
	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/storez/storez"
	"google.golang.org/api/iterator"
)

func queryConvertCursor(query *storez.Query, dsquery *datastore.Query) *datastore.Query {
	if query.StartCursor != "" {
		startCursor, err := datastore.DecodeCursor(query.StartCursor)
		errorz.Check(err)
		dsquery = dsquery.Start(startCursor)
	}
	return dsquery
}

func queryConvertFilter(dsquery *datastore.Query, table string, filter *storez.Filter) *datastore.Query {
	if filter.Field == "id" {
		id := filter.Value.(string)
		key := datastore.NameKey(table, id, nil)
		return dsquery.FilterField("__key__", filter.Op, key)
	}
	return dsquery.FilterField(filter.Field, filter.Op, filter.Value)
}

func queryConvertOrderBy(query *storez.Query, dsquery *datastore.Query) *datastore.Query {
	for _, orderBy := range query.OrderBys {
		dsfield := orderBy.Field
		if dsfield == "id" {
			dsfield = "__key__"
		}
		if orderBy.Desc {
			dsfield = fmt.Sprintf("-%s", dsfield)
		}
		dsquery = dsquery.Order(dsfield)
	}
	return dsquery
}

func queryConvertProjection(query *storez.Query, dsquery *datastore.Query) *datastore.Query {
	if query.Projections == nil {
		return dsquery
	}
	if len(query.Projections) == 1 && query.Projections[0] == "id" {
		return dsquery.KeysOnly()
	}
	for _, projection := range query.Projections {
		fieldName := projection
		if fieldName == "id" {
			fieldName = "__key__"
		}
		dsquery = dsquery.Project(fieldName)
	}
	return dsquery
}

func queryConvert(query *storez.Query) *datastore.Query {
	dsquery := datastore.NewQuery(query.Table)
	dsquery = queryConvertProjection(query, dsquery)
	dsquery = queryConvertCursor(query, dsquery)
	for _, filter := range query.Filters {
		dsquery = queryConvertFilter(dsquery, query.Table, &filter)
	}
	dsquery = queryConvertOrderBy(query, dsquery)
	dsquery = dsquery.Limit(query.Limit)
	return dsquery
}

func convertQueryResultValue(it *datastore.Iterator) []map[string]*storez.Value {
	ret := []map[string]*storez.Value{}
	for {
		element := datastore.PropertyList{}
		key, err := it.Next(&element)
		if err == iterator.Done {
			it.Cursor()
			break
		}
		errorz.Check(err)
		row := map[string]*storez.Value{}
		for _, property := range element {
			row[property.Name] = &storez.Value{
				Value:   property.Value,
				Indexed: !property.NoIndex,
			}
		}
		row["id"] = &storez.Value{
			Value:   key.Name,
			Indexed: true,
		}
		ret = append(ret, row)
	}
	return ret
}

func (me *StorezStrategyDatastore) Query(ctx context.Context, query *storez.Query) (string, []map[string]*storez.Value) {
	dsquery := queryConvert(query)
	it := me.client.Run(ctx, dsquery)
	ret := convertQueryResultValue(it)
	cursor, err := it.Cursor()
	errorz.Check(err)
	return cursor.String(), ret
}
