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
	err := r.pool.QueryRow(ctx, query, fact.Comment, fact.Chat.ChatID).Scan(&fact.ID)
	return err
}

func (r *FactRepository) GetRandomByChat(ctx context.Context, chatID int64) (*model.Fact, error) {
	query := `
		SELECT f.fact_id, f.comment, c.chat_id, c.chat_name
		FROM facts f
		JOIN chats c ON f.chat_id = c.chat_id
		WHERE f.chat_id = $1
		ORDER BY RANDOM()
		LIMIT 1
	`

	row := r.pool.QueryRow(ctx, query, chatID)

	var fact model.Fact
	fact.Chat = &model.Chat{}

	err := row.Scan(&fact.ID, &fact.Comment, &fact.Chat.ChatID, &fact.Chat.ChatName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &fact, nil
}

func (r *FactRepository) ListByChat(ctx context.Context, chatID int64) ([]*model.Fact, error) {
	query := `
		SELECT f.fact_id, f.comment, c.chat_id, c.chat_name
		FROM facts f
		JOIN chats c ON f.chat_id = c.chat_id
		WHERE f.chat_id = $1
		ORDER BY f.fact_id DESC
	`

	rows, err := r.pool.Query(ctx, query, chatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var facts []*model.Fact
	for rows.Next() {
		var fact model.Fact
		fact.Chat = &model.Chat{}
		err := rows.Scan(&fact.ID, &fact.Comment, &fact.Chat.ChatID, &fact.Chat.ChatName)
		if err != nil {
			return nil, err
		}
		facts = append(facts, &fact)
	}

	return facts, nil
}
