package bqzschema

import (
	"testing"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/stretchr/testify/require"
)

type testSku struct {
	SkuId string  `bigquery:"sku_id"`
	Qtty  int64   `bigquery:"qtty"`
	Price float64 `bigquery:"price"`
}

type testPayload struct {
	Skus    []*testSku `bigquery:"skus"`
	OrderId string     `bigquery:"order_id"`
}

type testRow struct {
	ID        string              `bigquery:"id"`
	Dataset   bigquery.NullString `bigquery:"dataset"`
	Payload   *testPayload        `bigquery:"payload,nullable"`
	CreatedAt time.Time           `bigquery:"created_at"`
}

type testRowCustomNull struct {
	ID        string     `bigquery:"id"`
	Dataset   *string    `bigquery:"dataset"`
	Count     *int64     `bigquery:"count"`
	Rate      *float64   `bigquery:"rate"`
	Approved  *bool      `bigquery:"approved"`
	UpdatedAt *time.Time `bigquery:"updated_at"`
}

type testRowJsonTags struct {
	ID      string `json:"custom_id"`
	Age     int    `json:"custom_age" bigquery:"ignored_bq_name"`
	Ignored string `json:"-"`
}

func TestUnitNativeTableJSON(t *testing.T) {
	type testScenario struct {
		input any
		want  string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		got, err := NativeTableJSON(s.input)
		require.NoError(t, err)
		require.JSONEq(t, s.want, string(got))
		require.Equal(t, byte('\n'), got[len(got)-1])
	}

	t.Run("Struct is sorted recursively", func(t *testing.T) {
		check(t, testScenario{
			input: testRow{},
			want: `[
  {"name": "created_at", "type": "TIMESTAMP", "mode": "REQUIRED"},
  {"name": "dataset", "type": "STRING", "mode": "NULLABLE"},
  {"name": "id", "type": "STRING", "mode": "REQUIRED"},
  {"name": "payload", "type": "RECORD", "mode": "NULLABLE", "fields": [
    {"name": "order_id", "type": "STRING", "mode": "REQUIRED"},
    {"name": "skus", "type": "RECORD", "mode": "REPEATED", "fields": [
      {"name": "price", "type": "FLOAT64", "mode": "REQUIRED"},
      {"name": "qtty", "type": "INTEGER", "mode": "REQUIRED"},
      {"name": "sku_id", "type": "STRING", "mode": "REQUIRED"}
    ]}
  ]}
]`,
		})
	})

	t.Run("Struct with custom basic pointer types", func(t *testing.T) {
		check(t, testScenario{
			input: testRowCustomNull{},
			want: `[
  {"name": "approved", "type": "BOOLEAN", "mode": "NULLABLE"},
  {"name": "count", "type": "INTEGER", "mode": "NULLABLE"},
  {"name": "dataset", "type": "STRING", "mode": "NULLABLE"},
  {"name": "id", "type": "STRING", "mode": "REQUIRED"},
  {"name": "rate", "type": "FLOAT64", "mode": "NULLABLE"},
  {"name": "updated_at", "type": "TIMESTAMP", "mode": "NULLABLE"}
]`,
		})
	})

	t.Run("Struct with BigQuery tags prioritised over JSON tags", func(t *testing.T) {
		check(t, testScenario{
			input: testRowJsonTags{},
			want: `[
  {"name": "custom_id", "type": "STRING", "mode": "REQUIRED"},
  {"name": "ignored_bq_name", "type": "INTEGER", "mode": "REQUIRED"}
]`,
		})
	})

	t.Run("Explicit bigquery.Schema is sorted too", func(t *testing.T) {
		check(t, testScenario{
			input: bigquery.Schema{
				{Name: "b", Type: bigquery.StringFieldType},
				{Name: "a", Type: bigquery.StringFieldType, Required: true},
			},
			want: `[
  {"name": "a", "type": "STRING", "mode": "REQUIRED"},
  {"name": "b", "type": "STRING", "mode": "NULLABLE"}
]`,
		})
	})
}

func TestUnitFixDoesNotMutateInput(t *testing.T) {
	in := bigquery.Schema{
		{Name: "z", Type: bigquery.RecordFieldType, Schema: bigquery.Schema{{Name: "y"}, {Name: "x"}}},
		{Name: "a"},
	}
	out := fix(in)
	require.Equal(t, "a", out[0].Name)
	require.Equal(t, "x", out[1].Schema[0].Name)
	require.Equal(t, "z", in[0].Name)
	require.Equal(t, "y", in[0].Schema[0].Name)
}
