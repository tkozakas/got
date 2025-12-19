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
		INSERT INTO chats (chat_id, chat_name, language)
		VALUES ($1, $2, $3)
		ON CONFLICT (chat_id) DO UPDATE 
		SET chat_name = EXCLUDED.chat_name
	`
	_, err := r.pool.Exec(ctx, query, chat.ChatID, chat.ChatName, chat.Language)
	return err
}

func (r *ChatRepository) Get(ctx context.Context, chatID int64) (*model.Chat, error) {
	query := `SELECT chat_id, chat_name, language FROM chats WHERE chat_id = $1`

	row := r.pool.QueryRow(ctx, query, chatID)

	var chat model.Chat
	err := row.Scan(&chat.ChatID, &chat.ChatName, &chat.Language)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &chat, nil
}

func (r *ChatRepository) ListAll(ctx context.Context) ([]*model.Chat, error) {
	query := `SELECT chat_id, chat_name, language FROM chats`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chats []*model.Chat
	for rows.Next() {
		var chat model.Chat
		if err := rows.Scan(&chat.ChatID, &chat.ChatName, &chat.Language); err != nil {
			return nil, err
		}
		chats = append(chats, &chat)
	}

	return chats, rows.Err()
}

func (r *ChatRepository) SetLanguage(ctx context.Context, chatID int64, language string) error {
	query := `UPDATE chats SET language = $1 WHERE chat_id = $2`
	_, err := r.pool.Exec(ctx, query, language, chatID)
	return err
}

func (r *ChatRepository) GetLanguage(ctx context.Context, chatID int64) (string, error) {
	query := `SELECT language FROM chats WHERE chat_id = $1`
	var lang string
	err := r.pool.QueryRow(ctx, query, chatID).Scan(&lang)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	return lang, nil
}
