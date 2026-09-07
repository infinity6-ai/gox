package staticzloader_test

import (
	"testing"

	"github.com/infinity6-ai/gox/commonz/internal/stzfiles"
	"github.com/infinity6-ai/gox/commonz/staticz/staticzloader"
)

func TestUnitBasic(t *testing.T) {
	staticzloader.GetCode(stzfiles.Name)
}
