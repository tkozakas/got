package app

import (
	"context"
	"fmt"
	"got/internal/app/model"
)

func (s *Service) AddFact(ctx context.Context, text string, chatID int64) error {
	chat, err := s.chats.Get(ctx, chatID)
	if err != nil {
		return fmt.Errorf("failed to get chat: %w", err)
	}
	if chat == nil {
		return fmt.Errorf("chat not registered")
	}

	fact := &model.Fact{
		Comment: text,
		Chat:    chat,
	}
	return s.facts.Save(ctx, fact)
}

func (s *Service) GetRandomFact(ctx context.Context, chatID int64) (*model.Fact, error) {
	return s.facts.GetRandomByChat(ctx, chatID)
}

func (s *Service) ListFacts(ctx context.Context, chatID int64) ([]*model.Fact, error) {
	return s.facts.ListByChat(ctx, chatID)
}
