package storezfile

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"sync"

	"github.com/infinity6-ai/gox/commonz/configz"
	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/filez"
	"github.com/infinity6-ai/gox/commonz/idgen"
	"github.com/infinity6-ai/gox/commonz/ioz"
	"github.com/infinity6-ai/gox/storez/storez"
)

var I6StorezFileBaseDirEncoded = configz.Create("I6_STOREZ_FILE_BASE_DIR", "")

var mu sync.Mutex
var mutexes = map[string]*sync.RWMutex{}

type StorezStrategyFile struct {
	basedir string
}

func InitEmulator(ctx context.Context, projectId string) (string, io.Closer) {
	if projectId == "" {
		projectId = fmt.Sprintf("demo-%s", idgen.Hex())
	}
	tempDir := filez.CreateTempDir("i6-storez")
	reverter, err := I6StorezFileBaseDirEncoded.SetEncoded(ctx, "%s", tempDir)
	errorz.Check(err)
	init := Init(ctx)
	return projectId, ioz.CloserV(func() {
		defer init.Close()
		defer reverter.Close()
	})
}

func Init(ctx context.Context) io.Closer {
	return errorz.Check2(storez.I6StorezStrategyEncoded.SetEncoded(ctx, "file"))
}

func New(ctx context.Context, projectId string, db string, schema *storez.StorezSchema) storez.StorezStrategy {
	return Open(ctx, projectId, db)
}

func Open(ctx context.Context, projectId string, db string) *StorezStrategyFile {
	basedir := I6StorezFileBaseDirEncoded.ReqDecoded(ctx)
	basedir = filepath.Join(basedir, projectId, db)

	return &StorezStrategyFile{basedir: basedir}
}

func (me *StorezStrategyFile) getOrCreateDB(table string) *DiskDB {
	mu.Lock()
	defer mu.Unlock()

	mutex, ok := mutexes[table]
	if !ok {
		mutex = &sync.RWMutex{}
		mutexes[table] = mutex
	}

	db := NewDiskDB(me.basedir, table, mutex)

	return db
}

func (me *StorezStrategyFile) Close() error {
	return nil
}

func (me *StorezStrategyFile) Transaction(ctx context.Context, callback func(storez.StorezStrategy), opts ...storez.TransactionOption) {
	callback(me)
}

func (me *StorezStrategyFile) PutAll(ctx context.Context, table string, entities []map[string]*storez.Value) {
	db := me.getOrCreateDB(table)
	for _, entity := range entities {
		db.Upsert(entity)
	}
}

func (me *StorezStrategyFile) DeleteAll(ctx context.Context, table string, ids []string) {
	db := me.getOrCreateDB(table)
	for _, id := range ids {
		db.Delete(id)
	}
}

func (me *StorezStrategyFile) GetAll(ctx context.Context, table string, ids []string) []map[string]*storez.Value {
	db := me.getOrCreateDB(table)
	ret := []map[string]*storez.Value{}
	for _, id := range ids {
		row := db.Get(id)
		if row != nil {
			ret = append(ret, row)
		}
	}
	return ret
}

func (me *StorezStrategyFile) GetTableNames(ctx context.Context) []string {
	return GetTableNames(ctx, me.basedir)
}
