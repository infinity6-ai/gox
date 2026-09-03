package msgz

import (
	"context"
)

type PushedMessageBody struct {
	Message      *PushedMessage `json:"message"`
	Subscription string         `json:"subscription"`
}

type PushedMessage struct {
	Attributes map[string]string `json:"attributes"`
	MessageId  string            `json:"message_id"`
	Data       []byte            `json:"data,omitempty"`
}

type Message struct {
	Id          string            `json:"id,omitempty"`
	PublishedAt int64             `json:"published_at,omitempty"`
	Attributes  map[string]string `json:"attributes,omitempty"`
	Payload     []byte            `json:"payload,omitempty"`
	OrderingKey string            `json:"ordering_key,omitempty"`
}

type Messages struct {
	Messages []*Message `json:"messages,omitempty"`
}

func (m *Messages) AddMessages(msgs ...*Message) {
	m.Messages = append(m.Messages, msgs...)
}

type ManagedMessage struct {
	AckId   string   `json:"ack_id,omitempty"`
	Message *Message `json:"message,omitempty"`
}

type ManagedMessages struct {
	Messages []*ManagedMessage `json:"messages,omitempty"`
}

func (me *ManagedMessages) AckIds() *Ids {
	ret := &Ids{}
	ret.Ids = make([]string, len(me.Messages))
	for i, msg := range me.Messages {
		ret.Ids[i] = msg.AckId
	}
	return ret
}

type Ids struct {
	Ids []string `json:"ids,omitempty"`
}

type Publisher interface {
	Publish(ctx context.Context, topic string, msgs *Messages)
	Close() error
}

type PullOptions struct {
	ReturnImmediately bool
}

type Puller interface {
	Pull(ctx context.Context, sub string, limit int, opts PullOptions) *ManagedMessages
	Ack(ctx context.Context, ids *Ids)
	Nack(ctx context.Context, ids *Ids)
	NackAll(ctx context.Context, sub string)
	Close() error
}
