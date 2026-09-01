package protoz_test

import (
	"testing"

	"github.com/infinity6-ai/gox/commonz/protoz"
	"github.com/infinity6-ai/gox/commonz/protoz/protozsample"
	"github.com/stretchr/testify/assert"
)

func TestUnitParseFormat(t *testing.T) {
	// Create a dummy message
	originalDummy := &protozsample.Dummy{}
	originalDummy.Id = 1
	originalDummy.Name = "dummy1"

	// Format the dummy message to bytes
	data := protoz.Format(originalDummy)

	// Parse the bytes back into a dummy message
	parsedDummy := &protozsample.Dummy{}
	ret := protoz.Parse(data, parsedDummy)
	assert.Same(t, parsedDummy, ret)

	// Assert that the parsed message matches the original
	assert.Equal(t, originalDummy.Id, parsedDummy.Id)
	assert.Equal(t, originalDummy.Name, parsedDummy.Name)
}

// {Id: 1, Name: "dummy1"},
// {Id: 2, Name: "dummy2"},
// {Id: 3, Name: "dummy3"},

func TestUnitFormatParseSlice(t *testing.T) {
	// Create a slice of dummy messages
	dummies := []*protozsample.Dummy{
		protozsample.NewDummy(1, "dummy1"),
		protozsample.NewDummy(2, "dummy2"),
		protozsample.NewDummy(3, "dummy3"),
	}

	// Format the slice to bytes
	data := protoz.FormatSlice(dummies)

	// // Parse the bytes back into a slice of dummy messages
	parsedDummies := protoz.ParseSlice[*protozsample.Dummy](data)

	// // Assert that the parsed slice matches the original
	assert.Equal(t, len(dummies), len(parsedDummies))
	for i := range dummies {
		assert.Equal(t, dummies[i].Id, parsedDummies[i].Id)
		assert.Equal(t, dummies[i].Name, parsedDummies[i].Name)
	}
}
