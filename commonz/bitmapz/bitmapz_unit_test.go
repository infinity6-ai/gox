package bitmapz

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

type Perm struct {
	Value uint64
	Name  string
}

// Reset registry for tests to avoid state bleed
func resetRegistry() {
	registry = make(map[reflect.Type]any)
}

func TestUnitBitmapExpand(t *testing.T) {
	resetRegistry()
	Read := Perm{Value: 1, Name: "Read"}
	Write := Perm{Value: 2, Name: "Write"}
	Execute := Perm{Value: 4, Name: "Execute"}
	None := Perm{Value: 0, Name: "None"}

	Register(Read, Write, Execute, None)

	type testScenario struct {
		mask uint64
		want []Perm
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		got := Expand[Perm](s.mask)
		// We expect the result to match 'want' exactly in length and content.
		require.Equal(t, len(s.want), len(got), "length mismatch")
		for _, w := range s.want {
			require.Contains(t, got, w)
		}
	}

	t.Run("Empty mask returns None", func(t *testing.T) {
		check(t, testScenario{
			mask: 0,
			want: []Perm{None},
		})
	})

	t.Run("ReadWrite mask returns Read and Write without None", func(t *testing.T) {
		check(t, testScenario{
			mask: 3,
			want: []Perm{Read, Write},
		})
	})

	t.Run("Full mask returns all except None", func(t *testing.T) {
		check(t, testScenario{
			mask: 7,
			want: []Perm{Read, Write, Execute},
		})
	})

	t.Run("Unregistered mask bits are ignored", func(t *testing.T) {
		check(t, testScenario{
			mask: 11, // 8 (unregistered) + 2 (Write) + 1 (Read)
			want: []Perm{Read, Write},
		})
	})
}

func TestUnitBitmapMultipleRegister(t *testing.T) {
	resetRegistry()
	Read := Perm{Value: 1, Name: "Read"}
	Write := Perm{Value: 2, Name: "Write"}
	
	// Register twice to cover the 'existing' branch in Register
	Register(Read)
	Register(Write)
	
	got := Expand[Perm](uint64(3))
	require.Len(t, got, 2)
	require.Contains(t, got, Read)
	require.Contains(t, got, Write)
}

type Role struct {
	Value uint64
	Name  string
}

func TestUnitBitmapUnregisteredType(t *testing.T) {
	resetRegistry()
	t.Run("Expand on unregistered type returns nil", func(t *testing.T) {
		got := Expand[Role](uint64(3))
		require.Nil(t, got)
	})
}
