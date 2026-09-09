package storez_test

import (
	"context"
	"testing"

	"github.com/infinity6-ai/gox/storez/storezfile"
)

func TestUnitFileClient(t *testing.T) {
	ctx := context.Background()
	projectId, storeReverter := storezfile.InitEmulator(ctx, "")
	defer storeReverter.Close()
	client := Open(ctx, "file", projectId, "mytest")
	defer client.Close()
	CheckClient(t, client)
}

func TestUnitFileGetTableNames(t *testing.T) {
	ctx := context.Background()
	projectId, storeReverter := storezfile.InitEmulator(ctx, "")
	defer storeReverter.Close()
	client := Open(ctx, "file", projectId, "mytest")
	defer client.Close()
	CheckGetTableNames(t, client)
}

func TestUnitFileClientTransaction(t *testing.T) {
	ctx := context.Background()
	projectId, storeReverter := storezfile.InitEmulator(ctx, "")
	defer storeReverter.Close()
	client := Open(ctx, "file", projectId, "mytest")
	defer client.Close()
	CheckClientTransaction(t, client)
}

func TestUnitFileBasic(t *testing.T) {
	ctx := context.Background()
	projectId, storeReverter := storezfile.InitEmulator(ctx, "")
	defer storeReverter.Close()
	client := Open(ctx, "file", projectId, "mytest")
	defer client.Close()
	CheckBasic(t, client)
}

func TestUnitFileTransaction(t *testing.T) {
	ctx := context.Background()
	projectId, storeReverter := storezfile.InitEmulator(ctx, "")
	defer storeReverter.Close()
	client := Open(ctx, "file", projectId, "mytest")
	defer client.Close()
	CheckTransaction(t, client)
}

func TestUnitFileQuery(t *testing.T) {
	ctx := context.Background()
	projectId, storeReverter := storezfile.InitEmulator(ctx, "")
	defer storeReverter.Close()
	client := Open(ctx, "file", projectId, "mytest")
	defer client.Close()
	CheckQuery(t, client)
}
