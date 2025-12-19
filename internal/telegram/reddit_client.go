package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"got/internal/app/model"
	"net/http"
	"time"
)

const (
	memeAPIURL = "https://meme-api.com/gimme/%s/%d"
)

func (h *BotHandlers) fetchMemes(ctx context.Context, subreddit string, count int) ([]model.RedditMeme, error) {
	url := fmt.Sprintf(memeAPIURL, subreddit, count)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch memes: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("api returned status: %d", resp.StatusCode)
	}

	var response model.RedditResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Memes, nil
}
