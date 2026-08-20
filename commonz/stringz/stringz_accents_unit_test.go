package stringz_test

import (
	"testing"

	"github.com/infinity6-ai/gox/commonz/stringz"
	"github.com/stretchr/testify/assert"
)

func TestUnitRemoveAccents(t *testing.T) {
	assert.Equal(t, "aeiou", stringz.RemoveAccents("\u00e1\u00e9\u00ed\u00f3\u00fa"))
	assert.Equal(t, "AEIOU", stringz.RemoveAccents("\u00c1\u00c9\u00cd\u00d3\u00da"))
	assert.Equal(t, "Caiu o acento", stringz.RemoveAccents("Caiu o \u00e2\u00e7ento"))
	assert.Equal(t, "Nao ha acentos aqui", stringz.RemoveAccents("N\u00e3o h\u00e1 acentos aqui"))
	assert.Equal(t, "Nem aqui", stringz.RemoveAccents("Nem aqui"))
	assert.Equal(t, "", stringz.RemoveAccents(""))
	assert.Equal(t, "123", stringz.RemoveAccents("123"))
}
