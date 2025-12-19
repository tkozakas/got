package telegram

import (
	"context"
	"net/http"
	"sync/atomic"
	"testing"
	"time"
)

func TestTypingIndicatorStartSendsImmediately(t *testing.T) {
	var callCount atomic.Int32

	server := newTestServerWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		callCount.Add(1)
		w.WriteHeader(http.StatusOK)
	})

	client := newTestClient(server.URL)
	indicator := NewTypingIndicator(client, testChatID)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	indicator.Start(ctx, ActionTyping)
	time.Sleep(50 * time.Millisecond)
	indicator.Stop()

	if callCount.Load() < 1 {
		t.Error("expected at least one chat action to be sent immediately")
	}
}

func TestTypingIndicatorStopsOnStop(t *testing.T) {
	var callCount atomic.Int32

	server := newTestServerWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		callCount.Add(1)
		w.WriteHeader(http.StatusOK)
	})

	client := newTestClient(server.URL)
	indicator := NewTypingIndicator(client, testChatID)

	ctx := context.Background()
	indicator.Start(ctx, ActionTyping)
	time.Sleep(50 * time.Millisecond)
	indicator.Stop()

	countAfterStop := callCount.Load()
	time.Sleep(100 * time.Millisecond)

	if callCount.Load() != countAfterStop {
		t.Error("typing indicator continued after Stop() was called")
	}
}

func TestTypingIndicatorStopsOnContextCancel(t *testing.T) {
	var callCount atomic.Int32

	server := newTestServerWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		callCount.Add(1)
		w.WriteHeader(http.StatusOK)
	})

	client := newTestClient(server.URL)
	indicator := NewTypingIndicator(client, testChatID)

	ctx, cancel := context.WithCancel(context.Background())
	indicator.Start(ctx, ActionTyping)
	time.Sleep(50 * time.Millisecond)
	cancel()

	countAfterCancel := callCount.Load()
	time.Sleep(100 * time.Millisecond)

	if callCount.Load() != countAfterCancel {
		t.Error("typing indicator continued after context was cancelled")
	}
}

func TestNewTypingIndicator(t *testing.T) {
	client := &Client{}
	indicator := NewTypingIndicator(client, testChatID)

	if indicator.client != client {
		t.Error("client not set correctly")
	}

	if indicator.chatID != testChatID {
		t.Errorf("chatID = %d, want %d", indicator.chatID, testChatID)
	}

	if indicator.done == nil {
		t.Error("done channel not initialized")
	}
}
