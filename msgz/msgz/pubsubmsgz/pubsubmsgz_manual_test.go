package pubsubmsgz_test

import (
	"context"
	"testing"
	"time"

	"github.com/infinity6-ai/gox/msgz/msgz"
	"github.com/infinity6-ai/gox/msgz/msgz/pubsubmsgz"
	"github.com/stretchr/testify/assert"
)

const (
	projectID      = "i6-core-prod"
	topicID        = "demo-topic"
	subscriptionID = "demo-sub"
)

func printNow(msg string) {
	println(msg, time.Now().UTC().Format("2006-01-02T15:04:05.000Z"))
}

func TestManualPubsub(t *testing.T) {
	ctx := context.Background()
	publisher := pubsubmsgz.NewPublisher(ctx, projectID)
	publisher.Publish(ctx, topicID, &msgz.Messages{
		Messages: []*msgz.Message{
			{
				Attributes:  map[string]string{"a": "v"},
				Payload:     []byte("mypayload"),
				OrderingKey: "key",
			},
		},
	})

	puller := pubsubmsgz.NewPuller(ctx, projectID)
	printNow("pulling")
	messages := puller.Pull(ctx, subscriptionID, 10, msgz.PullOptions{
		ReturnImmediately: true,
	})
	printNow("pulled")

	ids := messages.AckIds()
	for _, id := range ids.Ids {
		assert.NotEmpty(t, id)
	}
	puller.Ack(ctx, ids)

	assert.Equal(t, "mypayload", string(messages.Messages[0].Message.Payload))
	assert.Equal(t, "v", messages.Messages[0].Message.Attributes["a"])
	assert.NotEmpty(t, messages.Messages[0].Message.Id)
	assert.Greater(t, messages.Messages[0].Message.PublishedAt, int64(0))
	assert.Equal(t, "key", messages.Messages[0].Message.OrderingKey)

	printNow("pulling")
	messages = puller.Pull(ctx, subscriptionID, 10, msgz.PullOptions{
		ReturnImmediately: true,
	})
	printNow("pulled")
	assert.Equal(t, 0, len(messages.Messages))

}
