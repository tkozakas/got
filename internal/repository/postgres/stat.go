package postgres

import (
	"context"
	"errors"
	"got/internal/app/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StatRepository struct {
	pool *pgxpool.Pool
}

func NewStatRepository(pool *pgxpool.Pool) *StatRepository {
	return &StatRepository{pool: pool}
}

func (r *StatRepository) Save(ctx context.Context, stat *model.Stat) error {
	query := `
		INSERT INTO stats (user_id, chat_id, score, year, is_winner)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, chat_id, year) DO UPDATE 
		SET score = EXCLUDED.score, is_winner = EXCLUDED.is_winner
		RETURNING stat_id
	`
	err := r.pool.QueryRow(ctx, query,
		stat.User.UserID,
		stat.Chat.ChatID,
		stat.Score,
		stat.Year,
		stat.IsWinner,
	).Scan(&stat.StatID)
	return err
}

func (r *StatRepository) FindByUserChatYear(ctx context.Context, userID, chatID int64, year int) (*model.Stat, error) {
	query := `
		SELECT s.stat_id, s.score, s.year, s.is_winner,
		       u.user_id, u.username,
		       c.chat_id, c.chat_name
		FROM stats s
		JOIN users u ON s.user_id = u.user_id
		JOIN chats c ON s.chat_id = c.chat_id
		WHERE s.user_id = $1 AND s.chat_id = $2 AND s.year = $3
	`

	row := r.pool.QueryRow(ctx, query, userID, chatID, year)
	return r.scanStat(row)
}

func (r *StatRepository) FindWinnerByChat(ctx context.Context, chatID int64, year int) (*model.Stat, error) {
	query := `
		SELECT s.stat_id, s.score, s.year, s.is_winner,
		       u.user_id, u.username,
		       c.chat_id, c.chat_name
		FROM stats s
		JOIN users u ON s.user_id = u.user_id
		JOIN chats c ON s.chat_id = c.chat_id
		WHERE s.chat_id = $1 AND s.year = $2 AND s.is_winner = true
	`

	row := r.pool.QueryRow(ctx, query, chatID, year)
	return r.scanStat(row)
}

func (r *StatRepository) ListByChatAndYear(ctx context.Context, chatID int64, year int) ([]*model.Stat, error) {
	query := `
		SELECT s.stat_id, s.score, s.year, s.is_winner,
		       u.user_id, u.username,
		       c.chat_id, c.chat_name
		FROM stats s
		JOIN users u ON s.user_id = u.user_id
		JOIN chats c ON s.chat_id = c.chat_id
		WHERE s.chat_id = $1 AND s.year = $2
		ORDER BY s.score DESC
	`

	return r.queryStats(ctx, query, chatID, year)
}

func (r *StatRepository) ListByChat(ctx context.Context, chatID int64) ([]*model.Stat, error) {
	query := `
		SELECT s.stat_id, s.score, s.year, s.is_winner,
		       u.user_id, u.username,
		       c.chat_id, c.chat_name
		FROM stats s
		JOIN users u ON s.user_id = u.user_id
		JOIN chats c ON s.chat_id = c.chat_id
		WHERE s.chat_id = $1
		ORDER BY s.score DESC
	`

	return r.queryStats(ctx, query, chatID)
}

func (r *StatRepository) ResetDailyWinners(ctx context.Context) error {
	query := `UPDATE stats SET is_winner = false WHERE is_winner = true`
	_, err := r.pool.Exec(ctx, query)
	return err
}

func (r *StatRepository) ResetWinnerByChat(ctx context.Context, chatID int64, year int) error {
	query := `UPDATE stats SET is_winner = false WHERE chat_id = $1 AND year = $2 AND is_winner = true`
	_, err := r.pool.Exec(ctx, query, chatID, year)
	return err
}

func (r *StatRepository) Update(ctx context.Context, statID int64, score int64, isWinner bool) error {
	query := `UPDATE stats SET score = $1, is_winner = $2 WHERE stat_id = $3`
	_, err := r.pool.Exec(ctx, query, score, isWinner, statID)
	return err
}

func (r *StatRepository) queryStats(ctx context.Context, query string, args ...any) ([]*model.Stat, error) {
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []*model.Stat
	for rows.Next() {
		stat, err := r.scanStatFromRows(rows)
		if err != nil {
			return nil, err
		}
		stats = append(stats, stat)
	}

	return stats, nil
}

func (r *StatRepository) scanStat(row pgx.Row) (*model.Stat, error) {
	var stat model.Stat
	stat.User = &model.User{}
	stat.Chat = &model.Chat{}

	err := row.Scan(
		&stat.StatID, &stat.Score, &stat.Year, &stat.IsWinner,
		&stat.User.UserID, &stat.User.Username,
		&stat.Chat.ChatID, &stat.Chat.ChatName,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &stat, nil
}

func (r *StatRepository) scanStatFromRows(rows pgx.Rows) (*model.Stat, error) {
	var stat model.Stat
	stat.User = &model.User{}
	stat.Chat = &model.Chat{}

	err := rows.Scan(
		&stat.StatID, &stat.Score, &stat.Year, &stat.IsWinner,
		&stat.User.UserID, &stat.User.Username,
		&stat.Chat.ChatID, &stat.Chat.ChatName,
	)
	if err != nil {
		return nil, err
	}

	return &stat, nil
}
