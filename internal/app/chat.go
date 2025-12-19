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
