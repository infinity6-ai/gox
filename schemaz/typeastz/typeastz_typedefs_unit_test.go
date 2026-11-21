package typeastz_test

import (
	"encoding/json"
	"testing"

	"github.com/infinity6-ai/gox/schemaz/typeastz"
	"github.com/stretchr/testify/require"
)

func TestUnitTuples(t *testing.T) {
	type testScenario struct {
		name       string
		tuples     []typeastz.Tuple
		jsonInput  string
		wantTuples []typeastz.Tuple
		wantJSON   string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()

		if s.tuples != nil {
			tups := typeastz.NewTuples(s.tuples)
			require.Equal(t, s.wantTuples, tups.All())

			data, err := json.Marshal(tups)
			require.NoError(t, err)
			if s.wantJSON != "" {
				require.JSONEq(t, s.wantJSON, string(data))
			}

			var unmarshaled typeastz.Tuples
			err = json.Unmarshal(data, &unmarshaled)
			require.NoError(t, err)
			require.Equal(t, s.wantTuples, unmarshaled.All())
		}

		if s.jsonInput != "" {
			var unmarshaled typeastz.Tuples
			err := json.Unmarshal([]byte(s.jsonInput), &unmarshaled)
			require.NoError(t, err)
			require.Equal(t, s.wantTuples, unmarshaled.All())
		}
	}

	t.Run("tuples preserves fields through creation and serialization", func(t *testing.T) {
		check(t, testScenario{
			tuples: []typeastz.Tuple{
				{Typ: typeastz.TypString, Name: "id", Summary: "Identifier"},
				{Typ: typeastz.TypNumber, Name: "count", Summary: "Counter"},
			},
			wantTuples: []typeastz.Tuple{
				{Typ: typeastz.TypString, Name: "id", Summary: "Identifier"},
				{Typ: typeastz.TypNumber, Name: "count", Summary: "Counter"},
			},
			wantJSON: `[
				{"typ":"string","name":"id","summary":"Identifier"},
				{"typ":"number","name":"count","summary":"Counter"}
			]`,
		})
	})

	t.Run("tuples unmarshal from json", func(t *testing.T) {
		check(t, testScenario{
			jsonInput: `[
				{"typ":"object","name":"file","summary":"File stat"}
			]`,
			wantTuples: []typeastz.Tuple{
				{Typ: "object", Name: "file", Summary: "File stat"},
			},
		})
	})

	t.Run("tuples skip identical duplicate tuple", func(t *testing.T) {
		tups := typeastz.NewTuples([]typeastz.Tuple{
			{Typ: typeastz.TypString, Name: "foo", Summary: "summary"},
		})
		tups.Add(typeastz.Tuple{Typ: typeastz.TypString, Name: "foo", Summary: "summary"})
		require.Len(t, tups.All(), 1)
	})

	t.Run("tuples panic on collision with different tuple fields", func(t *testing.T) {
		tups := typeastz.NewTuples([]typeastz.Tuple{
			{Typ: typeastz.TypNumber, Name: "field", Summary: "desc"},
		})
		require.Panics(t, func() {
			tups.Add(typeastz.Tuple{Typ: typeastz.TypString, Name: "field", Summary: "desc"})
		})
	})
}

func TestUnitTypedefs(t *testing.T) {
	t.Run("typedef in Head and Properties", func(t *testing.T) {
		td := &typeastz.Typedef{
			Head: &typeastz.Tuple{
				Typ:     typeastz.TypObject,
				Name:    "MyStruct",
				Summary: "Sample struct",
			},
			Properties: typeastz.NewTuples([]typeastz.Tuple{
				{Typ: typeastz.TypString, Name: "name", Summary: "Name"},
				{Typ: typeastz.TypNumber, Name: "ratio", Summary: "Ratio"},
			}),
		}

		tds := typeastz.NewTypedefs([]*typeastz.Typedef{td})
		all := tds.All()
		require.Len(t, all, 1)
		require.Equal(t, "MyStruct", all[0].Head.Name)

		props := all[0].Properties.All()
		require.Len(t, props, 2)
		require.Equal(t, "string", props[0].Typ)
		require.Equal(t, "number", props[1].Typ)

		data, err := json.Marshal(&tds)
		require.NoError(t, err)

		var unmarshaled typeastz.Typedefs
		err = json.Unmarshal(data, &unmarshaled)
		require.NoError(t, err)

		unmarshaledAll := unmarshaled.All()
		require.Len(t, unmarshaledAll, 1)
		require.Equal(t, "MyStruct", unmarshaledAll[0].Head.Name)
		unmarshaledProps := unmarshaledAll[0].Properties.All()
		require.Len(t, unmarshaledProps, 2)
		require.Equal(t, "string", unmarshaledProps[0].Typ)
		require.Equal(t, "number", unmarshaledProps[1].Typ)
	})

	t.Run("typedefs panic on collision with different tuple fields", func(t *testing.T) {
		td1 := &typeastz.Typedef{
			Head: &typeastz.Tuple{Typ: typeastz.TypObject, Name: "Item", Summary: "Summary A"},
		}
		td2 := &typeastz.Typedef{
			Head: &typeastz.Tuple{Typ: typeastz.TypObject, Name: "Item", Summary: "Summary B"},
		}
		tds := typeastz.NewTypedefs([]*typeastz.Typedef{td1})
		require.Panics(t, func() {
			tds.Add(td2)
		})
	})
}
