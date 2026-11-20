package handlers

import (
	"context"
	"fmt"
	"sync"

	"github.com/infinity6-ai/gox/brwjsz/brwclient"
)

type tabConnection struct {
	tab     int
	cmdChan chan *brwclient.Command
	done    chan struct{}
}

// Hub coordinates active browser tab connections, command dispatching, and reply handling.
type Hub struct {
	mu          sync.Mutex
	tabs        map[int]*tabConnection
	readyTabs   map[int]bool
	readyNotify chan struct{}
	pending     map[string]chan *brwclient.Reply
}

// NewHub creates a new Hub instance.
func NewHub() *Hub {
	return &Hub{
		tabs:        make(map[int]*tabConnection),
		readyTabs:   make(map[int]bool),
		readyNotify: make(chan struct{}, 1),
		pending:     make(map[string]chan *brwclient.Reply),
	}
}

// RegisterTab registers an active SSE connection for a tab and returns its connection struct.
func (h *Hub) RegisterTab(tab int) (*tabConnection, func()) {
	h.mu.Lock()
	conn := &tabConnection{
		tab:     tab,
		cmdChan: make(chan *brwclient.Command, 32),
		done:    make(chan struct{}),
	}
	h.tabs[tab] = conn
	h.readyTabs[tab] = true

	// Notify any waiters that a tab has checked in
	select {
	case h.readyNotify <- struct{}{}:
	default:
	}
	h.mu.Unlock()

	cleanup := func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		close(conn.done)
		delete(h.tabs, tab)
	}
	return conn, cleanup
}

// WaitReady blocks until all tabs from 0 to tabCount-1 have registered their SSE connections.
func (h *Hub) WaitReady(ctx context.Context, tabCount int) error {
	for {
		h.mu.Lock()
		allReady := true
		for i := 0; i < tabCount; i++ {
			if !h.readyTabs[i] {
				allReady = false
				break
			}
		}
		notify := h.readyNotify
		h.mu.Unlock()

		if allReady {
			return nil
		}

		select {
		case <-notify:
			// A new tab connected, check again
		case <-ctx.Done():
			h.mu.Lock()
			var connected []int
			for i := 0; i < tabCount; i++ {
				if h.readyTabs[i] {
					connected = append(connected, i)
				}
			}
			h.mu.Unlock()
			return fmt.Errorf("timeout waiting for %d tabs to connect (connected: %v): %w", tabCount, connected, ctx.Err())
		}
	}
}

// Dispatch sends a command to the specified tab and awaits the corresponding reply.
func (h *Hub) Dispatch(ctx context.Context, tab int, cmd *brwclient.Command) (*brwclient.Reply, error) {
	h.mu.Lock()
	conn, ok := h.tabs[tab]
	if !ok {
		h.mu.Unlock()
		return nil, fmt.Errorf("tab %d is not connected", tab)
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
		return nil, fmt.Errorf("tab %d disconnected before command was sent", tab)
	case <-ctx.Done():
		return nil, fmt.Errorf("failed to send command to tab %d: %w", tab, ctx.Err())
	}

	select {
	case reply := <-replyChan:
		return reply, nil
	case <-conn.done:
		return nil, fmt.Errorf("tab %d disconnected while waiting for reply", tab)
	case <-ctx.Done():
		return nil, fmt.Errorf("command execution timed out or canceled on tab %d: %w", tab, ctx.Err())
	}
}

// HandleReply routes a reply received from a browser controller to its pending dispatch caller.
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
