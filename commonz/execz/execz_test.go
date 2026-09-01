package execz_test

import (
	"context"
	"testing"
	"time"

	"github.com/infinity6-ai/gox/commonz/execz"
	"github.com/infinity6-ai/gox/commonz/filez"
	"github.com/infinity6-ai/gox/commonz/syncz/promise"
	"github.com/infinity6-ai/gox/commonz/syncz/timerz"
	"github.com/stretchr/testify/assert"
)

func TestUnitBasic(t *testing.T) {
	ctx := context.Background()
	file := filez.CreateTempFile("myfile", nil)
	defer filez.Remove(file)
	cmd := execz.New(ctx, "bash", "-xc", "echo aaa > \"$1\"", "--", file)
	assert.NoError(t, cmd.Run())
	assert.Equal(t, "aaa\n", filez.ReadFile(file, 10).String())
	assert.Error(t, cmd.Run())
}

func TestUnitWait(t *testing.T) {
	ctx := context.Background()
	file := filez.CreateTempFile("myfile", nil)
	defer filez.Remove(file)
	cmd := execz.New(ctx, "bash", "-xc", "echo before > \"$1\" && sleep 0.2 && echo after > \"$1\"", "--", file)
	go cmd.Run()
	assert.NoError(t, cmd.Wait())
	assert.Equal(t, "after\n", filez.ReadFile(file, 10).String())
}

func TestUnitKillSimple(t *testing.T) {
	ctx := context.Background()
	file := filez.CreateTempFile("myfile", nil)
	defer filez.Remove(file)
	cmd := execz.New(ctx, "bash", "-c", "echo before > \"$1\" && sleep 2 && echo after > \"$1\"", "--", file)
	waitPromise := promise.Async(ctx, func() (any, error) {
		err := cmd.Wait()
		return nil, err
	})

	time.Sleep(50 * time.Millisecond)
	runPromise := promise.Async(ctx, func() (any, error) {
		err := cmd.Run()
		return nil, err
	})

	timerz.DelayFor(ctx, 250*time.Millisecond, cmd.Kill)
	assert.Equal(t, "before\n", filez.ReadFile(file, 10).String())

	err := cmd.Wait()
	assert.Error(t, err)
	assert.ErrorIs(t, err, cmd.Error())
	assert.Equal(t, "before\n", filez.ReadFile(file, 10).String())

	_, err = waitPromise.Get()
	assert.Error(t, err)
	assert.ErrorIs(t, err, cmd.Error())

	_, runError := runPromise.Get()
	assert.Error(t, runError)
	assert.ErrorIs(t, err, runError)
}
