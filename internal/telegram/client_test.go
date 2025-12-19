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

func TestNewClient(t *testing.T) {
	client := NewClient("my-token")

	if client == nil {
		t.Fatal("NewClient() returned nil")
	}
	if client.token != "my-token" {
		t.Errorf("token = %q, want %q", client.token, "my-token")
	}
	if client.httpClient == nil {
		t.Error("httpClient should not be nil")
	}
	if client.baseURL != "https://api.telegram.org/botmy-token" {
		t.Errorf("baseURL = %q, want %q", client.baseURL, "https://api.telegram.org/botmy-token")
	}
}

func TestClientSendChatAction(t *testing.T) {
	server := newTestServerWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		payload := decodeJSONPayload(t, r)
		assertPayloadInt(t, payload, "chat_id", testChatID)
		assertPayloadString(t, payload, "action", "typing")
		w.WriteHeader(http.StatusOK)
	})

	client := newTestClient(server.URL)
	err := client.SendChatAction(testChatID, "typing")

	assertNoError(t, err)
}

func TestClientSendMediaGroup(t *testing.T) {
	server := newTestServerWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		payload := decodeJSONPayload(t, r)
		assertPayloadInt(t, payload, "chat_id", testChatID)

		media, ok := payload["media"].([]any)
		if !ok {
			t.Error("media should be an array")
			return
		}
		if len(media) != 2 {
			t.Errorf("media length = %d, want 2", len(media))
		}
		w.WriteHeader(http.StatusOK)
	})

	client := newTestClient(server.URL)
	media := []InputMediaPhoto{
		{Type: "photo", Media: "https://example.com/1.jpg", Caption: "Photo 1"},
		{Type: "photo", Media: "https://example.com/2.jpg", Caption: "Photo 2"},
	}
	err := client.SendMediaGroup(testChatID, media)

	assertNoError(t, err)
}

func TestClientSendAnimation(t *testing.T) {
	server := newTestServerWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		payload := decodeJSONPayload(t, r)
		assertPayloadInt(t, payload, "chat_id", testChatID)
		assertPayloadString(t, payload, "animation", "https://example.com/anim.gif")
		assertPayloadString(t, payload, "caption", "Funny gif")
		w.WriteHeader(http.StatusOK)
	})

	client := newTestClient(server.URL)
	err := client.SendAnimation(testChatID, "https://example.com/anim.gif", "Funny gif")

	assertNoError(t, err)
}

func TestClientSetMyCommands(t *testing.T) {
	server := newTestServerWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		payload := decodeJSONPayload(t, r)
		commands, ok := payload["commands"].([]any)
		if !ok {
			t.Error("commands should be an array")
			return
		}
		if len(commands) != 2 {
			t.Errorf("commands length = %d, want 2", len(commands))
		}
		w.WriteHeader(http.StatusOK)
	})

	client := newTestClient(server.URL)
	commands := []BotCommand{
		{Command: "start", Description: "Start the bot"},
		{Command: "help", Description: "Show help"},
	}
	err := client.SetMyCommands(commands)

	assertNoError(t, err)
}

func TestClientSendVoice(t *testing.T) {
	server := newTestServerWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			t.Fatalf("failed to parse multipart form: %v", err)
		}

		chatID := r.FormValue("chat_id")
		if chatID != "123" {
			t.Errorf("chat_id = %q, want %q", chatID, "123")
		}

		file, header, err := r.FormFile("voice")
		if err != nil {
			t.Fatalf("failed to get voice file: %v", err)
		}
		defer file.Close()

		if header.Filename != "audio.mp3" {
			t.Errorf("filename = %q, want %q", header.Filename, "audio.mp3")
		}

		w.WriteHeader(http.StatusOK)
	})

	client := newTestClient(server.URL)
	err := client.SendVoice(testChatID, []byte("audio data"), "audio.mp3")

	assertNoError(t, err)
}

func TestClientSendDocument(t *testing.T) {
	server := newTestServerWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			t.Fatalf("failed to parse multipart form: %v", err)
		}

		chatID := r.FormValue("chat_id")
		if chatID != "123" {
			t.Errorf("chat_id = %q, want %q", chatID, "123")
		}

		caption := r.FormValue("caption")
		if caption != "My Document" {
			t.Errorf("caption = %q, want %q", caption, "My Document")
		}

		file, header, err := r.FormFile("document")
		if err != nil {
			t.Fatalf("failed to get document file: %v", err)
		}
		defer file.Close()

		if header.Filename != "file.txt" {
			t.Errorf("filename = %q, want %q", header.Filename, "file.txt")
		}

		w.WriteHeader(http.StatusOK)
	})

	client := newTestClient(server.URL)
	err := client.SendDocument(testChatID, []byte("file content"), "file.txt", "My Document")

	assertNoError(t, err)
}

func TestClientSendVoiceError(t *testing.T) {
	server := newTestServerWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	client := newTestClient(server.URL)
	err := client.SendVoice(testChatID, []byte("audio"), "audio.mp3")

	if err == nil {
		t.Error("expected error for server error response")
	}
}
