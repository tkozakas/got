package postgres

import (
	"context"
	"errors"
	"got/internal/app/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ChatRepository struct {
	pool *pgxpool.Pool
}

func NewChatRepository(pool *pgxpool.Pool) *ChatRepository {
	return &ChatRepository{pool: pool}
}

func (r *ChatRepository) Save(ctx context.Context, chat *model.Chat) error {
	query := `
		INSERT INTO chats (chat_id, chat_name)
		VALUES ($1, $2)
		ON CONFLICT (chat_id) DO UPDATE 
		SET chat_name = EXCLUDED.chat_name
	`
	_, err := r.pool.Exec(ctx, query, chat.ChatID, chat.ChatName)
	return err
}

func (r *ChatRepository) Get(ctx context.Context, chatID int64) (*model.Chat, error) {
	query := `SELECT chat_id, chat_name FROM chats WHERE chat_id = $1`

	row := r.pool.QueryRow(ctx, query, chatID)

	var chat model.Chat
	err := row.Scan(&chat.ChatID, &chat.ChatName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &chat, nil
}
