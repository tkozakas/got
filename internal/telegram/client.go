package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

const (
	defaultTimeout    = 30 * time.Second
	getUpdatesCMD     = "/getUpdates"
	sendMessageCMD    = "/sendMessage"
	sendPhotoCMD      = "/sendPhoto"
	sendStickerCMD    = "/sendSticker"
	sendVoiceCMD      = "/sendVoice"
	sendDocumentCMD   = "/sendDocument"
	sendAnimationCMD  = "/sendAnimation"
	sendMediaGroupCMD = "/sendMediaGroup"
	sendChatActionCMD = "/sendChatAction"
	setMyCommandsCMD  = "/setMyCommands"
	getStickerSetCMD  = "/getStickerSet"
)

type Client struct {
	token      string
	httpClient *http.Client
	baseURL    string
}

type InputMediaPhoto struct {
	Type    string `json:"type"`
	Media   string `json:"media"`
	Caption string `json:"caption,omitempty"`
}

type InputMediaAnimation struct {
	Type    string `json:"type"`
	Media   string `json:"media"`
	Caption string `json:"caption,omitempty"`
}

type BotCommand struct {
	Command     string `json:"command"`
	Description string `json:"description"`
}

type StickerSet struct {
	Name     string           `json:"name"`
	Title    string           `json:"title"`
	Stickers []StickerSetItem `json:"stickers"`
}

type StickerSetItem struct {
	FileID     string `json:"file_id"`
	SetName    string `json:"set_name"`
	IsVideo    bool   `json:"is_video"`
	IsAnimated bool   `json:"is_animated"`
}

type StickerSetResponse struct {
	Ok          bool       `json:"ok"`
	Result      StickerSet `json:"result"`
	Description string     `json:"description,omitempty"`
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
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": "Markdown",
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return c.postJSON(sendMessageCMD, data)
}

func (c *Client) SendPhoto(chatID int64, photoURL string, caption string) error {
	payload := map[string]any{
		"chat_id": chatID,
		"photo":   photoURL,
		"caption": caption,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return c.postJSON(sendPhotoCMD, data)
}

func (c *Client) SendSticker(chatID int64, stickerID string) error {
	payload := map[string]any{
		"chat_id": chatID,
		"sticker": stickerID,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return c.postJSON(sendStickerCMD, data)
}

func (c *Client) SendMediaGroup(chatID int64, media []InputMediaPhoto) error {
	payload := map[string]any{
		"chat_id": chatID,
		"media":   media,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return c.postJSON(sendMediaGroupCMD, data)
}

func (c *Client) SendAnimation(chatID int64, animationURL string, caption string) error {
	payload := map[string]any{
		"chat_id":   chatID,
		"animation": animationURL,
		"caption":   caption,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return c.postJSON(sendAnimationCMD, data)
}

func (c *Client) SendChatAction(chatID int64, action string) error {
	payload := map[string]any{
		"chat_id": chatID,
		"action":  action,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return c.postJSON(sendChatActionCMD, data)
}

func (c *Client) SendVoice(chatID int64, audioData []byte, filename string) error {
	return c.sendMultipartFile(chatID, sendVoiceCMD, "voice", audioData, filename, "")
}

func (c *Client) SendDocument(chatID int64, fileData []byte, filename string, caption string) error {
	return c.sendMultipartFile(chatID, sendDocumentCMD, "document", fileData, filename, caption)
}

func (c *Client) SetMyCommands(commands []BotCommand) error {
	payload := map[string]any{
		"commands": commands,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return c.postJSON(setMyCommandsCMD, data)
}

func (c *Client) GetStickerSet(name string) (*StickerSet, error) {
	url := fmt.Sprintf("%s%s?name=%s", c.baseURL, getStickerSetCMD, name)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResp StickerSetResponse
	if err := json.Unmarshal(data, &apiResp); err != nil {
		return nil, err
	}

	if !apiResp.Ok {
		return nil, fmt.Errorf("sticker set not found: %s", apiResp.Description)
	}

	return &apiResp.Result, nil
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

func (c *Client) sendMultipartFile(chatID int64, endpoint string, fieldName string, fileData []byte, filename string, caption string) error {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	if err := writer.WriteField("chat_id", fmt.Sprintf("%d", chatID)); err != nil {
		return err
	}

	if caption != "" {
		if err := writer.WriteField("caption", caption); err != nil {
			return err
		}
	}

	part, err := writer.CreateFormFile(fieldName, filename)
	if err != nil {
		return err
	}

	if _, err := part.Write(fileData); err != nil {
		return err
	}

	if err := writer.Close(); err != nil {
		return err
	}

	resp, err := c.httpClient.Post(
		c.baseURL+endpoint,
		writer.FormDataContentType(),
		&buf,
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send %s: %s", fieldName, resp.Status)
	}

	return nil
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
