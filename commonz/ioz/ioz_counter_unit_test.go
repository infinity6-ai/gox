package ioz_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/infinity6-ai/gox/commonz/ioz"
	"github.com/stretchr/testify/require"
)

func TestUnitReaderCounterCount(t *testing.T) {
	rc := &ioz.ReaderCounter{Reader: strings.NewReader("test")}
	require.Equal(t, uint64(0), rc.Count())
}

func TestUnitWriterCounterCount(t *testing.T) {
	var buf bytes.Buffer
	wc := &ioz.WriterCounter{Writer: &buf}
	require.Equal(t, uint64(0), wc.Count())
}
