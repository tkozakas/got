package app

import (
	"context"
	"fmt"
	"got/internal/app/model"
)

func (s *Service) AddSticker(ctx context.Context, fileID string, chatID int64) error {
	chat, err := s.chats.Get(ctx, chatID)
	if err != nil {
		return fmt.Errorf("failed to get chat: %w", err)
	}
	if chat == nil {
		return fmt.Errorf("chat not registered")
	}

	sticker := &model.Sticker{
		FileID: fileID,
		Chat:   chat,
	}
	return s.stickers.Save(ctx, sticker)
}

func (s *Service) GetRandomSticker(ctx context.Context, chatID int64) (*model.Sticker, error) {
	return s.stickers.GetRandomByChat(ctx, chatID)
}

func (s *Service) ListStickers(ctx context.Context, chatID int64) ([]*model.Sticker, error) {
	return s.stickers.ListByChat(ctx, chatID)
}
