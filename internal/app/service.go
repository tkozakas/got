package app

type Service struct {
	chats      ChatRepository
	users      UserRepository
	reminders  ReminderRepository
	facts      FactRepository
	stickers   StickerRepository
	subreddits SubredditRepository
	stats      StatRepository
}

func NewService(
	chats ChatRepository,
	users UserRepository,
	reminders ReminderRepository,
	facts FactRepository,
	stickers StickerRepository,
	subreddits SubredditRepository,
	stats StatRepository,
) *Service {
	return &Service{
		chats:      chats,
		users:      users,
		reminders:  reminders,
		facts:      facts,
		stickers:   stickers,
		subreddits: subreddits,
		stats:      stats,
	}
}
