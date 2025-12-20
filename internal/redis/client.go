package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"got/internal/groq"
	"net"
	"time"
)

const (
	defaultTimeout  = 5 * time.Second
	historyTTL      = 24 * time.Hour
	adminSessionTTL = 12 * time.Hour
	maxHistoryLen   = 10
	historyKeyFmt   = "gpt:history:%d"
	modelKeyFmt     = "gpt:model:%d"
	adminKeyFmt     = "admin:session:%d"
	commandSet      = "*3\r\n$3\r\nSET\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n"
	commandGet      = "*2\r\n$3\r\nGET\r\n$%d\r\n%s\r\n"
	commandExpire   = "*3\r\n$6\r\nEXPIRE\r\n$%d\r\n%s\r\n$%d\r\n%d\r\n"
	responseOK      = "+OK"
	responseNil     = "$-1"
	responseBulk    = '$'
)

type Client struct {
	addr string
}

func NewClient(addr string) *Client {
	return &Client{addr: addr}
}

func (c *Client) GetHistory(ctx context.Context, chatID int64) ([]groq.Message, error) {
	key := c.historyKey(chatID)

	data, err := c.get(ctx, key)
	if err != nil {
		return nil, err
	}
	if data == "" {
		return nil, nil
	}

	var history []groq.Message
	if err := json.Unmarshal([]byte(data), &history); err != nil {
		return nil, fmt.Errorf("failed to unmarshal history: %w", err)
	}

	return history, nil
}

func (c *Client) SaveHistory(ctx context.Context, chatID int64, history []groq.Message) error {
	if len(history) > maxHistoryLen*2 {
		history = history[len(history)-maxHistoryLen*2:]
	}

	data, err := json.Marshal(history)
	if err != nil {
		return fmt.Errorf("failed to marshal history: %w", err)
	}

	key := c.historyKey(chatID)
	return c.setWithTTL(ctx, key, string(data), historyTTL)
}

func (c *Client) ClearHistory(ctx context.Context, chatID int64) error {
	key := c.historyKey(chatID)
	return c.set(ctx, key, "[]")
}

func (c *Client) GetModel(ctx context.Context, chatID int64) (string, error) {
	key := c.modelKey(chatID)
	return c.get(ctx, key)
}

func (c *Client) SetModel(ctx context.Context, chatID int64, model string) error {
	key := c.modelKey(chatID)
	return c.set(ctx, key, model)
}

func (c *Client) SetAdminSession(ctx context.Context, userID int64, active bool) error {
	key := c.adminKey(userID)
	if !active {
		return c.set(ctx, key, "0")
	}
	return c.setWithTTL(ctx, key, "1", adminSessionTTL)
}

func (c *Client) GetAdminSession(ctx context.Context, userID int64) (bool, error) {
	key := c.adminKey(userID)
	val, err := c.get(ctx, key)
	if err != nil {
		return false, err
	}
	return val == "1", nil
}

func (c *Client) historyKey(chatID int64) string {
	return fmt.Sprintf(historyKeyFmt, chatID)
}

func (c *Client) modelKey(chatID int64) string {
	return fmt.Sprintf(modelKeyFmt, chatID)
}

func (c *Client) adminKey(userID int64) string {
	return fmt.Sprintf(adminKeyFmt, userID)
}

func (c *Client) get(ctx context.Context, key string) (string, error) {
	conn, err := c.dial(ctx)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	cmd := fmt.Sprintf(commandGet, len(key), key)
	if _, err := conn.Write([]byte(cmd)); err != nil {
		return "", fmt.Errorf("failed to write command: %w", err)
	}

	return c.readBulkString(conn)
}

func (c *Client) set(ctx context.Context, key, value string) error {
	conn, err := c.dial(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	cmd := fmt.Sprintf(commandSet, len(key), key, len(value), value)
	if _, err := conn.Write([]byte(cmd)); err != nil {
		return fmt.Errorf("failed to write command: %w", err)
	}

	return c.readOK(conn)
}

func (c *Client) setWithTTL(ctx context.Context, key, value string, ttl time.Duration) error {
	if err := c.set(ctx, key, value); err != nil {
		return err
	}

	conn, err := c.dial(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	seconds := int(ttl.Seconds())
	secStr := fmt.Sprintf("%d", seconds)
	cmd := fmt.Sprintf(commandExpire, len(key), key, len(secStr), seconds)
	if _, err := conn.Write([]byte(cmd)); err != nil {
		return fmt.Errorf("failed to write expire command: %w", err)
	}

	return nil
}

func (c *Client) dial(ctx context.Context) (net.Conn, error) {
	var d net.Dialer
	d.Timeout = defaultTimeout
	conn, err := d.DialContext(ctx, "tcp", c.addr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}
	return conn, nil
}

func (c *Client) readBulkString(conn net.Conn) (string, error) {
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	resp := string(buf[:n])
	if len(resp) >= 3 && resp[:3] == responseNil {
		return "", nil
	}

	if len(resp) > 0 && resp[0] == responseBulk {
		for i := 1; i < len(resp); i++ {
			if resp[i] == '\r' && i+2 < len(resp) && resp[i+1] == '\n' {
				return resp[i+2 : len(resp)-2], nil
			}
		}
	}

	return "", fmt.Errorf("unexpected response: %s", resp)
}

func (c *Client) readOK(conn net.Conn) error {
	buf := make([]byte, 64)
	n, err := conn.Read(buf)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	resp := string(buf[:n])
	if len(resp) >= 3 && resp[:3] == responseOK {
		return nil
	}

	return fmt.Errorf("unexpected response: %s", resp)
}
