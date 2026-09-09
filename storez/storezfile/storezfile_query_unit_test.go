package storezfile_test

import (
	"context"
	"os"
	"testing"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/storez/storez"
	"github.com/infinity6-ai/gox/storez/storezfile"
	"github.com/stretchr/testify/require"
)

func setupTestClient(t *testing.T) (*storezfile.StorezStrategyFile, func()) {
	ctx := context.Background()
	tempDir, err := os.MkdirTemp("", "i6-storez-query-*")
	errorz.Check(err)

	reverter, err := storezfile.I6StorezFileBaseDirEncoded.SetEncoded(ctx, "%s", tempDir)
	errorz.Check(err)

	client := storezfile.Open(ctx, "test-project", "test-db")

	cleanup := func() {
		reverter.Close()
		os.RemoveAll(tempDir)
	}

	return client, cleanup
}

func TestUnitQuery(t *testing.T) {
	client, cleanup := setupTestClient(t)
	defer cleanup()

	ctx := context.Background()
	tableName := "users"

	users := []map[string]*storez.Value{
		{"id": {Value: "1"}, "name": {Value: "Alice"}, "age": {Value: int64(25)}, "city": {Value: "New York"}},
		{"id": {Value: "2"}, "name": {Value: "Bob"}, "age": {Value: int64(30)}, "city": {Value: "London"}},
		{"id": {Value: "3"}, "name": {Value: "Charlie"}, "age": {Value: int64(35)}, "city": {Value: "New York"}},
		{"id": {Value: "4"}, "name": {Value: "David"}, "age": {Value: int64(25)}, "city": {Value: "Paris"}},
	}
	client.PutAll(ctx, tableName, users)

	type testScenario struct {
		name          string
		query         *storez.Query
		expectedIds   []string
		expectedCursor string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		cursor, result := client.Query(ctx, s.query)
		ids := []string{}
		for _, row := range result {
			ids = append(ids, row["id"].Value.(string))
		}
		require.Equal(t, s.expectedIds, ids, "for query %v", s.query)
		if s.expectedCursor != "" {
			require.Equal(t, s.expectedCursor, cursor, "for query %v", s.query)
		}
	}

	t.Run("Simple filter", func(t *testing.T) {
		check(t, testScenario{
			name: "Simple filter",
			query: &storez.Query{
				Table: tableName,
				Filters: []storez.Filter{
					{Field: "city", Op: "=", Value: "New York"},
				},
				Limit: 10,
			},
			expectedIds: []string{"1", "3"},
		})
	})

	t.Run("Filter with ordering", func(t *testing.T) {
		check(t, testScenario{
			name: "Filter with ordering",
			query: &storez.Query{
				Table: tableName,
				Filters: []storez.Filter{
					{Field: "city", Op: "=", Value: "New York"},
				},
				OrderBys: []storez.OrderBy{
					{Field: "age", Desc: true},
				},
				Limit: 10,
			},
			expectedIds: []string{"3", "1"},
		})
	})

	t.Run("Filter with multiple conditions", func(t *testing.T) {
		check(t, testScenario{
			name: "Filter with multiple conditions",
			query: &storez.Query{
				Table: tableName,
				Filters: []storez.Filter{
					{Field: "age", Op: "=", Value: int64(25)},
					{Field: "city", Op: "=", Value: "New York"},
				},
				Limit: 10,
			},
			expectedIds: []string{"1"},
		})
	})

	t.Run("Greater than filter", func(t *testing.T) {
		check(t, testScenario{
			name: "Greater than filter",
			query: &storez.Query{
				Table: tableName,
				Filters: []storez.Filter{
					{Field: "age", Op: ">", Value: int64(30)},
				},
				Limit: 10,
			},
			expectedIds: []string{"3"},
		})
	})

	t.Run("Pagination with limit", func(t *testing.T) {
		query := &storez.Query{
			Table: tableName,
			OrderBys: []storez.OrderBy{
				{Field: "name"},
			},
			Limit: 2,
		}
		cursor, result := client.Query(ctx, query)
		ids := []string{}
		for _, row := range result {
			ids = append(ids, row["id"].Value.(string))
		}
		require.Equal(t, []string{"1", "2"}, ids)
		require.NotEmpty(t, cursor)

		// Get next page
		query.StartCursor = cursor
		cursor, result = client.Query(ctx, query)
		ids = []string{}
		for _, row := range result {
			ids = append(ids, row["id"].Value.(string))
		}
		require.Equal(t, []string{"3", "4"}, ids)
		require.NotEmpty(t, cursor)
		
		// Get last page (should be empty)
		query.StartCursor = cursor
		_, result = client.Query(ctx, query)
		require.Empty(t, result)
	})
}
