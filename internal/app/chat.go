package app

import (
	"context"
	"got/internal/app/model"
)

func (s *Service) RegisterChat(ctx context.Context, chat *model.Chat) error {
	return s.chats.Save(ctx, chat)
}
