package groq

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	baseURL        = "https://api.groq.com/openai/v1/chat/completions"
	defaultTimeout = 30 * time.Second
	defaultModel   = "llama-3.3-70b-versatile"
	roleSystem     = "system"
	roleUser       = "user"
	systemPrompt   = "You are a helpful assistant in a Telegram chat. Keep responses concise and friendly."
)

type Client struct {
	apiKey     string
	httpClient *http.Client
	model      string
	baseURL    string
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Request struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Response struct {
	ID      string   `json:"id"`
	Choices []Choice `json:"choices"`
	Error   *Error   `json:"error,omitempty"`
}

type Choice struct {
	Message Message `json:"message"`
}

type Error struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
		model:   defaultModel,
		baseURL: baseURL,
	}
}

func (c *Client) Chat(ctx context.Context, prompt string, history []Message) (string, error) {
	messages := c.buildMessages(prompt, history)

	reqBody := Request{
		Model:    c.model,
		Messages: messages,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.doRequest(ctx, data)
	if err != nil {
		return "", err
	}

	return c.parseResponse(resp)
}

func (c *Client) ListModels() []string {
	return []string{
		"llama-3.3-70b-versatile",
		"llama-3.1-8b-instant",
		"mixtral-8x7b-32768",
		"gemma2-9b-it",
	}
}

func (c *Client) SetModel(model string) {
	c.model = model
}

func (c *Client) buildMessages(prompt string, history []Message) []Message {
	messages := make([]Message, 0, len(history)+2)
	messages = append(messages, Message{Role: roleSystem, Content: systemPrompt})
	messages = append(messages, history...)
	messages = append(messages, Message{Role: roleUser, Content: prompt})
	return messages
}

func (c *Client) doRequest(ctx context.Context, data []byte) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("api error: %s", string(body))
	}

	return body, nil
}

func (c *Client) parseResponse(data []byte) (string, error) {
	var resp Response
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if resp.Error != nil {
		return "", fmt.Errorf("groq error: %s", resp.Error.Message)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response choices")
	}

	return resp.Choices[0].Message.Content, nil
}
