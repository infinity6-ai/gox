package storez_test

import (
	"context"
	"testing"

	"github.com/infinity6-ai/gox/storez/storezdatastore"
)

// func TestManualDatastoreEmulatorClient(t *testing.T) {
// 	projectId, storeReverter := storezdatastore.InitEmulator("demo-project")
// 	defer storeReverter.Close()
// 	client := Open(projectId, "mytest")
// 	defer client.Close()
// 	CheckClient(t, client)
// }

func TestManualDatastoreEmulatorClientTransaction(t *testing.T) {
	ctx := context.Background()
	projectId, storeReverter := storezdatastore.InitEmulator(context.Background(), "")
	defer storeReverter.Close()
	client := Open(ctx, projectId, "mytest")
	defer client.Close()
	CheckClientTransaction(t, client)
}

func TestManualDatastoreEmulatorBasic(t *testing.T) {
	ctx := context.Background()
	projectId, storeReverter := storezdatastore.InitEmulator(context.Background(), "")
	defer storeReverter.Close()
	client := Open(ctx, projectId, "mytest")
	defer client.Close()
	CheckBasic(t, client)
}

func TestManualDatastoreEmulatorTransaction(t *testing.T) {
	ctx := context.Background()
	projectId, storeReverter := storezdatastore.InitEmulator(context.Background(), "")
	defer storeReverter.Close()
	client := Open(ctx, projectId, "mytest")
	defer client.Close()
	CheckTransaction(t, client)
}

func TestManualDatastoreEmulatorQuery(t *testing.T) {
	ctx := context.Background()
	projectId, storeReverter := storezdatastore.InitEmulator(context.Background(), "")
	defer storeReverter.Close()
	client := Open(ctx, projectId, "mytest")
	defer client.Close()
	CheckQuery(t, client)
}
