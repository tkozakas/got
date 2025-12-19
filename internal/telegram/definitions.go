package telegram

import "context"

type HandlerFunc func(ctx context.Context, update *Update) error
