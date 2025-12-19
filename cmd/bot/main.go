package main

import (
	"context"
	"fmt"
	"got/internal/app"
	"got/internal/app/model"
	"got/internal/groq"
	"got/internal/redis"
	"got/internal/repository/postgres"
	"got/internal/scheduler"
	"got/internal/telegram"
	"got/internal/tts"
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

	ttsClient := tts.NewClient()

	router := telegram.NewRouter()
	handlers := telegram.NewBotHandlers(client, svc, gptClient, redisClient, translator, ttsClient, &cfg.Commands, cfg.AdminPass)

	cmds := &cfg.Commands
	router.Register(cmds.Start, telegram.WithRecover(telegram.WithLogging(handlers.HandleStart)))
	router.Register(cmds.Help, telegram.WithRecover(telegram.WithLogging(handlers.HandleHelp)))
	router.Register(cmds.Gpt, telegram.WithRecover(telegram.WithLogging(handlers.HandleGPT)))
	router.Register(cmds.Remind, telegram.WithRecover(telegram.WithLogging(handlers.HandleRemind)))
	router.Register(cmds.Meme, telegram.WithRecover(telegram.WithLogging(handlers.HandleMeme)))
	router.Register(cmds.Sticker, telegram.WithRecover(telegram.WithLogging(handlers.HandleSticker)))
	router.Register(cmds.Fact, telegram.WithRecover(telegram.WithLogging(handlers.HandleFact)))
	router.Register(cmds.Roulette, telegram.WithRecover(telegram.WithLogging(handlers.HandleRoulette)))
	router.Register(cmds.Tts, telegram.WithRecover(telegram.WithLogging(handlers.HandleTTS)))
	router.Register(cmds.Admin, telegram.WithRecover(telegram.WithLogging(handlers.HandleAdmin)))

	autoRegister := telegram.NewAutoRegisterMiddleware(svc, router)
	bot := telegram.NewBot(client, autoRegister)

	sentences := telegram.NewSentenceProvider()

	registerBotCommands(client, translator, &cfg.Commands)
	sched := startScheduler(cfg, svc, client, translator, sentences)
	defer sched.Stop()

	go runReminderChecker(ctx, svc, client, translator)

	slog.Info("Bot started", "language", translator.Lang())
	bot.Start(ctx)
}

func startScheduler(cfg *config.Config, svc *app.Service, client *telegram.Client, t *i18n.Translator, sentences *telegram.SentenceProvider) *scheduler.Scheduler {
	sched := scheduler.New()

	_ = sched.Register(scheduler.Job{
		Name:     "reset_daily_winners",
		Schedule: cfg.Schedule.WinnerReset,
		Func:     svc.ResetDailyWinners,
	})

	_ = sched.Register(scheduler.Job{
		Name:     "auto_roulette",
		Schedule: cfg.Schedule.AutoRoulette,
		Func:     autoRouletteJob(svc, client, t, sentences),
	})

	sched.Start()
	return sched
}

func autoRouletteJob(svc *app.Service, client *telegram.Client, t *i18n.Translator, sentences *telegram.SentenceProvider) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		results, err := svc.RunAutoRoulette(ctx)
		if err != nil {
			return err
		}

		alias := t.Get(i18n.KeyRouletteAlias)
		for _, r := range results {
			winnerName := formatUserLink(r.Winner.User)
			fallbackMsg := fmt.Sprintf(t.Get(i18n.KeyRouletteAutoWinner), alias, winnerName)
			if err := sentences.SendSequence(client, r.ChatID, alias, winnerName, fallbackMsg); err != nil {
				slog.Error("Failed to send auto roulette result", "chat", r.ChatID, "error", err)
			}
		}

		return nil
	}
}

func formatUserLink(user *model.User) string {
	name := user.Username
	if name == "" {
		name = fmt.Sprintf("User%d", user.UserID)
	}
	return fmt.Sprintf("[%s](tg://user?id=%d)", name, user.UserID)
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

func registerBotCommands(client *telegram.Client, t *i18n.Translator, cmds *config.CommandsConfig) {
	commands := []telegram.BotCommand{
		{Command: cmds.Start, Description: t.Get(i18n.KeyCmdStart)},
		{Command: cmds.Help, Description: t.Get(i18n.KeyCmdHelp)},
		{Command: cmds.Gpt, Description: t.Get(i18n.KeyCmdGpt)},
		{Command: cmds.Remind, Description: t.Get(i18n.KeyCmdRemind)},
		{Command: cmds.Meme, Description: t.Get(i18n.KeyCmdMeme)},
		{Command: cmds.Sticker, Description: t.Get(i18n.KeyCmdSticker)},
		{Command: cmds.Fact, Description: t.Get(i18n.KeyCmdFact)},
		{Command: cmds.Roulette, Description: t.Get(i18n.KeyCmdRoulette)},
		{Command: cmds.Tts, Description: t.Get(i18n.KeyCmdTts)},
	}

	if err := client.SetMyCommands(commands); err != nil {
		slog.Error("Failed to register bot commands", "error", err)
		return
	}

	slog.Info("Bot commands registered")
}
