package app

import (
	"context"
	"got/internal/app/model"
)

func (s *Service) RegisterChat(ctx context.Context, chat *model.Chat) error {
	return s.chats.Save(ctx, chat)
}

func (s *Service) ListChats(ctx context.Context) ([]*model.Chat, error) {
	return s.chats.ListAll(ctx)
}

func (s *Service) SetChatLanguage(ctx context.Context, chatID int64, language string) error {
	return s.chats.SetLanguage(ctx, chatID, language)
}

func (s *Service) GetChatLanguage(ctx context.Context, chatID int64) (string, error) {
	return s.chats.GetLanguage(ctx, chatID)
}
