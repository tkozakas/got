package postgres

import (
	"context"
	"errors"
	"got/internal/app/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StickerRepository struct {
	pool *pgxpool.Pool
}

func NewStickerRepository(pool *pgxpool.Pool) *StickerRepository {
	return &StickerRepository{pool: pool}
}

func (r *StickerRepository) Save(ctx context.Context, sticker *model.Sticker) error {
	query := `
		INSERT INTO stickers (sticker_set_name, file_id, chat_id)
		VALUES ($1, $2, $3)
		RETURNING sticker_id
	`
	var chatID *int64
	if sticker.Chat != nil {
		chatID = &sticker.Chat.ChatID
	}

	err := r.pool.QueryRow(ctx, query, sticker.StickerSetName, sticker.FileID, chatID).Scan(&sticker.StickerID)
	return err
}

func (r *StickerRepository) GetRandom(ctx context.Context) (*model.Sticker, error) {
	query := `
		SELECT s.sticker_id, s.sticker_set_name, s.file_id, c.chat_id, c.chat_name
		FROM stickers s
		LEFT JOIN chats c ON s.chat_id = c.chat_id
		ORDER BY RANDOM()
		LIMIT 1
	`

	row := r.pool.QueryRow(ctx, query)

	var sticker model.Sticker
	var chatID *int64
	var chatName *string

	err := row.Scan(&sticker.StickerID, &sticker.StickerSetName, &sticker.FileID, &chatID, &chatName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if chatID != nil {
		sticker.Chat = &model.Chat{
			ChatID:   *chatID,
			ChatName: *chatName,
		}
	}

	return &sticker, nil
}
