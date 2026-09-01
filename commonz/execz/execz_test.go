package execz_test

import (
	"context"
	"testing"
	"time"

	"github.com/infinity6-ai/gox/commonz/execz"
	"github.com/stretchr/testify/assert"
	"go.code.infinity6.ai/platform/util"
	"go.code.infinity6.ai/platform/util/filez"
	"go.code.infinity6.ai/platform/util/promise"
	"go.code.infinity6.ai/platform/util/syncz/timerz"
)

func TestUnitBasic(t *testing.T) {
	ctx := context.Background()
	file := filez.CreateTempFile("myfile", nil)
	defer filez.Remove(file)
	cmd := execz.New(ctx, "bash", "-xc", "echo aaa > \"$1\"", "--", file)
	assert.NoError(t, cmd.Run())
	assert.Equal(t, "aaa\n", filez.ReadFileString(file, 10))
	assert.Error(t, cmd.Run())
}

func TestUnitWait(t *testing.T) {
	ctx := context.Background()
	file := filez.CreateTempFile("myfile", nil)
	defer filez.Remove(file)
	cmd := execz.New(ctx, "bash", "-xc", "echo before > \"$1\" && sleep 0.2 && echo after > \"$1\"", "--", file)
	go cmd.Run()
	assert.NoError(t, cmd.Wait())
	assert.Equal(t, "after\n", filez.ReadFileString(file, 10))
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

	util.Sleep(50)
	runPromise := promise.Async(ctx, func() (any, error) {
		err := cmd.Run()
		return nil, err
	})

	timerz.DelayFor(ctx, 250*time.Millisecond, cmd.Kill)
	assert.Equal(t, "before\n", filez.ReadFileString(file, 10))

	err := cmd.Wait()
	assert.Error(t, err)
	assert.ErrorIs(t, err, cmd.Error())
	assert.Equal(t, "before\n", filez.ReadFileString(file, 10))

	_, err = waitPromise.Get()
	assert.Error(t, err)
	assert.ErrorIs(t, err, cmd.Error())

	_, runError := runPromise.Get()
	assert.Error(t, runError)
	assert.ErrorIs(t, err, runError)
}
