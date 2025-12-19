package postgres

import (
	"context"
	"got/internal/app/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ReminderRepository struct {
	pool *pgxpool.Pool
}

func NewReminderRepository(pool *pgxpool.Pool) *ReminderRepository {
	return &ReminderRepository{pool: pool}
}

func (r *ReminderRepository) Save(ctx context.Context, reminder *model.Reminder) error {
	query := `
		INSERT INTO reminders (chat_id, user_id, message, remind_at, created_at, sent)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING reminder_id
	`
	err := r.pool.QueryRow(ctx, query,
		reminder.Chat.ChatID,
		reminder.User.UserID,
		reminder.Message,
		reminder.RemindAt,
		reminder.CreatedAt,
		reminder.Sent,
	).Scan(&reminder.ReminderID)
	return err
}

func (r *ReminderRepository) ListPending(ctx context.Context) ([]*model.Reminder, error) {
	query := `
		SELECT 
			r.reminder_id, r.message, r.remind_at, r.created_at, r.sent,
			c.chat_id, c.chat_name,
			u.user_id, u.username
		FROM reminders r
		JOIN chats c ON r.chat_id = c.chat_id
		JOIN users u ON r.user_id = u.user_id
		WHERE r.sent = false AND r.remind_at <= NOW()
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reminders []*model.Reminder
	for rows.Next() {
		var r model.Reminder
		r.Chat = &model.Chat{}
		r.User = &model.User{}

		err := rows.Scan(
			&r.ReminderID, &r.Message, &r.RemindAt, &r.CreatedAt, &r.Sent,
			&r.Chat.ChatID, &r.Chat.ChatName,
			&r.User.UserID, &r.User.Username,
		)
		if err != nil {
			return nil, err
		}
		reminders = append(reminders, &r)
	}

	return reminders, nil
}

func (r *ReminderRepository) ListByChat(ctx context.Context, chatID int64) ([]*model.Reminder, error) {
	query := `
		SELECT 
			r.reminder_id, r.message, r.remind_at, r.created_at, r.sent,
			c.chat_id, c.chat_name,
			u.user_id, u.username
		FROM reminders r
		JOIN chats c ON r.chat_id = c.chat_id
		JOIN users u ON r.user_id = u.user_id
		WHERE r.chat_id = $1 AND r.sent = false
		ORDER BY r.remind_at ASC
	`

	rows, err := r.pool.Query(ctx, query, chatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reminders []*model.Reminder
	for rows.Next() {
		var r model.Reminder
		r.Chat = &model.Chat{}
		r.User = &model.User{}

		err := rows.Scan(
			&r.ReminderID, &r.Message, &r.RemindAt, &r.CreatedAt, &r.Sent,
			&r.Chat.ChatID, &r.Chat.ChatName,
			&r.User.UserID, &r.User.Username,
		)
		if err != nil {
			return nil, err
		}
		reminders = append(reminders, &r)
	}

	return reminders, nil
}

func (r *ReminderRepository) MarkSent(ctx context.Context, reminderID int64) error {
	query := `UPDATE reminders SET sent = true WHERE reminder_id = $1`
	_, err := r.pool.Exec(ctx, query, reminderID)
	return err
}
