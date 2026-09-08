package storezfile_test

import (
	"os"
	"sync"
	"testing"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/storez/storez"
	"github.com/infinity6-ai/gox/storez/storezfile"
	"github.com/stretchr/testify/assert"
)

func TestUnitDB(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "i6-storez-*")
	errorz.Check(err)
	db := storezfile.NewDiskDB(tempDir, "t", &sync.RWMutex{})

	row1 := map[string]*storez.Value{"id": {Value: "1", Indexed: true}, "name": {Value: "Alice", Indexed: true}, "age": {Value: int64(25), Indexed: false}}
	row2 := map[string]*storez.Value{"id": {Value: "2", Indexed: true}, "name": {Value: "Bob", Indexed: true}, "age": {Value: int64(30), Indexed: false}}
	row3 := map[string]*storez.Value{"id": {Value: "1", Indexed: true}, "name": {Value: "Alice", Indexed: true}, "age": {Value: int64(26), Indexed: false}}
	db.Upsert(row1)
	db.Upsert(row2)
	db.Upsert(row3)

	assert.Equal(t, row3, db.Get("1"))

	assert.Nil(t, db.Get("notfound"))

	assert.Equal(t, []map[string]*storez.Value{row3, row2}, db.List())

	db.Delete("2")
	assert.Equal(t, []map[string]*storez.Value{row3}, db.List())

	db.Drop()
	assert.Empty(t, db.List())

}
