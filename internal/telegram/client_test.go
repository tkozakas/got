package telegram

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	testToken  = "test-token"
	testChatID = int64(123)
)

func TestClientSendMessage(t *testing.T) {
	tests := []struct {
		name       string
		chatID     int64
		text       string
		statusCode int
		wantErr    bool
	}{
		{
			name:       "Successful send",
			chatID:     testChatID,
			text:       "Hello",
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "Server error",
			chatID:     testChatID,
			text:       "Hello",
			statusCode: http.StatusInternalServerError,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := newTestServerWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
				payload := decodeJSONPayload(t, r)
				assertPayloadInt(t, payload, "chat_id", tt.chatID)
				assertPayloadString(t, payload, "text", tt.text)
				w.WriteHeader(tt.statusCode)
			})

			client := newTestClient(server.URL)
			err := client.SendMessage(tt.chatID, tt.text)

			assertError(t, err, tt.wantErr)
		})
	}
}

func TestClientSendPhoto(t *testing.T) {
	server := newTestServerWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		payload := decodeJSONPayload(t, r)
		assertPayloadInt(t, payload, "chat_id", testChatID)
		assertPayloadString(t, payload, "photo", "https://example.com/photo.jpg")
		assertPayloadString(t, payload, "caption", "caption text")
		w.WriteHeader(http.StatusOK)
	})

	client := newTestClient(server.URL)
	err := client.SendPhoto(testChatID, "https://example.com/photo.jpg", "caption text")

	assertNoError(t, err)
}

func TestClientSendSticker(t *testing.T) {
	server := newTestServerWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		payload := decodeJSONPayload(t, r)
		assertPayloadInt(t, payload, "chat_id", testChatID)
		assertPayloadString(t, payload, "sticker", "sticker-file-id")
		w.WriteHeader(http.StatusOK)
	})

	client := newTestClient(server.URL)
	err := client.SendSticker(testChatID, "sticker-file-id")

	assertNoError(t, err)
}

func TestClientGetUpdates(t *testing.T) {
	updates := []Update{
		{UpdateID: 1, Message: &Message{Text: "hello"}},
		{UpdateID: 2, Message: &Message{Text: "world"}},
	}

	server := newTestServerWithJSON(t, APIResponse{Ok: true, Result: updates})

	client := newTestClient(server.URL)
	got, err := client.GetUpdates(0)

	assertNoError(t, err)

	if len(got) != 2 {
		t.Errorf("got %d updates, want 2", len(got))
	}

	if got[0].UpdateID != 1 {
		t.Errorf("first update ID = %d, want 1", got[0].UpdateID)
	}
}

func TestClientGetUpdatesError(t *testing.T) {
	server := newTestServerWithJSON(t, APIResponse{Ok: false, Description: "Unauthorized"})

	client := newTestClient(server.URL)
	_, err := client.GetUpdates(0)

	if err == nil {
		t.Error("expected error for unauthorized request")
	}
}

func newTestClient(serverURL string) *Client {
	return &Client{
		token:      testToken,
		httpClient: http.DefaultClient,
		baseURL:    serverURL,
	}
}

func newTestServerWithHandler(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return server
}

func newTestServerWithJSON(t *testing.T, response any) *httptest.Server {
	t.Helper()
	return newTestServerWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(response)
	})
}

func decodeJSONPayload(t *testing.T, r *http.Request) map[string]any {
	t.Helper()
	var payload map[string]any
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode payload: %v", err)
	}
	return payload
}

func assertPayloadInt(t *testing.T, payload map[string]any, key string, want int64) {
	t.Helper()
	got, ok := payload[key].(float64)
	if !ok {
		t.Errorf("payload[%s] not found or not a number", key)
		return
	}
	if int64(got) != want {
		t.Errorf("payload[%s] = %v, want %v", key, int64(got), want)
	}
}

func assertPayloadString(t *testing.T, payload map[string]any, key string, want string) {
	t.Helper()
	got, ok := payload[key].(string)
	if !ok {
		t.Errorf("payload[%s] not found or not a string", key)
		return
	}
	if got != want {
		t.Errorf("payload[%s] = %v, want %v", key, got, want)
	}
}

func assertError(t *testing.T, err error, wantErr bool) {
	t.Helper()
	if (err != nil) != wantErr {
		t.Errorf("error = %v, wantErr %v", err, wantErr)
	}
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
