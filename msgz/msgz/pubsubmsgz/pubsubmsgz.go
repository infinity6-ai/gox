package pubsubmsgz

import (
	"context"
	"fmt"
	"strings"

	"cloud.google.com/go/pubsub"
	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/logz"
	"github.com/infinity6-ai/gox/commonz/validation/checker"
	"github.com/infinity6-ai/gox/msgz/msgz"

	"google.golang.org/api/option"

	pb "cloud.google.com/go/pubsub/apiv1"
	"cloud.google.com/go/pubsub/apiv1/pubsubpb"
)

type tlogger logz.Type

var logger = logz.Create(tlogger(true))

type pubsubStrategy struct {
	projectId string
}

func (ps *pubsubStrategy) NackAll(ctx context.Context, sub string) {
	panic("unsupported")
}

func openSubscriberClient(ctx context.Context) (*pb.SubscriberClient, error) {
	client, err := pb.NewSubscriberClient(ctx, option.WithTelemetryDisabled())
	if err != nil {
		return nil, err
	}
	return client, err
}

func openPublishClient(ctx context.Context, projectId string) (*pubsub.Client, error) {
	client, err := pubsub.NewClient(ctx, projectId, option.WithTelemetryDisabled())

	if err != nil {
		return nil, err
	}
	return client, err
}

func parseAckId(fullId string) (string, string) {
	parts := strings.SplitN(fullId, ".", 2)
	sub := parts[0]
	checker.StrNotEmpty(sub, "sub")
	ackId := parts[1]
	checker.StrNotEmpty(ackId, "ackId")
	return sub, ackId
}

func (ps *pubsubStrategy) Ack(ctx context.Context, ids *msgz.Ids) {
	if len(ids.Ids) == 0 {
		return
	}

	client, err := openSubscriberClient(ctx)
	errorz.Check(err)
	defer client.Close()

	ackIds := make([]string, len(ids.Ids))
	sub := ""
	for i, fullId := range ids.Ids {
		currentSub, ackId := parseAckId(fullId)
		if sub == "" {
			sub = currentSub
		}
		if sub != currentSub {
			panic("multiple subs not supported")
		}
		ackIds[i] = ackId
	}

	subName := fmt.Sprintf("projects/%s/subscriptions/%s", ps.projectId, sub)

	req := &pubsubpb.AcknowledgeRequest{
		Subscription: subName,
		AckIds:       ackIds,
	}

	err = client.Acknowledge(ctx, req)
	errorz.Check(err)
}

func (ps *pubsubStrategy) Nack(ctx context.Context, ids *msgz.Ids) {
	panic("unsupported")
}

func (ps *pubsubStrategy) Pull(ctx context.Context, sub string, limit int, opts msgz.PullOptions) *msgz.ManagedMessages {
	checker.Greater(limit, 0, "limit")
	client, err := openSubscriberClient(ctx)
	errorz.Check(err)
	defer client.Close()

	subName := fmt.Sprintf("projects/%s/subscriptions/%s", ps.projectId, sub)

	req := &pubsubpb.PullRequest{
		Subscription:      subName,
		MaxMessages:       int32(limit),
		ReturnImmediately: opts.ReturnImmediately,
	}

	resp, err := client.Pull(ctx, req)
	errorz.Check(err)

	if resp == nil {
		return &msgz.ManagedMessages{
			Messages: []*msgz.ManagedMessage{},
		}
	}
	ret := &msgz.ManagedMessages{}
	ret.Messages = make([]*msgz.ManagedMessage, len(resp.ReceivedMessages))
	for i, msg := range resp.ReceivedMessages {
		ret.Messages[i] = &msgz.ManagedMessage{
			AckId: fmt.Sprintf("%s.%s", sub, msg.AckId),
			Message: &msgz.Message{
				Id:          msg.Message.MessageId,
				Attributes:  msg.Message.Attributes,
				Payload:     msg.Message.Data,
				PublishedAt: msg.Message.PublishTime.AsTime().UnixMilli(),
				OrderingKey: msg.Message.OrderingKey,
			},
		}
	}
	return ret
}

func (ps *pubsubStrategy) Close() error {
	return nil

}

func (ps *pubsubStrategy) Publish(ctx context.Context, topic string, msgs *msgz.Messages) {
	if len(msgs.Messages) == 0 {
		return
	}

	client, err := openPublishClient(ctx, ps.projectId)
	errorz.Check(err)
	defer client.Close()

	t := client.Topic(topic)
	t.EnableMessageOrdering = msgs.Messages[0].OrderingKey != ""

	results := []*pubsub.PublishResult{}
	for _, msg := range msgs.Messages {
		results = append(results, t.Publish(ctx, &pubsub.Message{
			Data:        msg.Payload,
			Attributes:  msg.Attributes,
			OrderingKey: msg.OrderingKey,
		}))
	}

	for _, result := range results {
		_, err := result.Get(ctx)
		errorz.Check(err)
	}
}

func NewPublisher(ctx context.Context, projectId string) msgz.Publisher {
	return &pubsubStrategy{projectId: projectId}
}

func NewPuller(ctx context.Context, projectId string) msgz.Puller {
	return &pubsubStrategy{projectId: projectId}
}
