package noopz_test

import (
	"testing"

	"github.com/infinity6-ai/gox/noopz/noopz"
)

func TestUnitBasic(t *testing.T) {
	noopz.Noop(nil)
}
