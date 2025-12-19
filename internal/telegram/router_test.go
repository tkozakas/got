package telegram

import (
	"context"
	"errors"
	"testing"
)

func TestRouterHandle(t *testing.T) {
	mockErr := errors.New("handler error")

	successHandler := func(ctx context.Context, update *Update) error {
		return nil
	}

	failureHandler := func(ctx context.Context, update *Update) error {
		return mockErr
	}

	tests := []struct {
		name      string
		handlers  map[string]HandlerFunc
		update    *Update
		wantErr   bool
		targetErr error
	}{
		{
			name:     "Nil message",
			handlers: map[string]HandlerFunc{},
			update:   &Update{Message: nil},
			wantErr:  false,
		},
		{
			name:     "No command text",
			handlers: map[string]HandlerFunc{},
			update: &Update{
				Message: &Message{Text: "just text"},
			},
			wantErr: false,
		},
		{
			name: "Unregistered command",
			handlers: map[string]HandlerFunc{
				"start": successHandler,
			},
			update: &Update{
				Message: &Message{Text: "/help"},
			},
			wantErr: false,
		},
		{
			name: "Registered command success",
			handlers: map[string]HandlerFunc{
				"start": successHandler,
			},
			update: &Update{
				Message: &Message{Text: "/start"},
			},
			wantErr: false,
		},
		{
			name: "Registered command failure",
			handlers: map[string]HandlerFunc{
				"fail": failureHandler,
			},
			update: &Update{
				Message: &Message{Text: "/fail"},
			},
			wantErr:   true,
			targetErr: mockErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRouter()
			for cmd, h := range tt.handlers {
				r.Register(cmd, h)
			}

			err := r.Handle(context.Background(), tt.update)

			if (err != nil) != tt.wantErr {
				t.Errorf("Router.Handle() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.targetErr != nil && err != tt.targetErr {
				t.Errorf("Router.Handle() error = %v, want %v", err, tt.targetErr)
			}
		})
	}
}
