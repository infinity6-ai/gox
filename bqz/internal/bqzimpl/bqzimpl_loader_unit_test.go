package bqzimpl_test

import (
	"testing"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/infinity6-ai/gox/bqz/internal/bqzimpl"
	"github.com/stretchr/testify/require"
)

type innerTestStruct struct {
	Name string `bigquery:"name"`
}

type testStruct struct {
	ID        string           `json:"custom_id" bigquery:"ignored_bq_id"`
	Age       int              `bigquery:"age"`
	Score     float64          `bigquery:"score"`
	IsActive  bool             `bigquery:"is_active"`
	NullStr   *string          `bigquery:"null_str"`
	NullInt   *int64           `bigquery:"null_int"`
	NullFlt   *float64         `bigquery:"null_flt"`
	NullBl    *bool            `bigquery:"null_bl"`
	NullTm    *time.Time       `bigquery:"null_tm"`
	Inner     innerTestStruct  `bigquery:"inner"`
	InnerPtr  *innerTestStruct `bigquery:"inner_ptr"`
	StringPtr *string          `bigquery:"string_ptr"`
	Repeated  []string         `bigquery:"repeated"`
}

func TestUnitLoadRow(t *testing.T) {
	schema := bigquery.Schema{
		{Name: "ignored_bq_id", Type: bigquery.StringFieldType},
		{Name: "age", Type: bigquery.IntegerFieldType},
		{Name: "score", Type: bigquery.FloatFieldType},
		{Name: "is_active", Type: bigquery.BooleanFieldType},
		{Name: "null_str", Type: bigquery.StringFieldType},
		{Name: "null_int", Type: bigquery.IntegerFieldType},
		{Name: "null_flt", Type: bigquery.FloatFieldType},
		{Name: "null_bl", Type: bigquery.BooleanFieldType},
		{Name: "null_tm", Type: bigquery.TimestampFieldType},
		{
			Name: "inner",
			Type: bigquery.RecordFieldType,
			Schema: bigquery.Schema{
				{Name: "name", Type: bigquery.StringFieldType},
			},
		},
		{
			Name: "inner_ptr",
			Type: bigquery.RecordFieldType,
			Schema: bigquery.Schema{
				{Name: "name", Type: bigquery.StringFieldType},
			},
		},
		{Name: "string_ptr", Type: bigquery.StringFieldType},
		{Name: "repeated", Type: bigquery.StringFieldType, Repeated: true},
	}

	now := time.Now().Truncate(time.Second)

	type testScenario struct {
		values []bigquery.Value
		want   testStruct
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		var got testStruct
		err := bqzimpl.LoadRow(&got, schema, s.values)
		require.NoError(t, err)
		require.Equal(t, s.want, got)
	}

	t.Run("Map row with fully populated valid values", func(t *testing.T) {
		strPtr := "hello"
		nullStr := "nullable-string"
		nullInt := int64(42)
		nullFlt := 1.23
		nullBl := true
		check(t, testScenario{
			values: []bigquery.Value{
				"abc",
				int64(25),
				3.14,
				true,
				"nullable-string",
				int64(42),
				1.23,
				true,
				now,
				[]bigquery.Value{"nested-struct"},
				[]bigquery.Value{"nested-ptr-struct"},
				"hello",
				[]bigquery.Value{"a", "b", "c"},
			},
			want: testStruct{
				ID:        "abc",
				Age:       25,
				Score:     3.14,
				IsActive:  true,
				NullStr:   &nullStr,
				NullInt:   &nullInt,
				NullFlt:   &nullFlt,
				NullBl:    &nullBl,
				NullTm:    &now,
				Inner:     innerTestStruct{Name: "nested-struct"},
				InnerPtr:  &innerTestStruct{Name: "nested-ptr-struct"},
				StringPtr: &strPtr,
				Repeated:  []string{"a", "b", "c"},
			},
		})
	})

	t.Run("Map row with null values", func(t *testing.T) {
		check(t, testScenario{
			values: []bigquery.Value{
				"abc",
				int64(25),
				3.14,
				true,
				nil,
				nil,
				nil,
				nil,
				nil,
				[]bigquery.Value{"nested-struct"},
				nil,
				nil,
				[]bigquery.Value{"a", "b", "c"},
			},
			want: testStruct{
				ID:        "abc",
				Age:       25,
				Score:     3.14,
				IsActive:  true,
				NullStr:   nil,
				NullInt:   nil,
				NullFlt:   nil,
				NullBl:    nil,
				NullTm:    nil,
				Inner:     innerTestStruct{Name: "nested-struct"},
				InnerPtr:  nil,
				StringPtr: nil,
				Repeated:  []string{"a", "b", "c"},
			},
		})
	})

	t.Run("Map row to map destination", func(t *testing.T) {
		var got map[string]bigquery.Value
		err := bqzimpl.LoadRow(&got, schema, []bigquery.Value{
			"abc",
			int64(25),
			3.14,
			true,
			nil,
			nil,
			nil,
			nil,
			nil,
			[]bigquery.Value{"nested-struct"},
			nil,
			nil,
			[]bigquery.Value{"a", "b", "c"},
		})
		require.NoError(t, err)
		require.Equal(t, "abc", got["ignored_bq_id"])
		require.Equal(t, int64(25), got["age"])
		require.Equal(t, 3.14, got["score"])
		require.Equal(t, true, got["is_active"])
		require.Nil(t, got["null_str"])
		require.Equal(t, []bigquery.Value{"a", "b", "c"}, got["repeated"])
	})
}
