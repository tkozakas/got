package tts

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	baseURL        = "https://api.streamelements.com/kappa/v2/speech"
	defaultTimeout = 30 * time.Second
	defaultVoice   = "Brian"
)

type Client struct {
	httpClient *http.Client
	voice      string
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
		voice: defaultVoice,
	}
}

func (c *Client) GenerateSpeech(ctx context.Context, text string) ([]byte, error) {
	params := url.Values{}
	params.Set("voice", c.voice)
	params.Set("text", text)

	reqURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tts api error: %s", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("empty response from tts api")
	}

	return data, nil
}

func (c *Client) SetVoice(voice string) {
	c.voice = voice
}

func (c *Client) Voice() string {
	return c.voice
}
