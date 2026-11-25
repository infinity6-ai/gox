package dataz_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/infinity6-ai/gox/commonz/dataz"
	"github.com/stretchr/testify/require"
)

func TestUnitNullStringJSON(t *testing.T) {
	type testScenario struct {
		val       dataz.NullString
		wantJSON  string
		shouldErr bool
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		b, err := json.Marshal(s.val)
		if s.shouldErr {
			require.Error(t, err)
			return
		}
		require.NoError(t, err)
		require.JSONEq(t, s.wantJSON, string(b))

		var got dataz.NullString
		err = json.Unmarshal([]byte(s.wantJSON), &got)
		require.NoError(t, err)
		require.Equal(t, s.val, got)
	}

	t.Run("Valid NullString", func(t *testing.T) {
		check(t, testScenario{
			val:      dataz.NullString{StringVal: "hello", Valid: true},
			wantJSON: `"hello"`,
		})
	})

	t.Run("Invalid NullString", func(t *testing.T) {
		check(t, testScenario{
			val:      dataz.NullString{Valid: false},
			wantJSON: `null`,
		})
	})
}

func TestUnitNullInt64JSON(t *testing.T) {
	type testScenario struct {
		val       dataz.NullInt64
		wantJSON  string
		shouldErr bool
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		b, err := json.Marshal(s.val)
		if s.shouldErr {
			require.Error(t, err)
			return
		}
		require.NoError(t, err)
		require.JSONEq(t, s.wantJSON, string(b))

		var got dataz.NullInt64
		err = json.Unmarshal([]byte(s.wantJSON), &got)
		require.NoError(t, err)
		require.Equal(t, s.val, got)
	}

	t.Run("Valid NullInt64", func(t *testing.T) {
		check(t, testScenario{
			val:      dataz.NullInt64{Int64: 42, Valid: true},
			wantJSON: `42`,
		})
	})

	t.Run("Invalid NullInt64", func(t *testing.T) {
		check(t, testScenario{
			val:      dataz.NullInt64{Valid: false},
			wantJSON: `null`,
		})
	})
}

func TestUnitNullFloat64JSON(t *testing.T) {
	type testScenario struct {
		val       dataz.NullFloat64
		wantJSON  string
		shouldErr bool
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		b, err := json.Marshal(s.val)
		if s.shouldErr {
			require.Error(t, err)
			return
		}
		require.NoError(t, err)
		require.JSONEq(t, s.wantJSON, string(b))

		var got dataz.NullFloat64
		err = json.Unmarshal([]byte(s.wantJSON), &got)
		require.NoError(t, err)
		require.Equal(t, s.val, got)
	}

	t.Run("Valid NullFloat64", func(t *testing.T) {
		check(t, testScenario{
			val:      dataz.NullFloat64{Float64: 3.14, Valid: true},
			wantJSON: `3.14`,
		})
	})

	t.Run("Invalid NullFloat64", func(t *testing.T) {
		check(t, testScenario{
			val:      dataz.NullFloat64{Valid: false},
			wantJSON: `null`,
		})
	})
}

func TestUnitNullBoolJSON(t *testing.T) {
	type testScenario struct {
		val       dataz.NullBool
		wantJSON  string
		shouldErr bool
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		b, err := json.Marshal(s.val)
		if s.shouldErr {
			require.Error(t, err)
			return
		}
		require.NoError(t, err)
		require.JSONEq(t, s.wantJSON, string(b))

		var got dataz.NullBool
		err = json.Unmarshal([]byte(s.wantJSON), &got)
		require.NoError(t, err)
		require.Equal(t, s.val, got)
	}

	t.Run("Valid NullBool True", func(t *testing.T) {
		check(t, testScenario{
			val:      dataz.NullBool{Bool: true, Valid: true},
			wantJSON: `true`,
		})
	})

	t.Run("Valid NullBool False", func(t *testing.T) {
		check(t, testScenario{
			val:      dataz.NullBool{Bool: false, Valid: true},
			wantJSON: `false`,
		})
	})

	t.Run("Invalid NullBool", func(t *testing.T) {
		check(t, testScenario{
			val:      dataz.NullBool{Valid: false},
			wantJSON: `null`,
		})
	})
}

func TestUnitNullTimestampJSON(t *testing.T) {
	type testScenario struct {
		val       dataz.NullTimestamp
		wantJSON  string
		shouldErr bool
	}

	now := time.Now().UTC()

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		b, err := json.Marshal(s.val)
		if s.shouldErr {
			require.Error(t, err)
			return
		}
		require.NoError(t, err)
		require.JSONEq(t, s.wantJSON, string(b))

		var got dataz.NullTimestamp
		err = json.Unmarshal([]byte(s.wantJSON), &got)
		require.NoError(t, err)
		require.True(t, s.val.Timestamp.Equal(got.Timestamp))
		require.Equal(t, s.val.Valid, got.Valid)
	}

	t.Run("Valid NullTimestamp", func(t *testing.T) {
		nowB, _ := json.Marshal(now)
		check(t, testScenario{
			val:      dataz.NullTimestamp{Timestamp: now, Valid: true},
			wantJSON: string(nowB),
		})
	})

	t.Run("Invalid NullTimestamp", func(t *testing.T) {
		check(t, testScenario{
			val:      dataz.NullTimestamp{Valid: false},
			wantJSON: `null`,
		})
	})
}

func TestUnitNullStringRepresentation(t *testing.T) {
	nsValid := dataz.NullString{StringVal: "hello", Valid: true}
	nsInvalid := dataz.NullString{Valid: false}
	require.Equal(t, "hello", nsValid.String())
	require.Equal(t, "<null>", nsInvalid.String())

	niValid := dataz.NullInt64{Int64: 42, Valid: true}
	niInvalid := dataz.NullInt64{Valid: false}
	require.Equal(t, "42", niValid.String())
	require.Equal(t, "<null>", niInvalid.String())

	nfValid := dataz.NullFloat64{Float64: 1.23, Valid: true}
	nfInvalid := dataz.NullFloat64{Valid: false}
	require.Equal(t, "1.23", nfValid.String())
	require.Equal(t, "<null>", nfInvalid.String())

	nbValid := dataz.NullBool{Bool: true, Valid: true}
	nbInvalid := dataz.NullBool{Valid: false}
	require.Equal(t, "true", nbValid.String())
	require.Equal(t, "<null>", nbInvalid.String())

	ntTime := time.Date(2026, 11, 25, 12, 0, 0, 0, time.UTC)
	ntValid := dataz.NullTimestamp{Timestamp: ntTime, Valid: true}
	ntInvalid := dataz.NullTimestamp{Valid: false}
	require.Equal(t, "2026-11-25T12:00:00Z", ntValid.String())
	require.Equal(t, "<null>", ntInvalid.String())
}
