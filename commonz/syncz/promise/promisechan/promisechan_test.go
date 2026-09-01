package promisechan_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/infinity6-ai/gox/commonz/strconvz"
	"github.com/infinity6-ai/gox/commonz/syncz/promise"
	"github.com/infinity6-ai/gox/commonz/syncz/promise/promisechan"
	"github.com/stretchr/testify/assert"
)

func TestUnitBasics(t *testing.T) {
	ctx := context.Background()
	in := make(chan string, 1)
	out, end := promisechan.Async(ctx, in, 1, 10, func(out chan int) {
		for i := range in {
			fmt.Printf("consuming %s\n", i)
			out <- strconvz.MustParseNumber[int](i)
		}
	})
	defer end()
	inPromise := promise.AsyncV(ctx, func() {
		defer close(in)
		for i := 1; i <= 3; i++ {
			fmt.Printf("sending %d\n", i)
			in <- strconv.Itoa(i)
		}
	})
	defer inPromise.WaitDeferrable()

	result := 0
	for o := range out {
		fmt.Printf("received %d\n", o)
		result += o
	}
	assert.Equal(t, 6, result)

	end()
}
