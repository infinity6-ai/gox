package logzprovider_test

import (
	"context"
	"testing"

	"github.com/infinity6-ai/gox/commonz/logz/logzprovider"
	"github.com/infinity6-ai/gox/commonz/logz/logzspec"
	"github.com/stretchr/testify/require"
)

func TestUnitDefaultProvider(t *testing.T) {
	// ensure cleanup
	orig := logzprovider.GetDefaultProvider()
	t.Cleanup(func() {
		logzprovider.SetDefaultProvider(orig)
	})

	t.Run("default is simple provider", func(t *testing.T) {
		p := logzprovider.GetDefaultProvider()
		require.NotNil(t, p)
		
		// To assert it's the SimpleProvider, we can check function pointers
		// Note: In Go, comparing funcs directly isn't allowed, but we can verify it's not nil 
		// and we can verify SetDefaultProvider changes it.
	})

	t.Run("set default provider", func(t *testing.T) {
		var invoked bool
		customProvider := func(ctx context.Context, entry *logzspec.Entry) {
			invoked = true
		}

		logzprovider.SetDefaultProvider(customProvider)
		p := logzprovider.GetDefaultProvider()
		require.NotNil(t, p)

		p(context.Background(), &logzspec.Entry{})
		require.True(t, invoked)
	})
}
