package stringz_test

import (
	"testing"

	"github.com/infinity6-ai/gox/commonz/stringz"
	"github.com/stretchr/testify/assert"
)

func TestUnitShortPretty(t *testing.T) {
	result := stringz.ShortPretty(5, "-", "abcde fghij")
	assert.Equal(t, "abcde-fghij", result)
}

func TestUnitShortPrettyWithSpecialChars(t *testing.T) {
	result := stringz.ShortPretty(5, "_", "abcde@fghij")
	assert.Equal(t, "abcde_fghij", result)
}

func TestUnitShortPrettyWithLongPart(t *testing.T) {
	result := stringz.ShortPretty(3, "-", "abcdefghij")
	assert.Equal(t, "abc", result)
}

func TestUnitShortPrettyWithNoSpecialChars(t *testing.T) {
	result := stringz.ShortPretty(5, "-", "abcdefghij")
	assert.Equal(t, "abcde", result)
}

func TestUnitShortPrettyWithMultipleDivs(t *testing.T) {
	result := stringz.ShortPretty(5, "-", "ab c d-e_f@g.h i")
	assert.Equal(t, "ab-c-d-e-f-g-h-i", result)
}
