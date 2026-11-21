package schemamold_test

import (
	"testing"

	"github.com/infinity6-ai/gox/schemaz/schemamold"
	"github.com/infinity6-ai/gox/schemaz/typeastz"
	"github.com/stretchr/testify/require"
)

func TestUnitFormatTypedefs(t *testing.T) {
	type testScenario struct {
		name string
		in   typeastz.Typedefs
		want string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		got := schemamold.FormatTypedefs(&s.in)
		require.Equal(t, s.want, got)
	}

	t.Run("basic object", func(t *testing.T) {
		var tds typeastz.Typedefs
		td := &typeastz.Typedef{
			Head:       &typeastz.Tuple{Typ: typeastz.TypObject, Name: "MyType", Summary: "My summary"},
			Properties: &typeastz.Tuples{},
		}
		td.Properties.Add(
			typeastz.Tuple{Typ: typeastz.TypString, Name: "foo"},
			typeastz.Tuple{Typ: typeastz.TypNumber, Name: "bar", Summary: "bar summary"},
		)
		tds.Add(td)

		check(t, testScenario{
			name: "basic object",
			in:   tds,
			want: "/**\n * My summary\n * @typedef {object} MyType\n * @property {string} foo\n * @property {number} bar - bar summary\n */\n\n",
		})
	})

	t.Run("multiple typedefs", func(t *testing.T) {
		var tds typeastz.Typedefs
		td1 := &typeastz.Typedef{
			Head: &typeastz.Tuple{Typ: typeastz.TypString, Name: "AliasType"},
		}
		td2 := &typeastz.Typedef{
			Head:       &typeastz.Tuple{Typ: typeastz.TypObject, Name: "BType"},
			Properties: &typeastz.Tuples{},
		}
		td2.Properties.Add(typeastz.Tuple{Typ: typeastz.TypNumber, Name: "x"})

		tds.Add(td1, td2)

		check(t, testScenario{
			name: "multiple typedefs",
			in:   tds,
			want: "/**\n * @typedef {string} AliasType\n */\n\n/**\n * @typedef {object} BType\n * @property {number} x\n */\n\n",
		})
	})

	t.Run("various jstypes", func(t *testing.T) {
		var tds typeastz.Typedefs
		td := &typeastz.Typedef{
			Head:       &typeastz.Tuple{Typ: typeastz.TypObject, Name: "ComplexType"},
			Properties: &typeastz.Tuples{},
		}
		td.Properties.Add(
			typeastz.Tuple{Typ: typeastz.TypBoolean, Name: "flag"},
			typeastz.Tuple{Typ: "Array<boolean>", Name: "flags"},
			typeastz.Tuple{Typ: "Record<string, object>", Name: "nested"},
		)
		tds.Add(td)

		check(t, testScenario{
			name: "various jstypes",
			in:   tds,
			want: "/**\n * @typedef {object} ComplexType\n * @property {boolean} flag\n * @property {Array<boolean>} flags\n * @property {Record<string, object>} nested\n */\n\n",
		})
	})

	t.Run("typedef and properties with gotyp comments", func(t *testing.T) {
		var tds typeastz.Typedefs
		td1 := &typeastz.Typedef{
			Head: &typeastz.Tuple{Typ: typeastz.TypString, GoTyp: "time.Time", Name: "dszTime", Summary: "Timestamp"},
		}
		td2 := &typeastz.Typedef{
			Head:       &typeastz.Tuple{Typ: typeastz.TypObject, GoTyp: "models.User", Name: "UserType"},
			Properties: &typeastz.Tuples{},
		}
		td2.Properties.Add(
			typeastz.Tuple{Typ: typeastz.TypString, GoTyp: "string", Name: "id"},
			typeastz.Tuple{Typ: typeastz.TypNumber, GoTyp: "int64", Name: "age", Summary: "User age"},
		)
		tds.Add(td1, td2)

		check(t, testScenario{
			name: "typedef and properties with gotyp comments",
			in:   tds,
			want: "/**\n * [Go: models.User]\n * @typedef {object} UserType\n * @property {string} id - [Go: string]\n * @property {number} age - [Go: int64] User age\n */\n\n/**\n * [Go: time.Time] Timestamp\n * @typedef {string} dszTime\n */\n\n",
		})
	})
}
