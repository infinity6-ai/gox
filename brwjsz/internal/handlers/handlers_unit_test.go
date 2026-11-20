package handlers_test

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/infinity6-ai/gox/brwjsz/brwclient"
	"github.com/infinity6-ai/gox/brwjsz/internal/handlers"
	"github.com/infinity6-ai/gox/httpz/httpzserver"
	"github.com/stretchr/testify/require"
)

func TestUnitHandlersStaticServing(t *testing.T) {
	ctx := t.Context()
	srv := httpzserver.New(ctx, httpzserver.Options{LocalAddress: "localhost:0"})
	defer srv.Close()

	hub := handlers.NewHub()
	handlers.AddHandlers(ctx, srv, hub)
	srv.Listen()
	srv.Start()

	type testScenario struct {
		name         string
		path         string
		wantContains string
	}

	check := func(t *testing.T, s testScenario) {
		t.Helper()
		reqUrl := srv.Base().MustJoinPathString(s.path).String()
		resp, err := http.Get(reqUrl)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Contains(t, string(body), s.wantContains)
	}

	t.Run("Serves controller.html", func(t *testing.T) {
		check(t, testScenario{
			name:         "controller page",
			path:         handlers.ControllerPath,
			wantContains: "brwjsz controller",
		})
	})

	t.Run("Serves index.js", func(t *testing.T) {
		check(t, testScenario{
			name:         "index javascript",
			path:         handlers.BasePath + "/brwjsz/index.js",
			wantContains: "window.brwjsz",
		})
	})
}

func TestUnitHubDispatchAndReply(t *testing.T) {
	ctx := t.Context()
	srv := httpzserver.New(ctx, httpzserver.Options{LocalAddress: "localhost:0"})
	defer srv.Close()

	hub := handlers.NewHub()
	handlers.AddHandlers(ctx, srv, hub)
	srv.Listen()
	srv.Start()

	// 1. Simulate Controller connecting via SSE
	eventsUrl := srv.Base().MustJoinPathString(handlers.EventsPath).String()
	sseResp, err := http.Get(eventsUrl)
	require.NoError(t, err)
	defer sseResp.Body.Close()

	// WaitReady should succeed immediately since controller is connected
	waitCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	err = hub.WaitReady(waitCtx)
	require.NoError(t, err)

	// 2. Dispatch a command in background
	cmd := &brwclient.Command{
		ID:     "test-cmd-1",
		Action: "proc",
		Fn:     "console.log",
	}

	replyReceived := make(chan *brwclient.Reply, 1)
	go func() {
		reply, dispatchErr := hub.Dispatch(ctx, cmd)
		require.NoError(t, dispatchErr)
		replyReceived <- reply
	}()

	// 3. Controller replies to /api/brwjs/reply
	replyPayload := `{"id":"test-cmd-1","ok":true}`
	replyUrl := srv.Base().MustJoinPathString(handlers.ReplyPath).String()
	postResp, err := http.Post(replyUrl, "application/json", strings.NewReader(replyPayload))
	require.NoError(t, err)
	defer postResp.Body.Close()
	require.Equal(t, http.StatusOK, postResp.StatusCode)

	// 4. Verify Dispatch received reply
	select {
	case reply := <-replyReceived:
		require.Equal(t, "test-cmd-1", reply.ID)
		require.True(t, reply.Ok)
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for reply")
	}
}
