package postgres

import (
	"context"
	"errors"
	"got/internal/app/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FactRepository struct {
	pool *pgxpool.Pool
}

func NewFactRepository(pool *pgxpool.Pool) *FactRepository {
	return &FactRepository{pool: pool}
}

func (r *FactRepository) Save(ctx context.Context, fact *model.Fact) error {
	query := `
		INSERT INTO facts (comment, chat_id)
		VALUES ($1, $2)
		RETURNING fact_id
	`
	var chatID *int64
	if fact.Chat != nil {
		chatID = &fact.Chat.ChatID
	}

	err := r.pool.QueryRow(ctx, query, fact.Comment, chatID).Scan(&fact.ID)
	return err
}

func (r *FactRepository) GetRandom(ctx context.Context) (*model.Fact, error) {
	query := `
		SELECT f.fact_id, f.comment, c.chat_id, c.chat_name
		FROM facts f
		LEFT JOIN chats c ON f.chat_id = c.chat_id
		ORDER BY RANDOM()
		LIMIT 1
	`

	row := r.pool.QueryRow(ctx, query)

	var fact model.Fact
	var chatID *int64
	var chatName *string

	err := row.Scan(&fact.ID, &fact.Comment, &chatID, &chatName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if chatID != nil {
		fact.Chat = &model.Chat{
			ChatID:   *chatID,
			ChatName: *chatName,
		}
	}

	return &fact, nil
}
