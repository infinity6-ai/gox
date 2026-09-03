package filemsgz

import (
	"context"
	"fmt"
	"path"
	"strings"
	"sync"

	"github.com/infinity6-ai/gox/commonz/datez"
	"github.com/infinity6-ai/gox/commonz/errorz"
	"github.com/infinity6-ai/gox/commonz/filez"
	"github.com/infinity6-ai/gox/commonz/idgen"
	"github.com/infinity6-ai/gox/commonz/jsonz"
	"github.com/infinity6-ai/gox/commonz/validation/checker"
	"github.com/infinity6-ai/gox/commonz/zipz/gzipz"
	"github.com/infinity6-ai/gox/msgz/msgz"
)

type MessageStore struct {
	basedir string
	temp    bool

	lock         sync.Mutex
	shutdownOnce sync.Once
}

func (me *MessageStore) Close() error {
	me.Shutdown()
	return nil
}

func NewTemporaryMessageStore(ctx context.Context) *MessageStore {
	return &MessageStore{
		basedir: filez.CreateTempDir("filemsgz"),
		temp:    true,
	}
}

func NewMessageStore(ctx context.Context, basedir string) *MessageStore {
	checker.StrNotEmpty(basedir, "basedir")
	return &MessageStore{
		basedir: basedir,
		temp:    false,
	}
}

func (me *MessageStore) Basedir() string {
	return me.basedir
}

func (me *MessageStore) Shutdown() {
	me.shutdownOnce.Do(func() {
		me.lock.Lock()
		defer me.lock.Unlock()
		if me.temp && me.basedir != "" {
			errorz.Check(filez.RmTree(me.basedir))
		}
	})
}

func (me *MessageStore) Publish(ctx context.Context, topic string, msgs *msgz.Messages) {
	me.lock.Lock()
	defer me.lock.Unlock()
	checker.StrNotEmpty(me.basedir, "basedir")
	for _, msg := range msgs.Messages {
		checker.StrEmpty(msg.Id, "msg.ID")
		checker.NotEmpty(msg.Payload, "payload")
		msg.Id = fmt.Sprintf("%s-%s_%s", topic, datez.NowTZ(), idgen.Hex())
		msg.PublishedAt = datez.NowMilliInt()
		mm := msgz.ManagedMessage{
			AckId:   fmt.Sprintf("fileackid-%s", msg.Id),
			Message: msg,
		}
		dst := path.Join(me.basedir, "created", "topic", topic, fmt.Sprintf("%s.json.gz", msg.Id))
		filez.WriteFile(dst, gzipz.MustGzip(jsonz.MustFormat(mm).Bytes()))
	}
}

// func parseSub(sub string) (string, string) {
// 	parts := strings.SplitN(sub, "-", 1)
// 	return parts[0], parts[1]
// }

func (me *MessageStore) Pull(ctx context.Context, sub string, limit int, opts msgz.PullOptions) *msgz.ManagedMessages {
	me.lock.Lock()
	defer me.lock.Unlock()

	srcDir := path.Join(me.basedir, "created", "topic", sub)
	dstDir := path.Join(me.basedir, "fetched", "topic", sub)
	filez.CreateParentDirs(path.Join(srcDir, "file"))
	files := filez.DirListLimited(srcDir, ".*", limit)
	messages := []*msgz.ManagedMessage{}
	for _, file := range files {
		srcPath := path.Join(srcDir, file)
		dstPath := path.Join(dstDir, file)

		filez.Move(srcPath, dstPath)
		mm := readFile(dstPath)
		messages = append(messages, mm)
	}

	return &msgz.ManagedMessages{Messages: messages}
}

func readFile(p string) *msgz.ManagedMessage {
	data := filez.ReadFile(p, 4096)
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
		dst := path.Join(me.basedir, "fetched", "topic", topic, fmt.Sprintf("%s.json.gz", uid))
		errorz.Check(filez.Remove(dst))
	}
}

func (me *MessageStore) Nack(ctx context.Context, ids *msgz.Ids) {
	me.lock.Lock()
	defer me.lock.Unlock()
	me.internalNack(ids)
}

func (me *MessageStore) internalNack(ids *msgz.Ids) {
	for _, id := range ids.Ids {
		topic, uid := parseAckId(id)
		src := path.Join(me.basedir, "fetched", "topic", topic, fmt.Sprintf("%s.json.gz", uid))
		dst := path.Join(me.basedir, "created", "topic", topic, fmt.Sprintf("%s.json.gz", uid))
		filez.Move(src, dst)
	}
}

func (me *MessageStore) NackAll(ctx context.Context, topic string) {
	me.lock.Lock()
	defer me.lock.Unlock()
	srcDir := path.Join(me.basedir, "fetched", "topic", topic)
	for {
		srcs := filez.DirListLimited(srcDir, ".*", 1000)
		if len(srcs) == 0 {
			return
		}
		ids := &msgz.Ids{}
		for _, name := range srcs {
			src := path.Join(srcDir, name)
			msg := readFile(src)
			ids.Ids = append(ids.Ids, msg.AckId)
		}
		me.internalNack(ids)
	}
}
