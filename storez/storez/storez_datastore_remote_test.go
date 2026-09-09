package storez_test

import (
	"context"
	"testing"
)

func TestRemoteDatastoreClient(t *testing.T) {
	ctx := context.Background()
	client := Open(ctx, "datastore", "i6-core-prod", "bla1")
	defer client.Close()
	CheckClient(t, client)
}

func TestRemoteDatastoreGetTableNames(t *testing.T) {
	ctx := context.Background()
	client := Open(ctx, "datastore", "i6-core-prod", "bla1")
	defer client.Close()
	CheckGetTableNames(t, client)
}

func TestRemoteDatastoreClientTransaction(t *testing.T) {
	ctx := context.Background()
	client := Open(ctx, "datastore", "i6-core-prod", "bla1")
	defer client.Close()
	CheckClientTransaction(t, client)
}

func TestRemoteDatastoreBasic(t *testing.T) {
	ctx := context.Background()
	client := Open(ctx, "datastore", "i6-core-prod", "bla1")
	defer client.Close()
	CheckBasic(t, client)
}

func TestRemoteDatastoreTransaction(t *testing.T) {
	ctx := context.Background()
	client := Open(ctx, "datastore", "i6-core-prod", "bla1")
	defer client.Close()
	CheckTransaction(t, client)
}

func TestRemoteDatastoreQuery(t *testing.T) {
	ctx := context.Background()
	client := Open(ctx, "datastore", "i6-core-prod", "bla1")
	defer client.Close()
	CheckQuery(t, client)
}
