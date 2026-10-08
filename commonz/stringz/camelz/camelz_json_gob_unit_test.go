package camelz_test

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/infinity6-ai/gox/commonz/stringz/camelz"
)

func TestUnitJson(t *testing.T) {
	type testScenario struct {
		input       *camelz.Parsed
		expectedRaw string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		data, err := json.Marshal(s.input)
		require.NoError(t, err)
		require.Equal(t, s.expectedRaw, string(data))

		var unmarshaled *camelz.Parsed
		err = json.Unmarshal(data, &unmarshaled)
		require.NoError(t, err)

		if s.input == nil {
			require.Nil(t, unmarshaled)
		} else {
			require.NotNil(t, unmarshaled)
			require.Equal(t, s.input.Parts(), unmarshaled.Parts())
			require.Equal(t, s.input.SL(), unmarshaled.SL())
			require.Equal(t, s.input.P(), unmarshaled.P())
		}
	}

	t.Run("standard snake lower", func(t *testing.T) {
		check(t, testScenario{
			input:       camelz.P("foo_bar"),
			expectedRaw: `"foo_bar"`,
		})
	})

	t.Run("from PascalCase serialized as snake lower", func(t *testing.T) {
		check(t, testScenario{
			input:       camelz.P("FooBarBaz"),
			expectedRaw: `"foo_bar_baz"`,
		})
	})

	t.Run("mixed case serialized as snake lower", func(t *testing.T) {
		check(t, testScenario{
			input:       camelz.P("a_b-c-pUi"),
			expectedRaw: `"a_b_c_p_ui"`,
		})
	})

	t.Run("empty string parsed", func(t *testing.T) {
		check(t, testScenario{
			input:       camelz.P(""),
			expectedRaw: `""`,
		})
	})

	t.Run("nil parsed", func(t *testing.T) {
		check(t, testScenario{
			input:       nil,
			expectedRaw: "null",
		})
	})

	t.Run("unmarshal invalid json error", func(t *testing.T) {
		var p camelz.Parsed
		err := json.Unmarshal([]byte(`{"invalid": true}`), &p)
		require.Error(t, err)
		require.Contains(t, err.Error(), "cannot unmarshal camelz from json")
	})

	t.Run("unmarshal unsupported content error", func(t *testing.T) {
		var p camelz.Parsed
		err := json.Unmarshal([]byte(`"foo bar"`), &p)
		require.Error(t, err)
		require.Contains(t, err.Error(), "cannot parse camelz from json")
	})
}

func TestUnitGob(t *testing.T) {
	type testScenario struct {
		input *camelz.Parsed
	}

	checkDirect := func(t *testing.T, s testScenario) {
		t.Helper()
		encoded, err := s.input.GobEncode()
		require.NoError(t, err)
		require.NotNil(t, encoded)

		var decoded camelz.Parsed
		err = decoded.GobDecode(encoded)
		require.NoError(t, err)

		if s.input == nil {
			require.Nil(t, decoded.Parts())
		} else {
			require.Equal(t, s.input.Parts(), decoded.Parts())
			require.Equal(t, s.input.SL(), decoded.SL())
			require.Equal(t, s.input.P(), decoded.P())
		}
	}

	t.Run("standard identifier direct encode decode", func(t *testing.T) {
		checkDirect(t, testScenario{
			input: camelz.P("foo_bar"),
		})
	})

	t.Run("from PascalCase direct encode decode", func(t *testing.T) {
		checkDirect(t, testScenario{
			input: camelz.P("FooBarBaz"),
		})
	})

	t.Run("from mixed casing a_b-c-pUi direct encode decode", func(t *testing.T) {
		checkDirect(t, testScenario{
			input: camelz.P("a_b-c-pUi"),
		})
	})

	t.Run("from empty parsed direct encode decode", func(t *testing.T) {
		checkDirect(t, testScenario{
			input: camelz.P(""),
		})
	})

	t.Run("nil parsed direct encode decode", func(t *testing.T) {
		checkDirect(t, testScenario{
			input: nil,
		})
	})

	type gobContainer struct {
		Name *camelz.Parsed
	}

	t.Run("container struct with value via standard gob encoder", func(t *testing.T) {
		c := gobContainer{Name: camelz.P("hello_world")}
		var buf bytes.Buffer
		err := gob.NewEncoder(&buf).Encode(c)
		require.NoError(t, err)

		var decoded gobContainer
		err = gob.NewDecoder(&buf).Decode(&decoded)
		require.NoError(t, err)
		require.NotNil(t, decoded.Name)
		require.Equal(t, []string{"hello", "world"}, decoded.Name.Parts())
		require.Equal(t, "HelloWorld", decoded.Name.P())
	})

	t.Run("container struct with nil via standard gob encoder", func(t *testing.T) {
		c := gobContainer{Name: nil}
		var buf bytes.Buffer
		err := gob.NewEncoder(&buf).Encode(c)
		require.NoError(t, err)

		var decoded gobContainer
		err = gob.NewDecoder(&buf).Decode(&decoded)
		require.NoError(t, err)
		require.Nil(t, decoded.Name)
	})

	t.Run("gob decode invalid data", func(t *testing.T) {
		var p camelz.Parsed
		err := p.GobDecode([]byte{0x01, 0x02, 0x03})
		require.Error(t, err)
		require.Contains(t, err.Error(), "cannot unmarshal camelz from gob")
	})

	t.Run("gob decode unsupported content error", func(t *testing.T) {
		var buf bytes.Buffer
		err := gob.NewEncoder(&buf).Encode("foo bar")
		require.NoError(t, err)

		var p camelz.Parsed
		err = p.GobDecode(buf.Bytes())
		require.Error(t, err)
		require.Contains(t, err.Error(), "cannot parse camelz from gob")
	})
}
