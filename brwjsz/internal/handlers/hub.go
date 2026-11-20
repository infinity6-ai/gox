package handlers

import (
	"context"
	"fmt"
	"sync"

	"github.com/infinity6-ai/gox/brwjsz/brwclient"
)

type controllerConn struct {
	cmdChan chan *brwclient.Command
	done    chan struct{}
}

// Hub coordinates the active browser controller connection, command dispatching, and reply handling.
type Hub struct {
	mu          sync.Mutex
	conn        *controllerConn
	readyNotify chan struct{}
	readyOnce   sync.Once
	stopNotify  chan struct{}
	stopOnce    sync.Once
	pending     map[string]chan *brwclient.Reply
}

// NewHub creates a new Hub instance.
func NewHub() *Hub {
	return &Hub{
		readyNotify: make(chan struct{}),
		stopNotify:  make(chan struct{}),
		pending:     make(map[string]chan *brwclient.Reply),
	}
}

// RequestStop signals that the server and browser should stop.
func (h *Hub) RequestStop() {
	h.stopOnce.Do(func() {
		close(h.stopNotify)
	})
}

// WaitStop returns a channel that is closed when a stop is requested.
func (h *Hub) WaitStop() <-chan struct{} {
	return h.stopNotify
}

// RegisterController registers an active SSE connection for the browser controller.
func (h *Hub) RegisterController() (*controllerConn, func()) {
	h.mu.Lock()
	conn := &controllerConn{
		cmdChan: make(chan *brwclient.Command, 32),
		done:    make(chan struct{}),
	}
	h.conn = conn
	h.readyOnce.Do(func() {
		close(h.readyNotify)
	})
	h.mu.Unlock()

	cleanup := func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		close(conn.done)
		if h.conn == conn {
			h.conn = nil
		}
	}
	return conn, cleanup
}

// WaitReady blocks until the browser controller has connected via SSE.
func (h *Hub) WaitReady(ctx context.Context) error {
	select {
	case <-h.readyNotify:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("timeout waiting for browser controller to connect: %w", ctx.Err())
	}
}

// Dispatch sends a command to the browser controller and awaits the corresponding reply.
func (h *Hub) Dispatch(ctx context.Context, cmd *brwclient.Command) (*brwclient.Reply, error) {
	h.mu.Lock()
	conn := h.conn
	if conn == nil {
		h.mu.Unlock()
		return nil, fmt.Errorf("browser controller is not connected")
	}
	replyChan := make(chan *brwclient.Reply, 1)
	h.pending[cmd.ID] = replyChan
	h.mu.Unlock()

	defer func() {
		h.mu.Lock()
		delete(h.pending, cmd.ID)
		h.mu.Unlock()
	}()

	select {
	case conn.cmdChan <- cmd:
	case <-conn.done:
		return nil, fmt.Errorf("browser disconnected before command was sent")
	case <-ctx.Done():
		return nil, fmt.Errorf("failed to send command to browser: %w", ctx.Err())
	}

	select {
	case reply := <-replyChan:
		return reply, nil
	case <-conn.done:
		return nil, fmt.Errorf("browser disconnected while waiting for reply")
	case <-ctx.Done():
		return nil, fmt.Errorf("command execution timed out or canceled: %w", ctx.Err())
	}
}

// HandleReply routes a reply received from the browser controller to its pending dispatch caller.
func (h *Hub) HandleReply(reply *brwclient.Reply) bool {
	h.mu.Lock()
	replyChan, ok := h.pending[reply.ID]
	h.mu.Unlock()
	if ok {
		select {
		case replyChan <- reply:
			return true
		default:
			return false
		}
	}
	return false
}
