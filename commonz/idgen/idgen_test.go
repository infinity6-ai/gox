package idgen_test

import (
	"fmt"
	"testing"

	"github.com/infinity6-ai/gox/commonz/idgen"
	"github.com/stretchr/testify/require"
)

func TestUnitIdGen(t *testing.T) {
	require.Regexp(t, "^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$", idgen.String())
	require.Regexp(t, "^[A-Za-z0-9_\\-]{22}$", idgen.B64())
	require.Regexp(t, "^[0123456789abcdefghijklmnopqrstuv]{26}$", idgen.B32())
	require.Regexp(t, "^[A-f0-9]{32}$", idgen.Hex())

	require.Equal(t, "2849bb91-4780-5e13-aee7-45e164e36bd6", idgen.FromString("aaa"))

	require.Equal(t, "2d98e360-22f7-5f2c-93d8-5ef732c5a4b6", idgen.FromString(fmt.Sprintf("%s.%s", "2024-04", "3")))
}
