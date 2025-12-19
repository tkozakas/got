package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	defaultTimeout = 30 * time.Second
	getUpdatesCMD  = "/getUpdates"
	sendMessageCMD = "/sendMessage"
)

type Client struct {
	token      string
	httpClient *http.Client
	baseURL    string
}

func NewClient(token string) *Client {
	return &Client{
		token: token,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
		baseURL: "https://api.telegram.org/bot" + token,
	}
}

func (c *Client) GetUpdates(offset int) ([]Update, error) {
	url := fmt.Sprintf("%s%s?offset=%d&timeout=60", c.baseURL, getUpdatesCMD, offset)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return c.parseUpdatesResponse(resp.Body)
}

func (c *Client) SendMessage(chatID int64, text string) error {
	payload := map[string]any{
		"chat_id": chatID,
		"text":    text,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return c.postJSON(sendMessageCMD, data)
}

func (c *Client) parseUpdatesResponse(body io.Reader) ([]Update, error) {
	data, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}

	var apiResp APIResponse
	if err := json.Unmarshal(data, &apiResp); err != nil {
		return nil, err
	}

	if !apiResp.Ok {
		return nil, fmt.Errorf("telegram api error: %s", apiResp.Description)
	}

	return apiResp.Result, nil
}

func (c *Client) postJSON(endpoint string, data []byte) error {
	resp, err := c.httpClient.Post(
		c.baseURL+endpoint,
		"application/json",
		bytes.NewBuffer(data),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send request: %s", resp.Status)
	}

	return nil
}
