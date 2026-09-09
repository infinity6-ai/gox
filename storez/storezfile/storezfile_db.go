package storezfile

import (
	"context"
	"io/fs"
	"maps"
	"path/filepath"
	"sync"

	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/filez"
	"github.com/infinity6-ai/gox/commonz/jsonz"
	"github.com/infinity6-ai/gox/storez/storez"
	"github.com/infinity6-ai/gox/storez/storezvalidation"
)

type DiskDB struct {
	mu        *sync.RWMutex
	tablePath string
}

func NewDiskDB(basePath, tableName string, mu *sync.RWMutex) *DiskDB {
	storezvalidation.TableName(tableName)
	tablePath := filepath.Join(basePath, tableName)
	errorz.Check(filez.MkdirAll(tablePath))
	return &DiskDB{
		tablePath: tablePath,
		mu:        mu,
	}
}

func (me *DiskDB) Upsert(original map[string]*storez.Value) {
	ValidateRow(original)
	id := original["id"].Value.(string)
	data := maps.Clone(original)
	delete(data, "id")

	me.mu.Lock()
	defer me.mu.Unlock()

	rowFile := filepath.Join(me.tablePath, id+".json")
	errorz.Check(filez.WriteFile(rowFile, jsonz.MustFormat(data).Bytes()))
}

func (me *DiskDB) internalDelete(id string) {
	storezvalidation.Id(id)
	rowFile := filepath.Join(me.tablePath, id+".json")
	errorz.Check(filez.Remove(rowFile))
}

func (me *DiskDB) Delete(id string) {
	me.mu.Lock()
	defer me.mu.Unlock()
	me.internalDelete(id)
}

func (me *DiskDB) internalGet(id string) map[string]*storez.Value {
	storezvalidation.Id(id)

	rowFile := filepath.Join(me.tablePath, id+".json")
	ret := map[string]*storez.Value{}
	if !filez.FileExists(rowFile) {
		return nil
	}
	fileContent, err := filez.ReadFile(rowFile, 10*1024*1024)
	errorz.Check(err)
	jsonz.MustParse(fileContent.Bytes(), &ret)
	storez.FixRow(ret)
	ret["id"] = &storez.Value{
		Value:   id,
		Indexed: true,
	}
	return ret
}

func (me *DiskDB) Get(id string) map[string]*storez.Value {
	me.mu.RLock()
	defer me.mu.RUnlock()
	return me.internalGet(id)
}

func (me *DiskDB) Walk(callback func(row map[string]*storez.Value)) {
	me.mu.RLock()
	defer me.mu.RUnlock()

	err := filez.Ls(me.tablePath, func(idx int, path string, f fs.DirEntry) (bool, error) {
		id := f.Name()[:len(f.Name())-len(".json")]
		data := me.internalGet(id)
		callback(data)
		return false, nil
	})
	errorz.Check(err)
}

func (me *DiskDB) List() []map[string]*storez.Value {
	rows := []map[string]*storez.Value{}
	me.Walk(func(row map[string]*storez.Value) {
		rows = append(rows, row)
	})
	return rows
}

func (me *DiskDB) Drop() {
	me.mu.Lock()
	defer me.mu.Unlock()

	err := filez.Ls(me.tablePath, func(idx int, path string, f fs.DirEntry) (bool, error) {
		return false, filez.Remove(path)
	})
	errorz.Check(err)
}

func GetTableNames(ctx context.Context, basedir string) []string {
	ret := []string{}
	err := filez.Ls(basedir, func(idx int, path string, f fs.DirEntry) (bool, error) {
		if f.IsDir() {
			ret = append(ret, f.Name())
		}
		return false, nil
	})
	errorz.Check(err)
	return ret
}

func ValidateRow(rows ...map[string]*storez.Value) {
	for _, row := range rows {
		for k, v := range row {
			if k == "id" {
				storezvalidation.Id(v.Value.(string))
			} else {
				storezvalidation.Value(v.Value)
			}
		}
	}
}
