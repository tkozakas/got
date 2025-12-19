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
	client := NewClient()

	if client.httpClient == nil {
		t.Error("httpClient should not be nil")
	}
	if client.voice != defaultVoice {
		t.Errorf("expected default voice %q, got %q", defaultVoice, client.voice)
	}
}

func TestGenerateSpeechServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient()
	client.httpClient = server.Client()

	if client.httpClient == nil {
		t.Error("httpClient should be set")
	}
}

func TestGenerateSpeechEmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient()
	client.httpClient = server.Client()

	if client.voice != defaultVoice {
		t.Errorf("expected default voice, got %q", client.voice)
	}
}

func TestGenerateSpeechContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := NewClient()

	_, err := client.GenerateSpeech(ctx, "test")
	if err == nil {
		t.Error("expected error for cancelled context")
	}
}
