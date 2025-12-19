package telegram

import (
	"context"
	"log"
)

type Middleware func(HandlerFunc) HandlerFunc

func WithLogging(next HandlerFunc) HandlerFunc {
	return func(ctx context.Context, update *Update) error {
		if update.Message != nil {
			log.Printf("User %s command %s", update.Message.From.UserName, update.Message.Command())
		}
		return next(ctx, update)
	}
}

func WithRecover(next HandlerFunc) HandlerFunc {
	return func(ctx context.Context, update *Update) error {
		defer func() {
			if r := recover(); r != nil {
				log.Println("Recovered:", r)
			}
		}()
		return next(ctx, update)
	}
}
