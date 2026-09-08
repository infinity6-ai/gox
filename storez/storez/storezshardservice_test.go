package storez_test

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/infinity6-ai/gox/commonz/jsonz"
	"github.com/infinity6-ai/gox/commonz/validation/checker"
	"github.com/infinity6-ai/gox/storez/storezfile"
	"github.com/infinity6-ai/gox/storez/storezshardservice"
	"github.com/stretchr/testify/assert"
)

type MyEntity struct {
	Id        string    `json:"id"`
	Counter   int       `json:"counter"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (me *MyEntity) GetId() string {
	return me.Id
}

func (me *MyEntity) SetId(id string) {
	me.Id = id
}

func (me *MyEntity) GetUpdatedAt() time.Time {
	return me.UpdatedAt
}

func (me *MyEntity) SetUpdatedAt(updatedAt time.Time) {
	me.UpdatedAt = updatedAt
}

func (me *MyEntity) Merge(other *MyEntity) {
	checker.Equal(me.Id, other.Id, "ids must be equal")
	me.Counter += other.Counter
}

func TestUnitShardService(t *testing.T) {
	ctx := context.Background()
	projectId, storeReverter := storezfile.InitEmulator(ctx, "")
	defer storeReverter.Close()
	client := Open(ctx, projectId, "mytest")
	defer client.Close()

	fragmentor := storezshardservice.CreateShardServiceWithResolver(
		client,
		"counter",
		10,
		storezshardservice.DefaultResolver[*MyEntity](10),
		func(entity1, entity2 *MyEntity) *MyEntity {
			entity1.Merge(entity2)
			return entity1
		})

	entityA := &MyEntity{Id: "myid", Counter: 3}
	entityB := &MyEntity{Id: "myid", Counter: 5}

	fragmentor.Put(ctx, entityA)
	fragmentor.Put(ctx, entityB)

	loaded := fragmentor.Load(ctx, "myid")
	assert.NotNil(t, loaded)
	assert.Equal(t, 8, loaded.Counter)
}

type SessionEntity struct {
	Id        string    `json:"id"`
	Events    string    `json:"events"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Event struct {
	CreatedAt int64  `json:"created_at"`
	ItemId    string `json:"item_id"`
}

func (m *SessionEntity) GetId() string {
	return m.Id
}

func (m *SessionEntity) SetId(id string) {
	m.Id = id
}

func (me *SessionEntity) GetUpdatedAt() time.Time {
	return me.UpdatedAt
}

func (me *SessionEntity) SetUpdatedAt(updatedAt time.Time) {
	me.UpdatedAt = updatedAt
}

func (m *SessionEntity) Merge(other *SessionEntity) {
	checker.Equal(m.Id, other.Id, "ids must be equal")

	e1 := []Event{}
	jsonz.MustParse(m.Events, &e1)

	e2 := []Event{}
	jsonz.MustParse(other.Events, &e2)

	e1 = append(e1, e2...)

	slices.SortFunc(e1, func(a Event, b Event) int {
		return -int(a.CreatedAt - b.CreatedAt)
	})

	m.Events = jsonz.MustFormat(e1[:2]).String()
}

func TestUnitShardServiceSession(t *testing.T) {
	ctx := context.Background()
	projectId, storeReverter := storezfile.InitEmulator(ctx, "")
	defer storeReverter.Close()
	client := Open(ctx, projectId, "mytest")
	defer client.Close()

	fragmentor := storezshardservice.CreateShardServiceWithResolver(
		client,
		"sessions",
		10,
		storezshardservice.DefaultResolver[*SessionEntity](10),
		func(entity1, entity2 *SessionEntity) *SessionEntity {
			entity1.Merge(entity2)
			return entity1
		})

	entityA := &SessionEntity{Id: "myid", Events: jsonz.MustFormat([]Event{
		{
			ItemId:    "1",
			CreatedAt: 1,
		},
		{
			ItemId:    "3",
			CreatedAt: 3,
		},
	}).String()}
	entityB := &SessionEntity{Id: "myid", Events: jsonz.MustFormat([]Event{
		{
			ItemId:    "4",
			CreatedAt: 4,
		},
		{
			ItemId:    "2",
			CreatedAt: 2,
		},
	}).String()}

	fragmentor.Put(ctx, entityA)
	fragmentor.Put(ctx, entityB)

	loaded := fragmentor.Load(ctx, "myid")
	assert.NotNil(t, loaded)

	events := []Event{}
	jsonz.MustParse(loaded.Events, &events)

	expected := []Event{
		{
			ItemId:    "4",
			CreatedAt: 4,
		},
		{
			ItemId:    "3",
			CreatedAt: 3,
		},
	}
	assert.Equal(t, expected, events)

}
