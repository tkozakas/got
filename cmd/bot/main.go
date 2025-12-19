package main

import (
	"context"
	"fmt"
	"got/internal/app"
	"got/internal/groq"
	"got/internal/redis"
	"got/internal/repository/postgres"
	"got/internal/scheduler"
	"got/internal/telegram"
	"got/pkg/config"
	"got/pkg/i18n"
	"log/slog"
	"os"
	"os/signal"
	"time"
)

func main() {
	cfg := config.Load()
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	dbPool, err := postgres.NewDB(ctx, cfg.DBURL)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer dbPool.Close()

	chatRepo := postgres.NewChatRepository(dbPool)
	userRepo := postgres.NewUserRepository(dbPool)
	reminderRepo := postgres.NewReminderRepository(dbPool)
	factRepo := postgres.NewFactRepository(dbPool)
	stickerRepo := postgres.NewStickerRepository(dbPool)
	subredditRepo := postgres.NewSubredditRepository(dbPool)
	statRepo := postgres.NewStatRepository(dbPool)

	svc := app.NewService(chatRepo, userRepo, reminderRepo, factRepo, stickerRepo, subredditRepo, statRepo)

	client := telegram.NewClient(cfg.BotToken)
	translator := i18n.New(cfg.Bot.Language)

	var gptClient *groq.Client
	if cfg.GptKey != "" {
		gptClient = groq.NewClient(cfg.GptKey)
	}

	var redisClient *redis.Client
	if cfg.RedisAddr != "" {
		redisClient = redis.NewClient(cfg.RedisAddr)
	}

	router := telegram.NewRouter()
	handlers := telegram.NewBotHandlers(client, svc, gptClient, redisClient, translator)

	router.Register("start", telegram.WithRecover(telegram.WithLogging(handlers.HandleStart)))
	router.Register("gpt", telegram.WithRecover(telegram.WithLogging(handlers.HandleGPT)))
	router.Register("remind", telegram.WithRecover(telegram.WithLogging(handlers.HandleRemind)))
	router.Register("meme", telegram.WithRecover(telegram.WithLogging(handlers.HandleMeme)))
	router.Register("sticker", telegram.WithRecover(telegram.WithLogging(handlers.HandleSticker)))
	router.Register("fact", telegram.WithRecover(telegram.WithLogging(handlers.HandleFact)))
	router.Register("stats", telegram.WithRecover(telegram.WithLogging(handlers.HandleStats)))

	bot := telegram.NewBot(client, router)

	registerBotCommands(client, translator)
	sched := startScheduler(cfg, svc)
	defer sched.Stop()

	go runReminderChecker(ctx, svc, client, translator)

	slog.Info("Bot started", "language", translator.Lang())
	bot.Start(ctx)
}

func startScheduler(cfg *config.Config, svc *app.Service) *scheduler.Scheduler {
	sched := scheduler.New()

	_ = sched.Register(scheduler.Job{
		Name:     "reset_daily_winners",
		Schedule: cfg.Schedule.WinnerReset,
		Func:     svc.ResetDailyWinners,
	})

	sched.Start()
	return sched
}

func runReminderChecker(ctx context.Context, svc *app.Service, client *telegram.Client, t *i18n.Translator) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			checkReminders(ctx, svc, client, t)
		}
	}
}

func checkReminders(ctx context.Context, svc *app.Service, client *telegram.Client, t *i18n.Translator) {
	reminders, err := svc.CheckReminders(ctx)
	if err != nil {
		slog.Error("Failed to check reminders", "error", err)
		return
	}

	for _, r := range reminders {
		msg := fmt.Sprintf(t.Get(i18n.KeyReminderNotify), r.Message)
		if err := client.SendMessage(r.Chat.ChatID, msg); err != nil {
			slog.Error("Failed to send reminder", "id", r.ReminderID, "error", err)
		}
	}
}

func registerBotCommands(client *telegram.Client, t *i18n.Translator) {
	commands := []telegram.BotCommand{
		{Command: "start", Description: t.Get(i18n.KeyCmdStart)},
		{Command: "gpt", Description: t.Get(i18n.KeyCmdGpt)},
		{Command: "remind", Description: t.Get(i18n.KeyCmdRemind)},
		{Command: "meme", Description: t.Get(i18n.KeyCmdMeme)},
		{Command: "sticker", Description: t.Get(i18n.KeyCmdSticker)},
		{Command: "fact", Description: t.Get(i18n.KeyCmdFact)},
		{Command: "stats", Description: t.Get(i18n.KeyCmdStats)},
	}

	if err := client.SetMyCommands(commands); err != nil {
		slog.Error("Failed to register bot commands", "error", err)
		return
	}

	slog.Info("Bot commands registered")
}
