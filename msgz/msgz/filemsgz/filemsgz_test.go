package filemsgz_test

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/infinity6-ai/gox/commonz/filez"
	"github.com/infinity6-ai/gox/msgz/msgz"
	"github.com/infinity6-ai/gox/msgz/msgz/filemsgz"
	"github.com/stretchr/testify/require"
)

func createMessage(payload string) *msgz.Message {
	return &msgz.Message{
		Payload: []byte(payload),
	}
}

func TestUnitPublishAndPull(t *testing.T) {
	ctx := context.Background()
	store := filemsgz.NewTemporaryMessageStore(ctx)
	defer store.Shutdown()
	topic := "test-topic"

	msg1 := createMessage(`{"hello":"world1"}`)
	msg2 := createMessage(`{"hello":"world2"}`)
	msg2.OrderingKey = "key"

	pulled := store.Pull(ctx, topic, 2, msgz.PullOptions{})
	require.Equal(t, 0, len(pulled.Messages))

	store.Publish(ctx, topic, &msgz.Messages{Messages: []*msgz.Message{msg1}})
	time.Sleep(1 * time.Millisecond)
	store.Publish(ctx, topic, &msgz.Messages{Messages: []*msgz.Message{msg2}})

	files := filez.DirListLimited(store.Basedir().MustJoinString("created", "topic", topic).String(), ".*", 10)
	require.Len(t, files, 2)

	pulled = store.Pull(ctx, topic, 2, msgz.PullOptions{})
	require.NotEmpty(t, pulled.Messages[0].Message.Id)
	require.Greater(t, pulled.Messages[0].Message.PublishedAt, int64(0))
	require.NotEmpty(t, pulled.Messages[0].AckId)
	require.Empty(t, pulled.Messages[0].Message.OrderingKey)

	require.NotEmpty(t, pulled.Messages[1].Message.Id)
	require.Greater(t, pulled.Messages[1].Message.PublishedAt, int64(0))
	require.NotEmpty(t, pulled.Messages[1].AckId)
	require.Equal(t, "key", pulled.Messages[1].Message.OrderingKey)

	require.Len(t, pulled.Messages, 2)

	// Should have moved files to "fetched"
	fetchedFiles := filez.DirListLimited(store.Basedir().MustJoinString("fetched", "topic", topic).String(), ".*", 10)
	require.Len(t, fetchedFiles, 2)
}

func TestUnitAck(t *testing.T) {
	ctx := context.Background()
	store := filemsgz.NewTemporaryMessageStore(ctx)
	defer store.Shutdown()
	topic := "ack-topic"

	msg := createMessage(`{"ack":"me"}`)
	store.Publish(ctx, topic, &msgz.Messages{Messages: []*msgz.Message{msg}})

	pulled := store.Pull(ctx, topic, 1, msgz.PullOptions{})
	require.Len(t, pulled.Messages, 1)

	id := pulled.Messages[0].Message.Id
	fetchedPath := store.Basedir().MustJoinString("fetched", "topic", topic, id+".json.gz")
	require.FileExists(t, fetchedPath.String())

	store.Ack(ctx, &msgz.Ids{Ids: []string{pulled.Messages[0].AckId}})

	require.NoFileExists(t, fetchedPath.String())
}

func TestUnitNack(t *testing.T) {
	ctx := context.Background()
	store := filemsgz.NewTemporaryMessageStore(ctx)
	defer store.Shutdown()
	topic := "nack-topic"

	msg := createMessage(`{"nack":"me"}`)
	store.Publish(ctx, topic, &msgz.Messages{Messages: []*msgz.Message{msg}})

	pulled := store.Pull(ctx, topic, 1, msgz.PullOptions{})
	require.Len(t, pulled.Messages, 1)

	id := pulled.Messages[0].Message.Id
	fetchedPath := store.Basedir().MustJoinString("fetched", "topic", topic, id+".json.gz")
	createdPath := store.Basedir().MustJoinString("created", "topic", topic, id+".json.gz")

	require.FileExists(t, fetchedPath.String())

	store.Nack(ctx, &msgz.Ids{Ids: []string{pulled.Messages[0].AckId}})

	require.FileExists(t, createdPath.String())
	require.NoFileExists(t, fetchedPath.String())
}

func TestUnitNackAll(t *testing.T) {
	ctx := context.Background()
	store := filemsgz.NewTemporaryMessageStore(ctx)
	defer store.Shutdown()
	topic := "nackall-topic"

	// Publish and pull messages to move them to fetched
	for i := 0; i < 3; i++ {
		store.Publish(ctx, topic, &msgz.Messages{Messages: []*msgz.Message{createMessage(`{"msg":` + strconv.Itoa(i) + `}`)}})
	}
	store.Pull(ctx, topic, 3, msgz.PullOptions{})

	fetchedPath := store.Basedir().MustJoinString("fetched", "topic", topic)
	require.Len(t, filez.DirListLimited(fetchedPath.String(), ".*", 10), 3)

	store.NackAll(ctx, topic)

	createdPath := store.Basedir().MustJoinString("created", "topic", topic)
	require.Len(t, filez.DirListLimited(createdPath.String(), ".*", 10), 3)
	require.Len(t, filez.DirListLimited(fetchedPath.String(), ".*", 10), 0)
}
