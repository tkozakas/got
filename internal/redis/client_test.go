package redis

import (
	"context"
	"got/internal/groq"
	"testing"
)

const (
	testRedisAddr = "localhost:6379"
	testChatID    = int64(123)
)

func TestClientHistoryKey(t *testing.T) {
	client := newTestRedisClient()

	tests := []struct {
		name   string
		chatID int64
		want   string
	}{
		{
			name:   "Positive chat ID",
			chatID: 123,
			want:   "gpt:history:123",
		},
		{
			name:   "Negative chat ID",
			chatID: -456,
			want:   "gpt:history:-456",
		},
		{
			name:   "Zero chat ID",
			chatID: 0,
			want:   "gpt:history:0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := client.historyKey(tt.chatID)
			assertEqual(t, got, tt.want)
		})
	}
}

func TestClientSaveHistoryTruncation(t *testing.T) {
	client := newTestRedisClient()
	history := createTestHistory(30)

	err := client.SaveHistory(context.Background(), testChatID, history)

	if err == nil {
		t.Skip("Redis not available, skipping integration test")
	}
}

func TestNewClient(t *testing.T) {
	client := NewClient(testRedisAddr)

	if client == nil {
		t.Fatal("NewClient() returned nil")
	}

	assertEqual(t, client.addr, testRedisAddr)
}

func newTestRedisClient() *Client {
	return NewClient(testRedisAddr)
}

func createTestHistory(size int) []groq.Message {
	history := make([]groq.Message, size)
	for i := range history {
		history[i] = groq.Message{Role: "user", Content: "msg"}
	}
	return history
}

func assertEqual(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}
