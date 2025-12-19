package telegram

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"strings"
	"time"
)

const (
	defaultSentencesPath = "roulette_sentences.json"
	minDelay             = 700 * time.Millisecond
	maxDelay             = 1200 * time.Millisecond
)

type SentenceGroup struct {
	ID        string   `json:"id"`
	Sentences []string `json:"sentences"`
}

type SentencesFile struct {
	Groups []SentenceGroup `json:"groups"`
}

type SentenceProvider struct {
	groups []SentenceGroup
}

func NewSentenceProvider() *SentenceProvider {
	return NewSentenceProviderFromFile(defaultSentencesPath)
}

func NewSentenceProviderFromFile(path string) *SentenceProvider {
	provider := &SentenceProvider{
		groups: []SentenceGroup{},
	}

	data, err := os.ReadFile(path)
	if err != nil {
		slog.Warn("Failed to load sentences file, roulette animation disabled", "path", path, "error", err)
		return provider
	}

	var file SentencesFile
	if err := json.Unmarshal(data, &file); err != nil {
		slog.Error("Failed to parse sentences file", "path", path, "error", err)
		return provider
	}

	provider.groups = file.Groups
	slog.Info("Loaded roulette sentences", "groups", len(provider.groups))
	return provider
}

func (p *SentenceProvider) GetRandomGroup() []string {
	if len(p.groups) == 0 {
		return nil
	}
	group := p.groups[rand.Intn(len(p.groups))]
	return group.Sentences
}

func (p *SentenceProvider) HasSentences() bool {
	return len(p.groups) > 0
}

func randomDelay() time.Duration {
	return minDelay + time.Duration(rand.Int63n(int64(maxDelay-minDelay)))
}

func RandomDelay() time.Duration {
	return randomDelay()
}

func (p *SentenceProvider) SendSequence(client *Client, chatID int64, alias, winnerName, fallbackMsg string) error {
	sentences := p.GetRandomGroup()
	if len(sentences) == 0 {
		return client.SendMessage(chatID, fallbackMsg)
	}

	for _, sentence := range sentences {
		time.Sleep(randomDelay())
		msg := FormatSentence(sentence, alias, winnerName)
		if err := client.SendMessage(chatID, msg); err != nil {
			return err
		}
	}

	return nil
}

func FormatSentence(template, alias, winner string) string {
	count := strings.Count(template, "%s")
	switch count {
	case 0:
		return template
	case 1:
		return fmt.Sprintf(template, alias)
	default:
		return fmt.Sprintf(template, alias, winner)
	}
}
