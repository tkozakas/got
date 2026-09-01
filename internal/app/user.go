package app

import (
	"context"
	"got/internal/app/model"
)

func (s *Service) RegisterUser(ctx context.Context, user *model.User, chatID int64) error {
	if err := s.users.Save(ctx, user); err != nil {
		return err
	}
	return s.users.AddToChat(ctx, user.UserID, chatID)
}
