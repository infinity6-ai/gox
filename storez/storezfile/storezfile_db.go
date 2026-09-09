package storezfile

import (
	"context"
	"maps"
	"os"
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
	os.MkdirAll(tablePath, os.ModePerm)
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
	filez.Write(rowFile, jsonz.MustFormat(data).Bytes())
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
	jsonz.MustParse(filez.MustReadFile(rowFile, 10*1024*1024).Bytes(), &ret)
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

	files, err := os.ReadDir(me.tablePath)
	errorz.Check(err)

	for _, file := range files {
		id := file.Name()[:len(file.Name())-len(".json")]
		data := me.internalGet(id)
		callback(data)
	}
}

func (me *DiskDB) List() []map[string]*storez.Value {
	rows := []map[string]*storez.Value{}
	me.Walk(func(row map[string]*storez.Value) {
		rows = append(rows, row)
	})
	return rows
}

func (me *DiskDB) Drop() {
	me.Walk(func(row map[string]*storez.Value) {
		id := row["id"].Value.(string)
		me.internalDelete(id)
	})
}

func GetTableNames(ctx context.Context, basedir string) []string {
	dirs, err := os.ReadDir(basedir)
	errorz.Check(err)

	ret := []string{}
	for _, dir := range dirs {
		ret = append(ret, dir.Name())
	}

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
