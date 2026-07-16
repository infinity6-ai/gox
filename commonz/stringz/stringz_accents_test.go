package stringz

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnitRemoveAccents(t *testing.T) {
	assert.Equal(t, "aeiou", RemoveAccents("áéíóú"))
	assert.Equal(t, "AEIOU", RemoveAccents("ÁÉÍÓÚ"))
	assert.Equal(t, "Caiu o acento", RemoveAccents("Caiu o âçento"))
	assert.Equal(t, "Nao ha acentos aqui", RemoveAccents("Não há acentos aqui"))
	assert.Equal(t, "Nem aqui", RemoveAccents("Nem aqui"))
	assert.Equal(t, "", RemoveAccents(""))
	assert.Equal(t, "123", RemoveAccents("123"))
}
