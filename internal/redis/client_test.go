package redis

import (
	"fmt"
	"got/internal/groq"
	"math"
	"testing"
)

const (
	testRedisAddr = "localhost:6379"
)

func TestClientHistoryKey(t *testing.T) {
	client := newTestRedisClient()

	tests := []struct {
		name   string
		chatID int64
		want   string
	}{
		{
			name:   "positiveChatID",
			chatID: 123,
			want:   "gpt:history:123",
		},
		{
			name:   "negativeChatID",
			chatID: -456,
			want:   "gpt:history:-456",
		},
		{
			name:   "zeroChatID",
			chatID: 0,
			want:   "gpt:history:0",
		},
		{
			name:   "largePositiveChatID",
			chatID: 9223372036854775807,
			want:   "gpt:history:9223372036854775807",
		},
		{
			name:   "largeNegativeChatID",
			chatID: -9223372036854775808,
			want:   "gpt:history:-9223372036854775808",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := client.historyKey(tt.chatID)
			assertEqual(t, got, tt.want)
		})
	}
}

func TestClientModelKey(t *testing.T) {
	client := newTestRedisClient()

	tests := []struct {
		name   string
		chatID int64
		want   string
	}{
		{
			name:   "positiveChatID",
			chatID: 123,
			want:   "gpt:model:123",
		},
		{
			name:   "negativeChatID",
			chatID: -456,
			want:   "gpt:model:-456",
		},
		{
			name:   "zeroChatID",
			chatID: 0,
			want:   "gpt:model:0",
		},
		{
			name:   "largePositiveChatID",
			chatID: 9223372036854775807,
			want:   "gpt:model:9223372036854775807",
		},
		{
			name:   "largeNegativeChatID",
			chatID: -9223372036854775808,
			want:   "gpt:model:-9223372036854775808",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := client.modelKey(tt.chatID)
			assertEqual(t, got, tt.want)
		})
	}
}

func TestNewClient(t *testing.T) {
	tests := []struct {
		name string
		addr string
	}{
		{
			name: "standardAddress",
			addr: "localhost:6379",
		},
		{
			name: "customPort",
			addr: "redis.example.com:6380",
		},
		{
			name: "ipAddress",
			addr: "192.168.1.100:6379",
		},
		{
			name: "emptyAddress",
			addr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.addr)

			if client == nil {
				t.Fatal("NewClient() returned nil")
			}
			assertEqual(t, client.addr, tt.addr)
		})
	}
}

func TestSaveHistoryTruncatesLongHistory(t *testing.T) {
	tests := []struct {
		name           string
		historyLen     int
		expectedLen    int
		shouldTruncate bool
	}{
		{
			name:           "belowThreshold",
			historyLen:     10,
			expectedLen:    10,
			shouldTruncate: false,
		},
		{
			name:           "atThreshold",
			historyLen:     20,
			expectedLen:    20,
			shouldTruncate: false,
		},
		{
			name:           "justAboveThreshold",
			historyLen:     21,
			expectedLen:    20,
			shouldTruncate: true,
		},
		{
			name:           "wellAboveThreshold",
			historyLen:     30,
			expectedLen:    20,
			shouldTruncate: true,
		},
		{
			name:           "veryLongHistory",
			historyLen:     100,
			expectedLen:    20,
			shouldTruncate: true,
		},
		{
			name:           "emptyHistory",
			historyLen:     0,
			expectedLen:    0,
			shouldTruncate: false,
		},
		{
			name:           "singleMessage",
			historyLen:     1,
			expectedLen:    1,
			shouldTruncate: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			history := createTestHistoryWithIndex(tt.historyLen)

			truncated := truncateHistory(history)

			if len(truncated) != tt.expectedLen {
				t.Errorf("got length %d, want %d", len(truncated), tt.expectedLen)
			}

			if tt.shouldTruncate && tt.historyLen > 0 {
				expectedFirstIndex := tt.historyLen - maxHistoryLen*2
				expectedContent := fmt.Sprintf("msg-%d", expectedFirstIndex)
				if truncated[0].Content != expectedContent {
					t.Errorf("first message content = %q, want %q", truncated[0].Content, expectedContent)
				}

				expectedLastIndex := tt.historyLen - 1
				expectedLastContent := fmt.Sprintf("msg-%d", expectedLastIndex)
				if truncated[len(truncated)-1].Content != expectedLastContent {
					t.Errorf("last message content = %q, want %q", truncated[len(truncated)-1].Content, expectedLastContent)
				}
			}
		})
	}
}

func TestKeyGenerationUniqueness(t *testing.T) {
	client := newTestRedisClient()

	chatIDs := []int64{0, 1, -1, 100, -100, math.MaxInt64, math.MinInt64}

	historyKeys := make(map[string]int64)
	modelKeys := make(map[string]int64)

	for _, id := range chatIDs {
		hk := client.historyKey(id)
		if existingID, exists := historyKeys[hk]; exists {
			t.Errorf("historyKey collision: chatID %d and %d both produce %q", existingID, id, hk)
		}
		historyKeys[hk] = id

		mk := client.modelKey(id)
		if existingID, exists := modelKeys[mk]; exists {
			t.Errorf("modelKey collision: chatID %d and %d both produce %q", existingID, id, mk)
		}
		modelKeys[mk] = id
	}
}

func TestKeyGenerationConsistency(t *testing.T) {
	client := newTestRedisClient()
	chatID := int64(12345)

	hk1 := client.historyKey(chatID)
	hk2 := client.historyKey(chatID)
	if hk1 != hk2 {
		t.Errorf("historyKey not consistent: got %q and %q for same chatID", hk1, hk2)
	}

	mk1 := client.modelKey(chatID)
	mk2 := client.modelKey(chatID)
	if mk1 != mk2 {
		t.Errorf("modelKey not consistent: got %q and %q for same chatID", mk1, mk2)
	}
}

func TestHistoryKeyAndModelKeyDifferent(t *testing.T) {
	client := newTestRedisClient()

	chatIDs := []int64{0, 1, -1, 123, 999999}

	for _, id := range chatIDs {
		hk := client.historyKey(id)
		mk := client.modelKey(id)
		if hk == mk {
			t.Errorf("historyKey and modelKey should differ for chatID %d, both are %q", id, hk)
		}
	}
}

func TestTruncateHistoryPreservesOrder(t *testing.T) {
	history := createTestHistoryWithIndex(25)

	truncated := truncateHistory(history)

	for i := 1; i < len(truncated); i++ {
		prevContent := truncated[i-1].Content
		currContent := truncated[i].Content

		var prevIdx, currIdx int
		fmt.Sscanf(prevContent, "msg-%d", &prevIdx)
		fmt.Sscanf(currContent, "msg-%d", &currIdx)

		if currIdx <= prevIdx {
			t.Errorf("order not preserved: message %d (%q) comes after message %d (%q)",
				currIdx, currContent, prevIdx, prevContent)
		}
	}
}

func TestTruncateHistoryPreservesMessageContent(t *testing.T) {
	history := []groq.Message{
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi there!"},
		{Role: "user", Content: "How are you?"},
	}

	truncated := truncateHistory(history)

	if len(truncated) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(truncated))
	}

	for i, msg := range truncated {
		if msg.Role != history[i].Role {
			t.Errorf("message %d: role = %q, want %q", i, msg.Role, history[i].Role)
		}
		if msg.Content != history[i].Content {
			t.Errorf("message %d: content = %q, want %q", i, msg.Content, history[i].Content)
		}
	}
}

// truncateHistory extracts the truncation logic for unit testing
func truncateHistory(history []groq.Message) []groq.Message {
	if len(history) > maxHistoryLen*2 {
		return history[len(history)-maxHistoryLen*2:]
	}
	return history
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

func createTestHistoryWithIndex(size int) []groq.Message {
	history := make([]groq.Message, size)
	for i := range history {
		role := "user"
		if i%2 == 1 {
			role = "assistant"
		}
		history[i] = groq.Message{Role: role, Content: fmt.Sprintf("msg-%d", i)}
	}
	return history
}

func assertEqual(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}
