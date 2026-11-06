package bitmapz_test

import (
	"testing"

	"github.com/infinity6-ai/gox/commonz/bitmapz"
	"github.com/stretchr/testify/require"
)

type Grant uint64

const (
	read  Grant = 1
	write Grant = 2
	admin Grant = 4
)

func TestUnitBitmapzJoin(t *testing.T) {
	type testScenario struct {
		values []Grant
		want   Grant
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		got := bitmapz.Join(s.values...)
		require.Equal(t, s.want, got)
	}

	t.Run("Empty values returns 0", func(t *testing.T) {
		check(t, testScenario{
			values: nil,
			want:   0,
		})
	})

	t.Run("Single value returns value", func(t *testing.T) {
		check(t, testScenario{
			values: []Grant{read},
			want:   read,
		})
	})

	t.Run("Multiple values returns combined mask", func(t *testing.T) {
		check(t, testScenario{
			values: []Grant{read, admin},
			want:   read | admin, // 5
		})
	})

	t.Run("Duplicate values returns combined mask", func(t *testing.T) {
		check(t, testScenario{
			values: []Grant{write, write},
			want:   write,
		})
	})
}

func TestUnitBitmapzContainsAll(t *testing.T) {
	type testScenario struct {
		current  Grant
		expected []Grant
		want     bool
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		got := bitmapz.ContainsAll(s.current, s.expected...)
		require.Equal(t, s.want, got)
	}

	t.Run("Contains all of subset returns true", func(t *testing.T) {
		check(t, testScenario{
			current:  read | write | admin,
			expected: []Grant{read, write},
			want:     true,
		})
	})

	t.Run("Missing one returns false", func(t *testing.T) {
		check(t, testScenario{
			current:  read | admin,
			expected: []Grant{read, write},
			want:     false,
		})
	})

	t.Run("Empty expected returns true", func(t *testing.T) {
		check(t, testScenario{
			current:  read,
			expected: nil,
			want:     true,
		})
	})

	t.Run("Contains self returns true", func(t *testing.T) {
		check(t, testScenario{
			current:  write,
			expected: []Grant{write},
			want:     true,
		})
	})

	t.Run("Contains composite expected returns true if all bits match", func(t *testing.T) {
		check(t, testScenario{
			current:  read | write | admin,
			expected: []Grant{read | admin},
			want:     true,
		})
	})
}

func TestUnitBitmapzContainsAny(t *testing.T) {
	type testScenario struct {
		current  Grant
		expected []Grant
		want     bool
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		got := bitmapz.ContainsAny(s.current, s.expected...)
		require.Equal(t, s.want, got)
	}

	t.Run("Contains any of overlapping subset returns true", func(t *testing.T) {
		check(t, testScenario{
			current:  read | admin,
			expected: []Grant{read, write},
			want:     true,
		})
	})

	t.Run("No overlap returns false", func(t *testing.T) {
		check(t, testScenario{
			current:  admin,
			expected: []Grant{read, write},
			want:     false,
		})
	})

	t.Run("Empty expected returns false", func(t *testing.T) {
		check(t, testScenario{
			current:  read,
			expected: nil,
			want:     false,
		})
	})

	t.Run("Zero current returns false", func(t *testing.T) {
		check(t, testScenario{
			current:  0,
			expected: []Grant{read},
			want:     false,
		})
	})
}

func TestUnitBitmapzSplit(t *testing.T) {
	type testScenario struct {
		current Grant
		want    []Grant
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		got := bitmapz.Split(s.current)
		require.Equal(t, s.want, got)
	}

	t.Run("Zero returns nil", func(t *testing.T) {
		check(t, testScenario{
			current: 0,
			want:    nil,
		})
	})

	t.Run("Single bit returns slice with single element", func(t *testing.T) {
		check(t, testScenario{
			current: admin,
			want:    []Grant{admin},
		})
	})

	t.Run("Multiple bits returns sorted slice", func(t *testing.T) {
		check(t, testScenario{
			current: admin | read | write,
			want:    []Grant{read, write, admin},
		})
	})
}

func TestUnitBitmapzSet(t *testing.T) {
	type testScenario struct {
		current Grant
		enable  bool
		values  []Grant
		want    Grant
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		got := bitmapz.Set(s.current, s.enable, s.values...)
		require.Equal(t, s.want, got)
	}

	t.Run("Set true adds new bits", func(t *testing.T) {
		check(t, testScenario{
			current: read,
			enable:  true,
			values:  []Grant{write, admin},
			want:    read | write | admin,
		})
	})

	t.Run("Set true on existing bits is no-op", func(t *testing.T) {
		check(t, testScenario{
			current: read | admin,
			enable:  true,
			values:  []Grant{read},
			want:    read | admin,
		})
	})

	t.Run("Set false removes bits", func(t *testing.T) {
		check(t, testScenario{
			current: read | write | admin,
			enable:  false,
			values:  []Grant{read, admin},
			want:    write,
		})
	})

	t.Run("Set false on non-existing bits is no-op", func(t *testing.T) {
		check(t, testScenario{
			current: write,
			enable:  false,
			values:  []Grant{admin},
			want:    write,
		})
	})

	t.Run("Set with empty values is no-op", func(t *testing.T) {
		check(t, testScenario{
			current: read,
			enable:  true,
			values:  nil,
			want:    read,
		})
	})
}

type Uint8Flag uint8

const (
	flagA Uint8Flag = 1 << 0
	flagB Uint8Flag = 1 << 1
	flagC Uint8Flag = 1 << 2
)

func TestUnitBitmapzUint8(t *testing.T) {
	type testScenario struct {
		values []Uint8Flag
		want   Uint8Flag
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		got := bitmapz.Join(s.values...)
		require.Equal(t, s.want, got)
		require.True(t, bitmapz.ContainsAll(got, s.values...))
		require.Equal(t, s.values, bitmapz.Split(got))
	}

	t.Run("Join, ContainsAll, and Split with uint8 type", func(t *testing.T) {
		check(t, testScenario{
			values: []Uint8Flag{flagA, flagC},
			want:   flagA | flagC,
		})
	})
}
