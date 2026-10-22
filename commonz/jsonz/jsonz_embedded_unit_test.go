package jsonz

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

type embeddedTestPayload struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type embeddedTestOuter struct {
	ID      string                       `json:"id"`
	Payload EmbeddedJson[embeddedTestPayload] `json:"payload"`
}

func TestUnitEmbeddedJsonMarshal(t *testing.T) {
	type testScenario struct {
		input   any
		want    string
		wantErr string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		got, err := json.Marshal(s.input)
		if s.wantErr != "" {
			require.Error(t, err)
			require.Contains(t, err.Error(), s.wantErr)
			return
		}
		require.NoError(t, err)
		require.Equal(t, s.want, string(got))
	}

	t.Run("Struct payload marshaled as escaped JSON string", func(t *testing.T) {
		check(t, testScenario{
			input: NewEmbedded(embeddedTestPayload{Name: "Alice", Count: 42}),
			want:  `"{\"name\":\"Alice\",\"count\":42}"`,
		})
	})

	t.Run("Marshal via value receiver", func(t *testing.T) {
		val := NewEmbedded("hello")
		check(t, testScenario{
			input: val,
			want:  `"\"hello\""`,
		})
	})

	t.Run("Marshal via pointer", func(t *testing.T) {
		val := NewEmbedded("hello")
		check(t, testScenario{
			input: &val,
			want:  `"\"hello\""`,
		})
	})

	t.Run("Nested in outer struct", func(t *testing.T) {
		outer := embeddedTestOuter{
			ID:      "item-1",
			Payload: NewEmbedded(embeddedTestPayload{Name: "Bob", Count: 10}),
		}
		check(t, testScenario{
			input: outer,
			want:  `{"id":"item-1","payload":"{\"name\":\"Bob\",\"count\":10}"}`,
		})
	})

	t.Run("Primitive int value", func(t *testing.T) {
		check(t, testScenario{
			input: NewEmbedded(123),
			want:  `"123"`,
		})
	})

	t.Run("Nil pointer payload marshals as null string", func(t *testing.T) {
		check(t, testScenario{
			input: NewEmbedded[*embeddedTestPayload](nil),
			want:  `"null"`,
		})
	})

	t.Run("Unsupported type returns error", func(t *testing.T) {
		check(t, testScenario{
			input:   NewEmbedded(make(chan int)),
			wantErr: "cannot marshal EmbeddedJson",
		})
	})
}

func TestUnitEmbeddedJsonUnmarshal(t *testing.T) {
	type testScenario struct {
		jsonInput string
		want      embeddedTestPayload
		wantErr   string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		var got EmbeddedJson[embeddedTestPayload]
		err := json.Unmarshal([]byte(s.jsonInput), &got)
		if s.wantErr != "" {
			require.Error(t, err)
			require.Contains(t, err.Error(), s.wantErr)
			return
		}
		require.NoError(t, err)
		require.Equal(t, s.want, got.Get())
	}

	t.Run("Valid escaped JSON string", func(t *testing.T) {
		check(t, testScenario{
			jsonInput: `"{\"name\":\"Alice\",\"count\":42}"`,
			want:      embeddedTestPayload{Name: "Alice", Count: 42},
		})
	})

	t.Run("JSON null resets to zero value", func(t *testing.T) {
		check(t, testScenario{
			jsonInput: `null`,
			want:      embeddedTestPayload{},
		})
	})

	t.Run("Empty JSON string resets to zero value", func(t *testing.T) {
		check(t, testScenario{
			jsonInput: `""`,
			want:      embeddedTestPayload{},
		})
	})

	t.Run("Invalid outer JSON non-string fails", func(t *testing.T) {
		check(t, testScenario{
			jsonInput: `{"name":"Alice"}`,
			wantErr:   "cannot unmarshal EmbeddedJson",
		})
	})

	t.Run("Invalid inner JSON string fails", func(t *testing.T) {
		check(t, testScenario{
			jsonInput: `"{invalid json"`,
			wantErr:   "cannot unmarshal inner EmbeddedJson",
		})
	})

	t.Run("Unmarshal into outer struct", func(t *testing.T) {
		raw := `{"id":"item-2","payload":"{\"name\":\"Carol\",\"count\":99}"}`
		var outer embeddedTestOuter
		err := json.Unmarshal([]byte(raw), &outer)
		require.NoError(t, err)
		require.Equal(t, "item-2", outer.ID)
		require.Equal(t, embeddedTestPayload{Name: "Carol", Count: 99}, outer.Payload.Get())
	})

	t.Run("Unmarshal with nil receiver returns error", func(t *testing.T) {
		var nilPtr *EmbeddedJson[embeddedTestPayload]
		err := nilPtr.UnmarshalJSON([]byte(`"{\"name\":\"Alice\"}"`))
		require.Error(t, err)
		require.Contains(t, err.Error(), "cannot unmarshal into nil *EmbeddedJson")
	})
}

func TestUnitEmbeddedJsonAccessors(t *testing.T) {
	t.Run("NewEmbedded and Get", func(t *testing.T) {
		v := NewEmbedded(42)
		require.Equal(t, 42, v.Get())
	})

	t.Run("NewEmbeddedJson constructor equivalence", func(t *testing.T) {
		v := NewEmbeddedJson("hello")
		require.Equal(t, "hello", v.Get())
	})

	t.Run("Set updates underlying value", func(t *testing.T) {
		v := NewEmbedded("initial")
		v.Set("updated")
		require.Equal(t, "updated", v.Get())
	})

	t.Run("Value returns mutable pointer", func(t *testing.T) {
		v := NewEmbedded(embeddedTestPayload{Name: "Initial", Count: 1})
		ptr := v.Value()
		ptr.Name = "Mutated"
		ptr.Count = 100
		require.Equal(t, "Mutated", v.Get().Name)
		require.Equal(t, 100, v.Get().Count)
	})

	t.Run("Type alias Json[T] compatibility", func(t *testing.T) {
		var alias Json[int] = NewEmbedded(10)
		require.Equal(t, 10, alias.Get())
	})
}
