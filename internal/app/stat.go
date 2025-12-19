package app

import (
	"context"
	"fmt"
	"got/internal/app/model"
	"log/slog"
	"time"
)

type RouletteResult struct {
	ChatID int64
	Winner *model.Stat
}

func (s *Service) GetOrCreateStat(ctx context.Context, userID, chatID int64, year int) (*model.Stat, error) {
	stat, err := s.stats.FindByUserChatYear(ctx, userID, chatID, year)
	if err != nil {
		return nil, err
	}
	if stat != nil {
		return stat, nil
	}

	user, err := s.users.Get(ctx, userID)
	if err != nil || user == nil {
		return nil, fmt.Errorf("user not found")
	}

	chat, err := s.chats.Get(ctx, chatID)
	if err != nil || chat == nil {
		return nil, fmt.Errorf("chat not found")
	}

	stat = &model.Stat{
		User:     user,
		Chat:     chat,
		Score:    0,
		Year:     year,
		IsWinner: false,
	}

	if err := s.stats.Save(ctx, stat); err != nil {
		return nil, err
	}

	return stat, nil
}

func (s *Service) GetTodayWinner(ctx context.Context, chatID int64, year int) (*model.Stat, error) {
	return s.stats.FindWinnerByChat(ctx, chatID, year)
}

func (s *Service) SelectRandomWinner(ctx context.Context, chatID int64, year int) (*model.Stat, error) {
	user, err := s.users.GetRandomByChat(ctx, chatID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil
	}

	stat, err := s.GetOrCreateStat(ctx, user.UserID, chatID, year)
	if err != nil {
		return nil, err
	}

	stat.Score++
	stat.IsWinner = true

	if err := s.stats.Save(ctx, stat); err != nil {
		return nil, err
	}

	return stat, nil
}

func (s *Service) GetStatsByYear(ctx context.Context, chatID int64, year int) ([]*model.Stat, error) {
	return s.stats.ListByChatAndYear(ctx, chatID, year)
}

func (s *Service) GetAllStats(ctx context.Context, chatID int64) ([]*model.Stat, error) {
	return s.stats.ListByChat(ctx, chatID)
}

func (s *Service) ResetDailyWinners(ctx context.Context) error {
	return s.stats.ResetDailyWinners(ctx)
}

func (s *Service) RunAutoRoulette(ctx context.Context) ([]RouletteResult, error) {
	chats, err := s.chats.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list chats: %w", err)
	}

	year := time.Now().Year()
	var results []RouletteResult

	for _, chat := range chats {
		result, ok := s.runRouletteForChat(ctx, chat.ChatID, year)
		if ok {
			results = append(results, result)
		}
	}

	return results, nil
}

func (s *Service) runRouletteForChat(ctx context.Context, chatID int64, year int) (RouletteResult, bool) {
	existing, err := s.GetTodayWinner(ctx, chatID, year)
	if err != nil {
		slog.Error("failed to check existing winner", "chat", chatID, "error", err)
		return RouletteResult{}, false
	}
	if existing != nil {
		return RouletteResult{}, false
	}

	winner, err := s.SelectRandomWinner(ctx, chatID, year)
	if err != nil {
		slog.Error("failed to select winner", "chat", chatID, "error", err)
		return RouletteResult{}, false
	}
	if winner == nil {
		return RouletteResult{}, false
	}

	return RouletteResult{ChatID: chatID, Winner: winner}, true
}
