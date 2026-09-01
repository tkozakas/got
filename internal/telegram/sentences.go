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
	defaultSentencesPath = "translations/roulette_sentences.json"
	defaultLang          = "en"
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
	languages map[string][]SentenceGroup
}

func NewSentenceProvider() *SentenceProvider {
	return NewSentenceProviderFromFile(defaultSentencesPath)
}

func NewSentenceProviderFromFile(path string) *SentenceProvider {
	provider := &SentenceProvider{
		languages: make(map[string][]SentenceGroup),
	}

	data, err := os.ReadFile(path)
	if err != nil {
		slog.Warn("Failed to load sentences file, roulette animation disabled", "path", path, "error", err)
		return provider
	}

	var multiLang map[string]SentencesFile
	if err := json.Unmarshal(data, &multiLang); err != nil {
		slog.Error("Failed to parse sentences file", "path", path, "error", err)
		return provider
	}

	for lang, file := range multiLang {
		provider.languages[lang] = file.Groups
	}

	slog.Info("Loaded roulette sentences", "languages", len(provider.languages))
	return provider
}

func (p *SentenceProvider) GetRandomGroup(lang string) []string {
	groups := p.languages[lang]
	if len(groups) == 0 {
		groups = p.languages[defaultLang]
	}
	if len(groups) == 0 {
		return nil
	}
	group := groups[rand.Intn(len(groups))]
	return group.Sentences
}

func (p *SentenceProvider) HasSentences() bool {
	for _, groups := range p.languages {
		if len(groups) > 0 {
			return true
		}
	}
	return false
}

func randomDelay() time.Duration {
	return minDelay + time.Duration(rand.Int63n(int64(maxDelay-minDelay)))
}

func RandomDelay() time.Duration {
	return randomDelay()
}

func (p *SentenceProvider) SendSequence(client *Client, chatID int64, lang, alias, winnerName, fallbackMsg string) error {
	sentences := p.GetRandomGroup(lang)
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
