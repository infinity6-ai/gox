package storezfile_test

import (
	"os"
	"sync"
	"testing"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/storez/storez"
	"github.com/infinity6-ai/gox/storez/storezfile"
	"github.com/stretchr/testify/require"
)

func TestUnitDB(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "i6-storez-*")
	errorz.Check(err)
	defer os.RemoveAll(tempDir)
	db := storezfile.NewDiskDB(tempDir, "t", &sync.RWMutex{})

	row1 := map[string]*storez.Value{"id": {Value: "1", Indexed: true}, "name": {Value: "Alice", Indexed: true}, "age": {Value: int64(25), Indexed: false}}
	row2 := map[string]*storez.Value{"id": {Value: "2", Indexed: true}, "name": {Value: "Bob", Indexed: true}, "age": {Value: int64(30), Indexed: false}}
	row3 := map[string]*storez.Value{"id": {Value: "1", Indexed: true}, "name": {Value: "Alice", Indexed: true}, "age": {Value: int64(26), Indexed: false}}

	t.Run("A_Upsert", func(t *testing.T) {
		db.Upsert(row1)
		db.Upsert(row2)
		db.Upsert(row3)
	})

	t.Run("B_Get", func(t *testing.T) {
		require.Equal(t, row3, db.Get("1"))
		require.Nil(t, db.Get("notfound"))
	})

	t.Run("C_List", func(t *testing.T) {
		require.ElementsMatch(t, []map[string]*storez.Value{row3, row2}, db.List())
	})

	t.Run("D_Delete", func(t *testing.T) {
		db.Delete("2")
		require.Equal(t, []map[string]*storez.Value{row3}, db.List())
	})

	t.Run("E_Drop", func(t *testing.T) {
		db.Drop()
		require.Empty(t, db.List())
	})
}
