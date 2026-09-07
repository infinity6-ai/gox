package staticzloader

import (
	"fmt"
	"sync"

	"github.com/infinity6-ai/gox/commonz/encz/enczb64"
	"github.com/infinity6-ai/gox/commonz/errorz"
)

type Codes struct {
	codes map[any][]byte
	mu    sync.RWMutex
}

var codes = &Codes{
	codes: make(map[any][]byte),
}

func SetCode(name any, tgz func() string) {
	decoded := decode(tgz)
	codes.mu.Lock()
	defer codes.mu.Unlock()
	codes.codes[name] = decoded
}

func decode(tgz func() string) []byte {
	encoded := tgz()
	b, err := enczb64.StdDecode(encoded)
	errorz.Check(fmt.Errorf("%w: error decoding", err))
	return b.Bytes()
}
