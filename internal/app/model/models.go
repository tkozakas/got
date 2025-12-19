package model

import "time"

type Chat struct {
	ChatID   int64   `json:"chat_id"`
	ChatName string  `json:"chat_name"`
	Users    []*User `json:"users"`
}

type User struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
}

type Fact struct {
	ID      int64  `json:"id"`
	Comment string `json:"comment"`
	Chat    *Chat  `json:"chat"`
}

type Reminder struct {
	ReminderID int64     `json:"reminder_id"`
	Chat       *Chat     `json:"chat"`
	User       *User     `json:"user"`
	Message    string    `json:"message"`
	RemindAt   time.Time `json:"remind_at"`
	CreatedAt  time.Time `json:"created_at"`
	Sent       bool      `json:"sent"`
}

type Sentence struct {
	SentenceID  int64  `json:"sentence_id"`
	GroupID     int64  `json:"group_id"`
	OrderNumber int    `json:"order_number"`
	Text        string `json:"text"`
}

type Sticker struct {
	StickerID      string `json:"sticker_id"`
	StickerSetName string `json:"sticker_set_name"`
	FileID         string `json:"file_id"`
	Chat           *Chat  `json:"chat"`
}

type Subreddit struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}
