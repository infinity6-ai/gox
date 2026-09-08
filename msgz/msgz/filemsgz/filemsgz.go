package filemsgz

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/infinity6-ai/gox/commonz/constraintz/blobz"
	"github.com/infinity6-ai/gox/commonz/datez"
	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/filez"
	"github.com/infinity6-ai/gox/commonz/idgen"
	"github.com/infinity6-ai/gox/commonz/jsonz"
	"github.com/infinity6-ai/gox/commonz/pathz"
	"github.com/infinity6-ai/gox/commonz/urlz"
	"github.com/infinity6-ai/gox/commonz/validation/checker"
	"github.com/infinity6-ai/gox/commonz/zipz/gzipz"
	"github.com/infinity6-ai/gox/fsz/fsz"
	"github.com/infinity6-ai/gox/msgz/msgz"
)

type MessageStore struct {
	basedir *pathz.Path
	temp    bool

	lock         sync.Mutex
	shutdownOnce sync.Once
}

func (me *MessageStore) Close() error {
	me.Shutdown()
	return nil
}

func NewTemporaryMessageStore(ctx context.Context) *MessageStore {
	basedir := pathz.MustParse(filez.CreateTempDir("filemsgz"))
	return &MessageStore{
		basedir: basedir,
		temp:    true,
	}
}

func NewMessageStore(ctx context.Context, basedir *pathz.Path) *MessageStore {
	checker.NotNil(basedir, "basedir")
	return &MessageStore{
		basedir: basedir,
		temp:    false,
	}
}

func (me *MessageStore) Basedir() *pathz.Path {
	return me.basedir
}

func (ms *MessageStore) basedirUrl() *urlz.Url {
	return urlz.MustParse("file://" + ms.basedir.String())
}

func (me *MessageStore) Shutdown() {
	me.shutdownOnce.Do(func() {
		me.lock.Lock()
		defer me.lock.Unlock()
		if me.temp && me.basedir != nil {
			errorz.Check(fsz.RmTree(context.Background(), me.basedirUrl()))
		}
	})
}

func (me *MessageStore) Publish(ctx context.Context, topic string, msgs *msgz.Messages) {
	me.lock.Lock()
	defer me.lock.Unlock()
	checker.NotNil(me.basedir, "basedir")
	for _, msg := range msgs.Messages {
		checker.StrEmpty(msg.Id, "msg.ID")
		checker.NotEmpty(msg.Payload, "payload")
		msg.Id = fmt.Sprintf("%s-%s_%s", topic, datez.NowTZ(), idgen.Hex())
		msg.PublishedAt = datez.NowMilliInt()
		mm := msgz.ManagedMessage{
			AckId:   fmt.Sprintf("fileackid-%s", msg.Id),
			Message: msg,
		}
		dst := me.basedirUrl().MustJoinPathString("created", "topic", topic, fmt.Sprintf("%s.json.gz", msg.Id))
		fsz.Upload(ctx, dst, nil, bytes.NewReader(gzipz.MustGzip(jsonz.MustFormat(mm).Bytes())))
	}
}

func (me *MessageStore) Pull(ctx context.Context, sub string, limit int, opts msgz.PullOptions) *msgz.ManagedMessages {
	me.lock.Lock()
	defer me.lock.Unlock()

	srcDir := me.basedirUrl().MustJoinPathString("created", "topic", sub)
	dstDir := me.basedirUrl().MustJoinPathString("fetched", "topic", sub)
	ls, err := fsz.Ls(ctx, srcDir)
	errorz.Check(err)
	defer ls.Close()

	messages := []*msgz.ManagedMessage{}
	page, err := ls.Paginate(ctx, limit)
	errorz.Check(err)

	for _, file := range page {
		srcPath := file.Url
		dstPath := dstDir.MustJoinPathString(file.Url.Path.Base())

		errorz.Check(fsz.Move(ctx, srcPath, dstPath))

		mm := readFile(ctx, dstPath)
		messages = append(messages, mm)
	}

	return &msgz.ManagedMessages{Messages: messages}
}

func readFile(ctx context.Context, f *urlz.Url) *msgz.ManagedMessage {
	var data blobz.Blob
	err := fsz.Download(ctx, f, func(found bool, headers http.Header, reader io.Reader) error {
		if found {
			data = filez.ReadAllLimited(reader, 4096)
		}
		return nil
	})
	errorz.Check(err)
	data = gzipz.MustGunzip(data.Bytes())
	mm := jsonz.MustParse(data.Bytes(), &msgz.ManagedMessage{})
	return mm
}

func parseAckId(id string) (string, string) {
	if !strings.HasPrefix(id, "fileackid-") {
		panic("wrong id")
	}
	id = id[10:]
	idx := strings.LastIndex(id, "-")
	checker.Greater(idx, 0, "idx")
	topic := id[:idx]
	checker.StrNotEmpty(topic, "ret")
	uid := id[idx+1:]
	checker.StrNotEmpty(uid, "uid")
	return topic, id
}

func (me *MessageStore) Ack(ctx context.Context, ids *msgz.Ids) {
	me.lock.Lock()
	defer me.lock.Unlock()
	for _, id := range ids.Ids {
		topic, uid := parseAckId(id)
		dst := me.basedirUrl().MustJoinPathString("fetched", "topic", topic, fmt.Sprintf("%s.json.gz", uid))
		errorz.Check(fsz.Delete(ctx, dst))
	}
}

func (me *MessageStore) Nack(ctx context.Context, ids *msgz.Ids) {
	me.lock.Lock()
	defer me.lock.Unlock()
	me.internalNack(ctx, ids)
}

func (me *MessageStore) internalNack(ctx context.Context, ids *msgz.Ids) {
	for _, id := range ids.Ids {
		topic, uid := parseAckId(id)
		src := me.basedirUrl().MustJoinPathString("fetched", "topic", topic, fmt.Sprintf("%s.json.gz", uid))
		dst := me.basedirUrl().MustJoinPathString("created", "topic", topic, fmt.Sprintf("%s.json.gz", uid))
		errorz.Check(fsz.Move(ctx, src, dst))
	}
}

func (me *MessageStore) NackAll(ctx context.Context, topic string) {
	me.lock.Lock()
	defer me.lock.Unlock()
	srcDir := me.basedirUrl().MustJoinPathString("fetched", "topic", topic)

	p, err := fsz.Ls(ctx, srcDir)
	errorz.Check(err)
	defer p.Close()

	for {
		srcs, err := p.Paginate(ctx, 1000)
		errorz.Check(err)
		if len(srcs) == 0 {
			return
		}
		ids := &msgz.Ids{}
		for _, name := range srcs {
			msg := readFile(ctx, name.Url)
			ids.Ids = append(ids.Ids, msg.AckId)
		}
		me.internalNack(ctx, ids)
	}
}

func New(ctx context.Context, createOpts msgz.MsgzCreateOptions) *MessageStore {
	if createOpts.BaseDir == nil {
		return NewTemporaryMessageStore(ctx)
	}
	return NewMessageStore(ctx, createOpts.BaseDir)
}

func NewPublisher(ctx context.Context, createOpts msgz.MsgzCreateOptions) msgz.Publisher {
	return New(ctx, createOpts)
}

func NewPuller(ctx context.Context, createOpts msgz.MsgzCreateOptions) msgz.Puller {
	return New(ctx, createOpts)
}
