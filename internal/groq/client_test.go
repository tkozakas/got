package groq

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	testAPIKey = "test-key"
)

func TestClientChat(t *testing.T) {
	tests := []struct {
		name       string
		response   Response
		statusCode int
		wantErr    bool
		want       string
	}{
		{
			name:       "Successful response",
			response:   newSuccessResponse("Hello!"),
			statusCode: http.StatusOK,
			wantErr:    false,
			want:       "Hello!",
		},
		{
			name:       "Empty choices",
			response:   Response{ID: "test-id", Choices: []Choice{}},
			statusCode: http.StatusOK,
			wantErr:    true,
		},
		{
			name:       "API error response",
			response:   Response{Error: &Error{Message: "rate limit exceeded", Type: "rate_limit"}},
			statusCode: http.StatusOK,
			wantErr:    true,
		},
		{
			name:       "HTTP error",
			response:   Response{},
			statusCode: http.StatusInternalServerError,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := newGroqTestServer(t, tt.response, tt.statusCode)
			client := newTestGroqClient(server.URL)

			got, err := client.Chat(context.Background(), "test prompt", nil)

			assertError(t, err, tt.wantErr)

			if !tt.wantErr && got != tt.want {
				t.Errorf("Client.Chat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClientChatWithHistory(t *testing.T) {
	var receivedMessages []Message

	server := newTestServerWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assertGroqHeaders(t, r)
		var req Request
		json.NewDecoder(r.Body).Decode(&req)
		receivedMessages = req.Messages
		json.NewEncoder(w).Encode(newSuccessResponse("response"))
	})

	history := []Message{
		{Role: "user", Content: "previous question"},
		{Role: "assistant", Content: "previous answer"},
	}

	client := newTestGroqClient(server.URL)
	_, err := client.Chat(context.Background(), "new question", history)

	assertNoError(t, err)

	if len(receivedMessages) != 4 {
		t.Errorf("want 4 messages (system + 2 history + user), got %d", len(receivedMessages))
	}

	if receivedMessages[0].Role != "system" {
		t.Errorf("first message should be system, got %s", receivedMessages[0].Role)
	}

	if receivedMessages[3].Content != "new question" {
		t.Errorf("last message should be new question, got %s", receivedMessages[3].Content)
	}
}

func TestClientListModels(t *testing.T) {
	client := NewClient(testAPIKey)
	models := client.ListModels()

	if len(models) == 0 {
		t.Error("ListModels() returned empty list")
	}

	if !containsModel(models, defaultModel) {
		t.Error("default model not found in ListModels()")
	}
}

func TestClientSetModel(t *testing.T) {
	var receivedModel string

	server := newTestServerWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		var req Request
		json.NewDecoder(r.Body).Decode(&req)
		receivedModel = req.Model
		json.NewEncoder(w).Encode(newSuccessResponse("ok"))
	})

	client := newTestGroqClient(server.URL)
	client.SetModel("mixtral-8x7b-32768")

	_, err := client.Chat(context.Background(), "test", nil)
	assertNoError(t, err)

	if receivedModel != "mixtral-8x7b-32768" {
		t.Errorf("want model mixtral-8x7b-32768, got %s", receivedModel)
	}
}

func TestClientBuildMessages(t *testing.T) {
	client := NewClient(testAPIKey)

	history := []Message{
		{Role: "user", Content: "hi"},
		{Role: "assistant", Content: "hello"},
	}

	messages := client.buildMessages("new prompt", history)

	if len(messages) != 4 {
		t.Fatalf("want 4 messages, got %d", len(messages))
	}

	if messages[0].Role != "system" {
		t.Errorf("first message role = %s, want system", messages[0].Role)
	}

	if messages[1].Content != "hi" {
		t.Errorf("second message content = %s, want hi", messages[1].Content)
	}

	if messages[3].Content != "new prompt" {
		t.Errorf("last message content = %s, want new prompt", messages[3].Content)
	}
}

func TestClientParseResponse(t *testing.T) {
	client := NewClient(testAPIKey)

	tests := []struct {
		name    string
		data    string
		want    string
		wantErr bool
	}{
		{
			name:    "Valid response",
			data:    `{"id":"1","choices":[{"message":{"role":"assistant","content":"hello"}}]}`,
			want:    "hello",
			wantErr: false,
		},
		{
			name:    "Empty choices",
			data:    `{"id":"1","choices":[]}`,
			wantErr: true,
		},
		{
			name:    "Error in response",
			data:    `{"error":{"message":"bad request","type":"invalid_request"}}`,
			wantErr: true,
		},
		{
			name:    "Invalid JSON",
			data:    `{invalid`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := client.parseResponse([]byte(tt.data))

			assertError(t, err, tt.wantErr)

			if !tt.wantErr && got != tt.want {
				t.Errorf("parseResponse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func newTestGroqClient(serverURL string) *Client {
	client := NewClient(testAPIKey)
	client.baseURL = serverURL
	return client
}

func newGroqTestServer(t *testing.T, response Response, statusCode int) *httptest.Server {
	t.Helper()
	return newTestServerWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assertGroqHeaders(t, r)
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(response)
	})
}

func newTestServerWithHandler(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return server
}

func newSuccessResponse(content string) Response {
	return Response{
		ID:      "test-id",
		Choices: []Choice{{Message: Message{Role: "assistant", Content: content}}},
	}
}

func assertGroqHeaders(t *testing.T, r *http.Request) {
	t.Helper()
	if r.Header.Get("Authorization") != "Bearer "+testAPIKey {
		t.Error("missing or invalid authorization header")
	}
	if r.Header.Get("Content-Type") != "application/json" {
		t.Error("missing content-type header")
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

func containsModel(models []string, target string) bool {
	for _, m := range models {
		if m == target {
			return true
		}
	}
	return false
}

func TestClientFetchModels(t *testing.T) {
	tests := []struct {
		name       string
		response   ModelsResponse
		statusCode int
		wantErr    bool
		wantLen    int
	}{
		{
			name: "successfulResponse",
			response: ModelsResponse{
				Data: []ModelInfo{
					{ID: "llama-3.3-70b-versatile"},
					{ID: "mixtral-8x7b-32768"},
				},
			},
			statusCode: http.StatusOK,
			wantErr:    false,
			wantLen:    2,
		},
		{
			name: "filtersWhisperModels",
			response: ModelsResponse{
				Data: []ModelInfo{
					{ID: "llama-3.3-70b-versatile"},
					{ID: "whisper-large-v3"},
					{ID: "whisper-large-v3-turbo"},
				},
			},
			statusCode: http.StatusOK,
			wantErr:    false,
			wantLen:    1,
		},
		{
			name:       "httpError",
			response:   ModelsResponse{},
			statusCode: http.StatusInternalServerError,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := newTestServerWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("Authorization") != "Bearer "+testAPIKey {
					t.Error("missing authorization header")
				}
				w.WriteHeader(tt.statusCode)
				json.NewEncoder(w).Encode(tt.response)
			})

			client := newTestGroqClientWithModelsURL(server.URL)
			models, err := client.FetchModels(context.Background())

			assertError(t, err, tt.wantErr)

			if !tt.wantErr && len(models) != tt.wantLen {
				t.Errorf("got %d models, want %d", len(models), tt.wantLen)
			}
		})
	}
}

func TestClientFetchModelsSorted(t *testing.T) {
	response := ModelsResponse{
		Data: []ModelInfo{
			{ID: "zebra-model"},
			{ID: "alpha-model"},
			{ID: "beta-model"},
		},
	}

	server := newTestServerWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(response)
	})

	client := newTestGroqClientWithModelsURL(server.URL)
	models, err := client.FetchModels(context.Background())

	assertNoError(t, err)

	if models[0] != "alpha-model" {
		t.Errorf("models not sorted, first = %s, want alpha-model", models[0])
	}
}

func newTestGroqClientWithModelsURL(serverURL string) *Client {
	client := NewClient(testAPIKey)
	client.modelsURL = serverURL
	return client
}
