package pubsubsample1

import (
	"context"
	"fmt"
	"log"
	"time"

	"cloud.google.com/go/pubsub"
)

const (
	projectID      = "i6-core-prod"
	topicID        = "demo-topic"
	subscriptionID = "demo-sub"
)

func Abc() {
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("pubsub.NewClient: %v", err)
	}
	defer client.Close()

	// Ensure topic and subscription exist
	topic := ensureTopic(ctx, client, topicID)
	sub := ensureSubscription(ctx, client, subscriptionID, topic)

	for i := 0; i < 10; i++ {
		// Publish a message
		msgID, err := topic.Publish(ctx, &pubsub.Message{
			Data:       []byte("Hello, world!"),
			Attributes: map[string]string{"origin": "golang"},
		}).Get(ctx)
		if err != nil {
			log.Fatalf("Publish: %v", err)
		}
		fmt.Printf("Published message ID: %s\n", msgID)
	}

	// ---- Receive using callback ----
	fmt.Println("\n[Callback-based receive]")
	sub.ReceiveSettings = pubsub.ReceiveSettings{
		MaxOutstandingMessages: 10,               // Max messages processing concurrently
		MaxOutstandingBytes:    1e6,              // Max bytes outstanding (optional)
		NumGoroutines:          5,                // Number of goroutines pulling messages
		MaxExtension:           10 * time.Minute, // Max ack deadline extension duration
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	err = sub.Receive(ctx, func(ctx context.Context, m *pubsub.Message) {
		fmt.Printf("Received via callback: %s\n", string(m.Data))
		m.Ack()
	})
	if err != nil {
		log.Printf("Receive: %v", err)
	}

	// ---- Pull messages manually using .Next(n) ----
	// fmt.Println("\n[Pull-based receive]")
	// pullCtx, cancelPull := context.WithTimeout(context.Background(), 10*time.Second)
	// defer cancelPull()

	// it, err := sub.Pull(pullCtx)
	// if err != nil {
	// 	log.Fatalf("Pull: %v", err)
	// }
	// defer it.Stop()

	// for i := 0; i < 4; i++ {
	// 	msg, err := it.Next()
	// 	if err == iterator.Done {
	// 		fmt.Println("No more messages.")
	// 		break
	// 	}
	// 	if err != nil {
	// 		log.Fatalf("Next: %v", err)
	// 	}

	// 	fmt.Printf("Pulled message: %s\n", string(msg.Data))

	// 	// Ack message manually
	// 	msg.Ack()
	// 	fmt.Println("Acked message.")
	// }
}

// ensureTopic creates the topic if it doesn't exist
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

// ensureSubscription creates the subscription if it doesn't exist
func ensureSubscription(ctx context.Context, client *pubsub.Client, subID string, topic *pubsub.Topic) *pubsub.Subscription {
	sub := client.Subscription(subID)
	exists, err := sub.Exists(ctx)
	if err != nil {
		log.Fatalf("Checking subscription existence: %v", err)
	}
	if !exists {
		sub, err = client.CreateSubscription(ctx, subID, pubsub.SubscriptionConfig{
			Topic:             topic,
			AckDeadline:       10 * time.Second,
			RetentionDuration: 24 * time.Hour,
		})
		if err != nil {
			log.Fatalf("CreateSubscription: %v", err)
		}
	}
	return sub
}
