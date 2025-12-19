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
	err := r.pool.QueryRow(ctx, query, sticker.StickerSetName, sticker.FileID, sticker.Chat.ChatID).Scan(&sticker.StickerID)
	return err
}

func (r *StickerRepository) GetRandomByChat(ctx context.Context, chatID int64) (*model.Sticker, error) {
	query := `
		SELECT s.sticker_id, s.sticker_set_name, s.file_id, c.chat_id, c.chat_name
		FROM stickers s
		JOIN chats c ON s.chat_id = c.chat_id
		WHERE s.chat_id = $1
		ORDER BY RANDOM()
		LIMIT 1
	`

	row := r.pool.QueryRow(ctx, query, chatID)

	var sticker model.Sticker
	sticker.Chat = &model.Chat{}

	err := row.Scan(&sticker.StickerID, &sticker.StickerSetName, &sticker.FileID, &sticker.Chat.ChatID, &sticker.Chat.ChatName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &sticker, nil
}

func (r *StickerRepository) ListByChat(ctx context.Context, chatID int64) ([]*model.Sticker, error) {
	query := `
		SELECT s.sticker_id, s.sticker_set_name, s.file_id, c.chat_id, c.chat_name
		FROM stickers s
		JOIN chats c ON s.chat_id = c.chat_id
		WHERE s.chat_id = $1
		ORDER BY s.sticker_id DESC
	`

	rows, err := r.pool.Query(ctx, query, chatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stickers []*model.Sticker
	for rows.Next() {
		var sticker model.Sticker
		sticker.Chat = &model.Chat{}
		err := rows.Scan(&sticker.StickerID, &sticker.StickerSetName, &sticker.FileID, &sticker.Chat.ChatID, &sticker.Chat.ChatName)
		if err != nil {
			return nil, err
		}
		stickers = append(stickers, &sticker)
	}

	return stickers, nil
}
