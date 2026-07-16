package dataz_test

import (
	"testing"

	"github.com/infinity6-ai/gox/commonz/dataz"
	"github.com/stretchr/testify/assert"
)

func TestUnitLimitedString(t *testing.T) {
	s := "1234567890"
	limited := dataz.Limited(s, 5)
	assert.Equal(t, "12345", limited)
}

func TestUnitLimitedStringWithinLimit(t *testing.T) {
	s := "12345"
	limited := dataz.Limited(s, 5)
	assert.Equal(t, "12345", limited)
}

func TestUnitLimitedBytes(t *testing.T) {
	s := []byte("1234567890")
	limited := dataz.Limited(s, 5)
	assert.Equal(t, []byte("12345"), limited)
}

func TestUnitLimitedBytesWithinLimit(t *testing.T) {
	s := []byte("12345")
	limited := dataz.Limited(s, 5)
	assert.Equal(t, []byte("12345"), limited)
}
