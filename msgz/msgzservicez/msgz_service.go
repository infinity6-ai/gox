package msgzservicez

import (
	"context"
	"fmt"
	"io"

	"github.com/infinity6-ai/gox/commonz/ioz"
	"github.com/infinity6-ai/gox/msgz/msgz"
	"github.com/infinity6-ai/gox/msgz/msgz/pubsubmsgz"
)

type MsgzService struct {
	NewPuller    func(ctx context.Context, projectId string) msgz.Puller
	NewPublisher func(ctx context.Context, projectId string) msgz.Publisher
}

var msgzservices = map[string]*MsgzService{
	"pubsub": {
		NewPuller:    pubsubmsgz.NewPuller,
		NewPublisher: pubsubmsgz.NewPublisher,
	},
}

func NewPuller(ctx context.Context, strategy string, projectId string) (msgz.Puller, error) {
	ret := msgzservices[strategy]
	if ret == nil {
		return nil, fmt.Errorf("unknown strategy %s", strategy)
	}
	return ret.NewPuller(ctx, projectId), nil
}

func NewPublisher(ctx context.Context, strategy string, projectId string) (msgz.Publisher, error) {
	ret := msgzservices[strategy]
	if ret == nil {
		return nil, fmt.Errorf("unknown strategy %s", strategy)
	}
	return ret.NewPublisher(ctx, projectId), nil
}

func RegisterMsgzService(name string, service *MsgzService) io.Closer {
	old := msgzservices[name]
	closer := func() {
		msgzservices[name] = old
	}
	msgzservices[name] = service
	return ioz.CloserV(closer)
}
