package msgzservicez

import (
	"context"
	"fmt"
	"io"

	"github.com/infinity6-ai/gox/commonz/ioz"
	"github.com/infinity6-ai/gox/msgz/msgz"
	"github.com/infinity6-ai/gox/msgz/msgz/filemsgz"
	"github.com/infinity6-ai/gox/msgz/msgz/pubsubmsgz"
)

type MsgzService struct {
	NewPuller    func(ctx context.Context, createOpts msgz.MsgzCreateOptions) msgz.Puller
	NewPublisher func(ctx context.Context, createOpts msgz.MsgzCreateOptions) msgz.Publisher
}

var msgzservices = map[string]*MsgzService{
	"pubsub": {
		NewPuller:    pubsubmsgz.NewPuller,
		NewPublisher: pubsubmsgz.NewPublisher,
	},
	"file": {
		NewPuller:    filemsgz.NewPuller,
		NewPublisher: filemsgz.NewPublisher,
	},
}

func NewPuller(ctx context.Context, createOpts msgz.MsgzCreateOptions) (msgz.Puller, error) {
	ret := msgzservices[createOpts.Strategy]
	if ret == nil {
		return nil, fmt.Errorf("unknown strategy %s", createOpts.Strategy)
	}
	return ret.NewPuller(ctx, createOpts), nil
}

func NewPublisher(ctx context.Context, createOpts msgz.MsgzCreateOptions) (msgz.Publisher, error) {
	ret := msgzservices[createOpts.Strategy]
	if ret == nil {
		return nil, fmt.Errorf("unknown strategy %s", createOpts.Strategy)
	}
	return ret.NewPublisher(ctx, createOpts), nil
}

func RegisterMsgz(name string, service *MsgzService) io.Closer {
	old := msgzservices[name]
	closer := func() {
		msgzservices[name] = old
	}
	msgzservices[name] = service
	return ioz.CloserV(closer)
}
