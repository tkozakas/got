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

func (r *UserRepository) AddToChat(ctx context.Context, userID, chatID int64) error {
	query := `
		INSERT INTO chat_users (chat_id, user_id)
		VALUES ($1, $2)
		ON CONFLICT (chat_id, user_id) DO NOTHING
	`
	_, err := r.pool.Exec(ctx, query, chatID, userID)
	return err
}

func (r *UserRepository) GetRandomByChat(ctx context.Context, chatID int64) (*model.User, error) {
	query := `
		SELECT u.user_id, u.username
		FROM users u
		JOIN chat_users cu ON u.user_id = cu.user_id
		WHERE cu.chat_id = $1
		ORDER BY RANDOM()
		LIMIT 1
	`

	row := r.pool.QueryRow(ctx, query, chatID)

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
