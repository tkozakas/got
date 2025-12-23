package model

import "time"

type Chat struct {
	ChatID   int64   `json:"chat_id"`
	ChatName string  `json:"chat_name"`
	Language string  `json:"language"`
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

type Sticker struct {
	StickerID      string `json:"sticker_id"`
	StickerSetName string `json:"sticker_set_name"`
	FileID         string `json:"file_id"`
	Chat           *Chat  `json:"chat"`
}

type Subreddit struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	ChatID int64  `json:"chat_id"`
}

type Stat struct {
	StatID   int64 `json:"stat_id"`
	User     *User `json:"user"`
	Chat     *Chat `json:"chat"`
	Score    int64 `json:"score"`
	Year     int   `json:"year"`
	IsWinner bool  `json:"is_winner"`
}

type RedditResponse struct {
	Memes []RedditMeme `json:"memes"`
}

type RedditMeme struct {
	Title     string `json:"title"`
	URL       string `json:"url"`
	Author    string `json:"author"`
	Subreddit string `json:"subreddit"`
}
