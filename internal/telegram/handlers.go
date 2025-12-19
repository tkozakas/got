package telegram

import (
	"context"
	"fmt"
	"got/internal/app"
)

type BotHandlers struct {
	client  *Client
	service *app.Service
}

func NewBotHandlers(client *Client, service *app.Service) *BotHandlers {
	return &BotHandlers{
		client:  client,
		service: service,
	}
}

func (h *BotHandlers) HandleStart(ctx context.Context, update *Update) error {
	return h.client.SendMessage(update.Message.Chat.ID, "Welcome!")
}

func (h *BotHandlers) HandlePrice(ctx context.Context, update *Update) error {
	ticker := update.Message.CommandArguments()

	price, err := h.service.GetCryptoPrice(ctx, ticker)
	if err != nil {
		return err
	}

	return h.sendPrice(update.Message.Chat.ID, price)
}

func (h *BotHandlers) sendPrice(chatID int64, price float64) error {
	msg := fmt.Sprintf("Price: %f", price)
	return h.client.SendMessage(chatID, msg)
}
