package dataz

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnitLimitedString(t *testing.T) {
	s := "1234567890"
	limited := Limited(s, 5)
	assert.Equal(t, "12345", limited)
}

func TestUnitLimitedStringWithinLimit(t *testing.T) {
	s := "12345"
	limited := Limited(s, 5)
	assert.Equal(t, "12345", limited)
}

func TestUnitLimitedBytes(t *testing.T) {
	s := []byte("1234567890")
	limited := Limited(s, 5)
	assert.Equal(t, []byte("12345"), limited)
}

func TestUnitLimitedBytesWithinLimit(t *testing.T) {
	s := []byte("12345")
	limited := Limited(s, 5)
	assert.Equal(t, []byte("12345"), limited)
}
