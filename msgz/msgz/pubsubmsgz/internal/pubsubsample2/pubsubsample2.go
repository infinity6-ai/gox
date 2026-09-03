package pubsubsample2

import (
	"context"
	"fmt"
	"log"
	"time"

	"cloud.google.com/go/pubsub"
	pb "cloud.google.com/go/pubsub/apiv1"
	pubsubpb "cloud.google.com/go/pubsub/apiv1/pubsubpb"
)

const (
	projectID      = "i6-core-prod"
	topicID        = "demo-topic"
	subscriptionID = "demo-sub"
)

func Abc() {
	ctx := context.Background()

	// Ensure topic and subscription exist
	client2, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("pubsub.NewClient: %v", err)
	}
	defer client2.Close()
	topic := ensureTopic(ctx, client2, topicID)
	// Publish a message
	msgID, err := topic.Publish(ctx, &pubsub.Message{
		Data:       []byte("Hello, world!"),
		Attributes: map[string]string{"origin": "golang"},
	}).Get(ctx)
	if err != nil {
		log.Fatalf("Publish: %v", err)
	}
	fmt.Printf("Published message ID: %s\n", msgID)

	subName := fmt.Sprintf("projects/%s/subscriptions/%s", projectID, subscriptionID)

	client, err := pb.NewSubscriberClient(ctx)
	if err != nil {
		log.Fatalf("failed to create subscriber client: %v", err)
	}
	defer client.Close()

	req := &pubsubpb.PullRequest{
		Subscription:      subName,
		MaxMessages:       1,
		ReturnImmediately: false, // Set to false to wait for a message
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	resp, err := client.Pull(ctx, req)
	if err != nil {
		log.Fatalf("failed to pull message: %v", err)
	}

	if len(resp.ReceivedMessages) == 0 {
		fmt.Println("No messages received.")
		return
	}

	msg := resp.ReceivedMessages[0]
	fmt.Printf("Got message: %s\n", msg.Message.Data)

	// Acknowledge the message
	ackReq := &pubsubpb.AcknowledgeRequest{
		Subscription: subName,
		AckIds:       []string{msg.AckId},
	}
	if err := client.Acknowledge(ctx, ackReq); err != nil {
		log.Fatalf("failed to ack message: %v", err)
	}
}

func ensureTopic(ctx context.Context, client *pubsub.Client, topicID string) *pubsub.Topic {
	topic := client.Topic(topicID)
	exists, err := topic.Exists(ctx)
	if err != nil {
		log.Fatalf("Checking topic existence: %v", err)
	}
	if !exists {
		topic, err = client.CreateTopic(ctx, topicID)
		if err != nil {
			log.Fatalf("CreateTopic: %v", err)
		}
	}
	return topic
}
