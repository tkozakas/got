package telegram

import (
	"context"
	"log/slog"
)

type Router struct {
	handlers map[string]HandlerFunc
}

func NewRouter() *Router {
	return &Router{
		handlers: make(map[string]HandlerFunc),
	}
}

func (r *Router) Register(command string, handler HandlerFunc) {
	r.handlers[command] = handler
}

func (r *Router) Handle(ctx context.Context, update *Update) error {
	if update.Message == nil {
		return nil
	}

	cmd := update.Message.Command()
	if cmd == "" {
		return nil
	}

	return r.executeCommand(ctx, cmd, update)
}

func (r *Router) executeCommand(ctx context.Context, cmd string, update *Update) error {
	if handler, exists := r.handlers[cmd]; exists {
		return handler(ctx, update)
	}

	slog.Info("Unknown command", "command", cmd)
	return nil
}
