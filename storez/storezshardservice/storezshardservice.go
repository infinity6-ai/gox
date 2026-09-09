package storezshardservice

import (
	"context"
	"fmt"
	"math/rand/v2"
	"regexp"
	"time"

	"github.com/infinity6-ai/gox/commonz/datez"
	"github.com/infinity6-ai/gox/commonz/validation/checker"
	"github.com/infinity6-ai/gox/storez/storez"
)

type Fragmentable[T any] interface {
	GetId() string
	SetId(id string)
	GetUpdatedAt() time.Time
	SetUpdatedAt(updatedAt time.Time)
}

type ShardService[T Fragmentable[T]] struct {
	client    *storez.StorezClient
	TableName string
	Blocks    int
	resolver  func(entity T) int
	merger    func(entity1 T, entity2 T) T
}

func DefaultResolver[T Fragmentable[T]](blocks int) func(entity T) int {
	rnd := rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), uint64(time.Now().UnixNano())))
	resolver := func(entity T) int {
		return rnd.IntN(blocks)
	}

	return resolver
}

func CreateShardServiceWithResolver[T Fragmentable[T]](client *storez.StorezClient, table string, blocks int, resolver func(entity T) int, merger func(entity1 T, entity2 T) T) *ShardService[T] {
	return &ShardService[T]{
		client:    client,
		TableName: table,
		Blocks:    blocks,
		resolver:  resolver,
		merger:    merger,
	}
}

func (me *ShardService[T]) Put(ctx context.Context, entity T) {
	block := me.resolver(entity)
	checker.GreaterOrEqual(block, 0, "block")
	checker.Less(block, me.Blocks, "block")
	id := entity.GetId()

	blockId := fmt.Sprintf("%s_%d", id, block)

	storez.Transaction(ctx, me.client, func(client *storez.StorezClient) {
		query := &storez.Query{
			Table: me.TableName,
			Limit: 1,
			Filters: []storez.Filter{
				{
					Field: "id",
					Op:    "=",
					Value: blockId,
				},
			},
		}

		var blockEntity *T = nil
		storez.Paginate(ctx, me.client, query, func(idx uint64, row T) {
			temp := row
			blockEntity = &temp
		})

		entity.SetId(blockId)
		if blockEntity != nil {
			entity = me.merger(entity, *blockEntity)
		}

		entity.SetUpdatedAt(datez.NowTime())

		storez.Put(ctx, me.client, me.TableName, entity)
	})
}

func removeSuffix(id string) string {
	re := regexp.MustCompile(`_[0-9]+$`)
	return re.ReplaceAllString(id, "")
}

func (me *ShardService[T]) Load(ctx context.Context, id string) T {
	query := &storez.Query{
		Table: me.TableName,
		Limit: me.Blocks,
		Filters: []storez.Filter{
			{
				Field: "id",
				Op:    ">=",
				Value: id,
			},
			{
				Field: "id",
				Op:    "<",
				Value: fmt.Sprintf("%s%s", id, "\ufffd"),
			},
		},
	}
	fragments := []T{}
	storez.Paginate(ctx, me.client, query, func(idx uint64, row T) {
		fragments = append(fragments, row)
	})

	if len(fragments) == 0 {
		var zero T
		return zero
	}

	merged := fragments[0]
	merged.SetId(removeSuffix(merged.GetId()))
	for _, frag := range fragments[1:] {
		frag.SetId(removeSuffix(frag.GetId()))
		merged = me.merger(merged, frag)
	}
	return merged
}
