package postgres

import (
	"context"
	"errors"
	"got/internal/app/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Save(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (user_id, username)
		VALUES ($1, $2)
		ON CONFLICT (user_id) DO UPDATE 
		SET username = EXCLUDED.username
	`
	_, err := r.pool.Exec(ctx, query, user.UserID, user.Username)
	return err
}

func (r *UserRepository) Get(ctx context.Context, userID int64) (*model.User, error) {
	query := `SELECT user_id, username FROM users WHERE user_id = $1`

	row := r.pool.QueryRow(ctx, query, userID)

	var user model.User
	err := row.Scan(&user.UserID, &user.Username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}
