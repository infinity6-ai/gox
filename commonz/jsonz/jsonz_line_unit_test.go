package jsonz

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type lineTestItem struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func TestUnitLineParseReader(t *testing.T) {
	type testScenario struct {
		input   string
		want    []lineTestItem
		wantErr string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		var got []lineTestItem
		reader := strings.NewReader(s.input)
		err := LineParseReader(reader, &got)

		if s.wantErr != "" {
			require.Error(t, err)
			require.Contains(t, err.Error(), s.wantErr)
		} else {
			require.NoError(t, err)
			require.Equal(t, s.want, got)
		}
	}

	t.Run("Valid JSON lines", func(t *testing.T) {
		check(t, testScenario{
			input: "{\"id\":1,\"name\":\"A\"}\n{\"id\":2,\"name\":\"B\"}\n",
			want: []lineTestItem{
				{ID: 1, Name: "A"},
				{ID: 2, Name: "B"},
			},
		})
	})

	t.Run("Empty input", func(t *testing.T) {
		check(t, testScenario{
			input: "",
			want:  nil,
		})
	})

	t.Run("Blank lines ignored", func(t *testing.T) {
		check(t, testScenario{
			input: "\n{\"id\":1,\"name\":\"A\"}\n\n{\"id\":2,\"name\":\"B\"}\n\n",
			want: []lineTestItem{
				{ID: 1, Name: "A"},
				{ID: 2, Name: "B"},
			},
		})
	})

	t.Run("Invalid line JSON", func(t *testing.T) {
		check(t, testScenario{
			input:   "{\"id\":1,\"name\":\"A\"}\n{invalid}\n",
			wantErr: "failed to parse line",
		})
	})

	t.Run("Nil destination pointer", func(t *testing.T) {
		err := LineParseReader[[]lineTestItem](strings.NewReader(""), nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "destination slice pointer cannot be nil")
	})
}

func TestUnitLineParseInto(t *testing.T) {
	t.Run("Parse into from string", func(t *testing.T) {
		input := "{\"id\":1,\"name\":\"A\"}\n{\"id\":2,\"name\":\"B\"}"
		var got []lineTestItem
		err := LineParseInto(input, &got)
		require.NoError(t, err)
		require.Equal(t, []lineTestItem{
			{ID: 1, Name: "A"},
			{ID: 2, Name: "B"},
		}, got)
	})

	t.Run("Parse into with error", func(t *testing.T) {
		input := "invalid json line"
		var got []lineTestItem
		err := LineParseInto(input, &got)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to parse into slice")
	})
}

func TestUnitLineFormatWriter(t *testing.T) {
	t.Run("Valid slice of structs", func(t *testing.T) {
		input := []lineTestItem{
			{ID: 1, Name: "A"},
			{ID: 2, Name: "B"},
		}
		var buf bytes.Buffer
		err := LineFormatWriter(&buf, input)
		require.NoError(t, err)
		require.Equal(t, "{\"id\":1,\"name\":\"A\"}\n{\"id\":2,\"name\":\"B\"}\n", buf.String())
	})

	t.Run("Empty slice", func(t *testing.T) {
		var input []lineTestItem
		var buf bytes.Buffer
		err := LineFormatWriter(&buf, input)
		require.NoError(t, err)
		require.Empty(t, buf.String())
	})

	t.Run("Encoder error on unsupported type", func(t *testing.T) {
		type badItem struct {
			Ch chan int
		}
		input := []badItem{{Ch: make(chan int)}}
		var buf bytes.Buffer
		err := LineFormatWriter(&buf, input)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to encode item at index 0")
	})
}

func TestUnitLineFormatReadCloser(t *testing.T) {
	t.Run("Successful stream", func(t *testing.T) {
		input := []lineTestItem{
			{ID: 1, Name: "A"},
		}
		rc := LineFormatReadCloser(input)
		defer rc.Close()

		data, err := io.ReadAll(rc)
		require.NoError(t, err)
		require.Equal(t, "{\"id\":1,\"name\":\"A\"}\n", string(data))
	})

	t.Run("Encoder error inside pipe", func(t *testing.T) {
		type badItem struct {
			Ch chan int
		}
		input := []badItem{{Ch: make(chan int)}}
		rc := LineFormatReadCloser(input)
		defer rc.Close()

		_, err := io.ReadAll(rc)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to encode item at index 0")
	})
}

func TestUnitLineFormatReader(t *testing.T) {
	t.Run("Success path", func(t *testing.T) {
		input := []lineTestItem{
			{ID: 1, Name: "A"},
		}
		var output string
		err := LineFormatReader(input, func(r io.Reader) error {
			data, readErr := io.ReadAll(r)
			if readErr != nil {
				return readErr
			}
			output = string(data)
			return nil
		})
		require.NoError(t, err)
		require.Equal(t, "{\"id\":1,\"name\":\"A\"}\n", output)
	})

	t.Run("Callback failure", func(t *testing.T) {
		input := []lineTestItem{{ID: 1, Name: "A"}}
		err := LineFormatReader(input, func(r io.Reader) error {
			return fmt.Errorf("simulated error")
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "callback failed")
		require.Contains(t, err.Error(), "simulated error")
	})
}

func TestUnitLineFormatAndBytes(t *testing.T) {
	t.Run("LineFormat success", func(t *testing.T) {
		input := []lineTestItem{
			{ID: 1, Name: "A"},
		}
		blob, err := LineFormat(input)
		require.NoError(t, err)
		require.Equal(t, "{\"id\":1,\"name\":\"A\"}\n", blob.String())
	})

	t.Run("LineFormat error", func(t *testing.T) {
		type badItem struct {
			Ch chan int
		}
		input := []badItem{{Ch: make(chan int)}}
		_, err := LineFormat(input)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to format slice")
	})

	t.Run("LineFormatBytes success", func(t *testing.T) {
		input := []lineTestItem{
			{ID: 1, Name: "A"},
		}
		bytesData, err := LineFormatBytes(input)
		require.NoError(t, err)
		require.Equal(t, []byte("{\"id\":1,\"name\":\"A\"}\n"), bytesData)
	})

	t.Run("LineFormatBytes error", func(t *testing.T) {
		type badItem struct {
			Ch chan int
		}
		input := []badItem{{Ch: make(chan int)}}
		_, err := LineFormatBytes(input)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to format bytes")
	})
}
