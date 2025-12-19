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
		INSERT INTO subreddits (name)
		VALUES ($1)
		RETURNING id
	`
	err := r.pool.QueryRow(ctx, query, sub.Name).Scan(&sub.ID)
	return err
}

func (r *SubredditRepository) GetRandom(ctx context.Context) (*model.Subreddit, error) {
	query := `SELECT id, name FROM subreddits ORDER BY RANDOM() LIMIT 1`

	row := r.pool.QueryRow(ctx, query)

	var sub model.Subreddit
	err := row.Scan(&sub.ID, &sub.Name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &sub, nil
}
