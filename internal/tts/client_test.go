package tts

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	client := NewClient()

	if client == nil {
		t.Fatal("NewClient() returned nil")
	}
	if client.voice != defaultVoice {
		t.Errorf("expected default voice %q, got %q", defaultVoice, client.voice)
	}
	if client.httpClient == nil {
		t.Error("httpClient should not be nil")
	}
	if client.baseURL != baseURL {
		t.Errorf("expected baseURL %q, got %q", baseURL, client.baseURL)
	}
}

func TestSetVoice(t *testing.T) {
	client := NewClient()
	client.SetVoice("Amy")

	if client.voice != "Amy" {
		t.Errorf("expected voice 'Amy', got %q", client.voice)
	}
}

func TestVoice(t *testing.T) {
	client := NewClient()

	if client.Voice() != defaultVoice {
		t.Errorf("expected %q, got %q", defaultVoice, client.Voice())
	}

	client.SetVoice("Emma")
	if client.Voice() != "Emma" {
		t.Errorf("expected 'Emma', got %q", client.Voice())
	}
}

func TestGenerateSpeechSuccess(t *testing.T) {
	tests := []struct {
		name          string
		text          string
		voice         string
		responseBody  []byte
		expectedVoice string
		expectedText  string
	}{
		{
			name:          "basic text with default voice",
			text:          "Hello world",
			voice:         "",
			responseBody:  []byte("fake audio data"),
			expectedVoice: defaultVoice,
			expectedText:  "Hello world",
		},
		{
			name:          "text with custom voice",
			text:          "Testing TTS",
			voice:         "Amy",
			responseBody:  []byte{0x00, 0x01, 0x02, 0x03},
			expectedVoice: "Amy",
			expectedText:  "Testing TTS",
		},
		{
			name:          "text with special characters",
			text:          "Hello & goodbye!",
			voice:         "",
			responseBody:  []byte("audio bytes"),
			expectedVoice: defaultVoice,
			expectedText:  "Hello & goodbye!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("expected GET request, got %s", r.Method)
				}

				voice := r.URL.Query().Get("voice")
				if voice != tt.expectedVoice {
					t.Errorf("expected voice %q, got %q", tt.expectedVoice, voice)
				}

				text := r.URL.Query().Get("text")
				if text != tt.expectedText {
					t.Errorf("expected text %q, got %q", tt.expectedText, text)
				}

				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(tt.responseBody)
			}))
			defer server.Close()

			client := NewClient()
			client.baseURL = server.URL

			if tt.voice != "" {
				client.SetVoice(tt.voice)
			}

			result, err := client.GenerateSpeech(context.Background(), tt.text)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(result) != len(tt.responseBody) {
				t.Errorf("expected %d bytes, got %d bytes", len(tt.responseBody), len(result))
			}

			for i, b := range result {
				if b != tt.responseBody[i] {
					t.Errorf("byte mismatch at index %d: expected %v, got %v", i, tt.responseBody[i], b)
					break
				}
			}
		})
	}
}

func TestGenerateSpeechServerError(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		expectedErrMsg string
	}{
		{
			name:           "internal server error",
			statusCode:     http.StatusInternalServerError,
			expectedErrMsg: "tts api error: 500 Internal Server Error",
		},
		{
			name:           "bad request",
			statusCode:     http.StatusBadRequest,
			expectedErrMsg: "tts api error: 400 Bad Request",
		},
		{
			name:           "service unavailable",
			statusCode:     http.StatusServiceUnavailable,
			expectedErrMsg: "tts api error: 503 Service Unavailable",
		},
		{
			name:           "not found",
			statusCode:     http.StatusNotFound,
			expectedErrMsg: "tts api error: 404 Not Found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			client := NewClient()
			client.baseURL = server.URL

			_, err := client.GenerateSpeech(context.Background(), "test text")
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if err.Error() != tt.expectedErrMsg {
				t.Errorf("expected error %q, got %q", tt.expectedErrMsg, err.Error())
			}
		})
	}
}

func TestGenerateSpeechEmptyResponse(t *testing.T) {
	tests := []struct {
		name         string
		responseBody []byte
	}{
		{
			name:         "nil response body",
			responseBody: nil,
		},
		{
			name:         "empty byte slice",
			responseBody: []byte{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				if tt.responseBody != nil {
					_, _ = w.Write(tt.responseBody)
				}
			}))
			defer server.Close()

			client := NewClient()
			client.baseURL = server.URL

			_, err := client.GenerateSpeech(context.Background(), "test text")
			if err == nil {
				t.Fatal("expected error for empty response, got nil")
			}

			expectedErr := "empty response from tts api"
			if err.Error() != expectedErr {
				t.Errorf("expected error %q, got %q", expectedErr, err.Error())
			}
		})
	}
}

func TestGenerateSpeechContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("audio data"))
	}))
	defer server.Close()

	client := NewClient()
	client.baseURL = server.URL

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.GenerateSpeech(ctx, "test")
	if err == nil {
		t.Error("expected error for cancelled context")
	}
}
