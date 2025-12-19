package postgres

import (
	"context"
	"errors"
	"got/internal/app/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SubredditRepository struct {
	pool *pgxpool.Pool
}

func NewSubredditRepository(pool *pgxpool.Pool) *SubredditRepository {
	return &SubredditRepository{pool: pool}
}

func (r *SubredditRepository) Save(ctx context.Context, sub *model.Subreddit) error {
	query := `
		INSERT INTO subreddits (name, chat_id)
		VALUES ($1, $2)
		ON CONFLICT (name, chat_id) DO NOTHING
		RETURNING id
	`
	err := r.pool.QueryRow(ctx, query, sub.Name, sub.ChatID).Scan(&sub.ID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil // Already exists
	}
	return err
}

func (r *SubredditRepository) GetRandomByChat(ctx context.Context, chatID int64) (*model.Subreddit, error) {
	query := `SELECT id, name, chat_id FROM subreddits WHERE chat_id = $1 ORDER BY RANDOM() LIMIT 1`

	row := r.pool.QueryRow(ctx, query, chatID)

	var sub model.Subreddit
	err := row.Scan(&sub.ID, &sub.Name, &sub.ChatID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &sub, nil
}

func (r *SubredditRepository) ListByChat(ctx context.Context, chatID int64) ([]*model.Subreddit, error) {
	query := `SELECT id, name, chat_id FROM subreddits WHERE chat_id = $1 ORDER BY name`

	rows, err := r.pool.Query(ctx, query, chatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []*model.Subreddit
	for rows.Next() {
		var sub model.Subreddit
		err := rows.Scan(&sub.ID, &sub.Name, &sub.ChatID)
		if err != nil {
			return nil, err
		}
		subs = append(subs, &sub)
	}

	return subs, nil
}

func (r *SubredditRepository) Delete(ctx context.Context, name string, chatID int64) error {
	query := `DELETE FROM subreddits WHERE name = $1 AND chat_id = $2`
	_, err := r.pool.Exec(ctx, query, name, chatID)
	return err
}
