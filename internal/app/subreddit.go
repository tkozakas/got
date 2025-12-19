package app

import (
	"context"
	"got/internal/app/model"
)

func (s *Service) AddSubreddit(ctx context.Context, name string, chatID int64) error {
	sub := &model.Subreddit{
		Name:   name,
		ChatID: chatID,
	}
	return s.subreddits.Save(ctx, sub)
}

func (s *Service) GetRandomSubreddit(ctx context.Context, chatID int64) (*model.Subreddit, error) {
	return s.subreddits.GetRandomByChat(ctx, chatID)
}

func (s *Service) ListSubreddits(ctx context.Context, chatID int64) ([]*model.Subreddit, error) {
	return s.subreddits.ListByChat(ctx, chatID)
}

func (s *Service) RemoveSubreddit(ctx context.Context, name string, chatID int64) error {
	return s.subreddits.Delete(ctx, name, chatID)
}
