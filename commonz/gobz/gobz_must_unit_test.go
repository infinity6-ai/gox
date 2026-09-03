package gobz

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

type mustTest struct {
	Val int
}

func TestUnitGobzMustFormat(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		val := mustTest{Val: 42}
		var data []byte
		require.NotPanics(t, func() {
			data = MustFormat(val)
		})

		var result mustTest
		_, err := Parse(data, &result)
		require.NoError(t, err)
		require.Equal(t, val, result)
	})

	t.Run("panic on unserializable", func(t *testing.T) {
		require.Panics(t, func() {
			MustFormat(make(chan int))
		})
	})
}

func TestUnitGobzMustParse(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		val := mustTest{Val: 42}
		data := MustFormat(val)

		var result mustTest
		var parsed *mustTest
		require.NotPanics(t, func() {
			parsed = MustParse(data, &result)
		})
		require.Equal(t, val, result)
		require.Equal(t, &result, parsed)
	})

	t.Run("panic on bad data", func(t *testing.T) {
		require.Panics(t, func() {
			MustParse([]byte("bad data"), &mustTest{})
		})
	})
}

func TestUnitGobzMustParseReader(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		val := mustTest{Val: 42}
		data := MustFormat(val)
		reader := bytes.NewReader(data)

		var result mustTest
		var parsed *mustTest
		require.NotPanics(t, func() {
			parsed = MustParseReader(reader, &result)
		})
		require.Equal(t, val, result)
		require.Equal(t, &result, parsed)
	})

	t.Run("panic on bad data", func(t *testing.T) {
		reader := bytes.NewReader([]byte("bad data"))
		require.Panics(t, func() {
			MustParseReader(reader, &mustTest{})
		})
	})
}

func TestUnitGobzMustFormatWriter(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		val := mustTest{Val: 42}
		var buf bytes.Buffer
		require.NotPanics(t, func() {
			MustFormatWriter(&buf, val)
		})

		var result mustTest
		_, err := Parse(buf.Bytes(), &result)
		require.NoError(t, err)
		require.Equal(t, val, result)
	})

	t.Run("panic on unserializable", func(t *testing.T) {
		var buf bytes.Buffer
		require.Panics(t, func() {
			MustFormatWriter(&buf, make(chan int))
		})
	})
}

func TestUnitGobzMustCopy(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		input := &mustTest{Val: 42}
		var output mustTest
		var result *mustTest
		require.NotPanics(t, func() {
			result = MustCopy(input, &output)
		})
		require.Equal(t, *input, output)
		require.Equal(t, &output, result)
		require.NotSame(t, input, &output)
	})

	t.Run("panic on incompatible types", func(t *testing.T) {
		input := &mustTest{Val: 42}
		var output int // Incompatible type
		require.Panics(t, func() {
			MustCopy(input, &output)
		})
	})
}

func TestUnitGobzMustClone(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		input := &mustTest{Val: 42}
		var output mustTest
		var cloned *mustTest
		require.NotPanics(t, func() {
			cloned = MustClone(input, &output)
		})
		require.Equal(t, *input, *cloned)
		require.NotSame(t, input, cloned)
	})

	t.Run("panic on unserializable", func(t *testing.T) {
		require.Panics(t, func() {
			// This will fail during encoding
			MustClone(make(chan int), nil)
		})
	})
}
