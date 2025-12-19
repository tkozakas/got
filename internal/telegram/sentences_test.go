package telegram

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewSentenceProviderFromFile_ValidFile(t *testing.T) {
	content := `{
		"groups": [
			{"id": "group1", "sentences": ["Hello", "World"]},
			{"id": "group2", "sentences": ["Foo", "Bar", "Baz"]}
		]
	}`
	tmpFile := createTempJSONFile(t, content)
	defer os.Remove(tmpFile)

	provider := NewSentenceProviderFromFile(tmpFile)

	if !provider.HasSentences() {
		t.Error("expected provider to have sentences")
	}
	if len(provider.groups) != 2 {
		t.Errorf("expected 2 groups, got %d", len(provider.groups))
	}
}

func TestNewSentenceProviderFromFile_MissingFile(t *testing.T) {
	provider := NewSentenceProviderFromFile("/nonexistent/path/file.json")

	if provider.HasSentences() {
		t.Error("expected provider to have no sentences for missing file")
	}
}

func TestNewSentenceProviderFromFile_InvalidJSON(t *testing.T) {
	tmpFile := createTempJSONFile(t, "not valid json")
	defer os.Remove(tmpFile)

	provider := NewSentenceProviderFromFile(tmpFile)

	if provider.HasSentences() {
		t.Error("expected provider to have no sentences for invalid JSON")
	}
}

func TestNewSentenceProviderFromFile_EmptyGroups(t *testing.T) {
	content := `{"groups": []}`
	tmpFile := createTempJSONFile(t, content)
	defer os.Remove(tmpFile)

	provider := NewSentenceProviderFromFile(tmpFile)

	if provider.HasSentences() {
		t.Error("expected provider to have no sentences for empty groups")
	}
}

func TestSentenceProvider_GetRandomGroup(t *testing.T) {
	content := `{
		"groups": [
			{"id": "only_group", "sentences": ["One", "Two", "Three"]}
		]
	}`
	tmpFile := createTempJSONFile(t, content)
	defer os.Remove(tmpFile)

	provider := NewSentenceProviderFromFile(tmpFile)
	group := provider.GetRandomGroup()

	if group == nil {
		t.Fatal("expected non-nil group")
	}
	if len(group) != 3 {
		t.Errorf("expected 3 sentences, got %d", len(group))
	}
	if group[0] != "One" || group[1] != "Two" || group[2] != "Three" {
		t.Errorf("unexpected sentences: %v", group)
	}
}

func TestSentenceProvider_GetRandomGroup_Empty(t *testing.T) {
	provider := &SentenceProvider{groups: []SentenceGroup{}}
	group := provider.GetRandomGroup()

	if group != nil {
		t.Errorf("expected nil for empty provider, got %v", group)
	}
}

func TestSentenceProvider_HasSentences(t *testing.T) {
	tests := []struct {
		name   string
		groups []SentenceGroup
		want   bool
	}{
		{
			name:   "emptyGroups",
			groups: []SentenceGroup{},
			want:   false,
		},
		{
			name:   "nilGroups",
			groups: nil,
			want:   false,
		},
		{
			name: "hasGroups",
			groups: []SentenceGroup{
				{ID: "test", Sentences: []string{"Hello"}},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &SentenceProvider{groups: tt.groups}
			got := provider.HasSentences()
			if got != tt.want {
				t.Errorf("HasSentences() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRandomDelay(t *testing.T) {
	for i := 0; i < 100; i++ {
		delay := randomDelay()
		if delay < minDelay || delay > maxDelay {
			t.Errorf("randomDelay() = %v, want between %v and %v", delay, minDelay, maxDelay)
		}
	}
}

func createTempJSONFile(t *testing.T, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test_sentences.json")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	return tmpFile
}
