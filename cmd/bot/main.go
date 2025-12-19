package main

import (
	"context"
	"got/internal/app"
	"got/internal/telegram"
	"got/pkg/config"
	"log/slog"
	"os"
	"os/signal"
)

func main() {
	cfg := config.Load()

	client := telegram.NewClient(cfg.BotToken)
	svc := app.NewService()

	// Create router and handlers
	router := telegram.NewRouter()
	handlers := telegram.NewBotHandlers(client, svc)

	// Register handlers
	router.Register("start",
		telegram.WithRecover(
			telegram.WithLogging(handlers.HandleStart),
		),
	)

	router.Register("price",
		telegram.WithRecover(
			telegram.WithLogging(handlers.HandlePrice),
		),
	)

	// Create bot with router
	bot := telegram.NewBot(client, router)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	slog.Info("Bot started")
	bot.Start(ctx)
}
