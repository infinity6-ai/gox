package optionalz

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnitNew(t *testing.T) {
	t.Run("string present", func(t *testing.T) {
		got := New("hello", true)
		want := Optional[string]{value: "hello", present: true}
		require.Equal(t, want, got)
	})

	t.Run("int not present", func(t *testing.T) {
		// When creating an optional that is not present, the value is still stored
		// but should not be accessed. The IsPresent() method should be used to check for presence.
		got := New(123, false)
		want := Optional[int]{value: 123, present: false}
		require.Equal(t, want, got)
	})

	t.Run("New with zero value but present", func(t *testing.T) {
		got := New(0, true)
		want := Optional[int]{value: 0, present: true}
		require.Equal(t, want, got)
	})
}

func TestUnitPresent(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		got := Present("world")
		want := Optional[string]{value: "world", present: true}
		require.Equal(t, want, got)
	})

	t.Run("int", func(t *testing.T) {
		got := Present(456)
		want := Optional[int]{value: 456, present: true}
		require.Equal(t, want, got)
	})
}

func TestUnitEmpty(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		got := Empty[string]()
		// The value of an empty optional is the zero value of the type.
		want := Optional[string]{value: "", present: false}
		require.Equal(t, want, got)
	})

	t.Run("int", func(t *testing.T) {
		got := Empty[int]()
		want := Optional[int]{value: 0, present: false}
		require.Equal(t, want, got)
	})
}

func TestUnitIsPresent(t *testing.T) {
	t.Run("present", func(t *testing.T) {
		opt := Present("value")
		require.True(t, opt.IsPresent())
	})

	t.Run("not present", func(t *testing.T) {
		opt := Empty[int]()
		require.False(t, opt.IsPresent())
	})
}

func TestUnitGet(t *testing.T) {
	t.Run("present", func(t *testing.T) {
		opt := Present("hello")
		gotValue, gotPresent := opt.Get()
		require.Equal(t, "hello", gotValue)
		require.True(t, gotPresent)
	})

	t.Run("not present", func(t *testing.T) {
		opt := Empty[int]()
		gotValue, gotPresent := opt.Get()
		require.Equal(t, 0, gotValue) // Zero value for int
		require.False(t, gotPresent)
	})
}

func TestUnitMust(t *testing.T) {
	t.Run("present", func(t *testing.T) {
		opt := Present("data")
		got := opt.Must()
		require.Equal(t, "data", got)
	})

	t.Run("not present panics", func(t *testing.T) {
		opt := Empty[int]()
		require.PanicsWithValue(t, "value not present", func() {
			opt.Must()
		})
	})
}
