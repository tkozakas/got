package telegram

import "testing"

func TestMessageCommand(t *testing.T) {
	tests := []struct {
		name string
		text string
		want string
	}{
		{
			name: "Simple command",
			text: "/start",
			want: "start",
		},
		{
			name: "Command with argument",
			text: "/price BTC",
			want: "price",
		},
		{
			name: "Command with multiple arguments",
			text: "/echo hello world",
			want: "echo",
		},
		{
			name: "Command with bot mention",
			text: "/meme@maurynas_robot",
			want: "meme",
		},
		{
			name: "Command with bot mention and args",
			text: "/meme@maurynas_robot 5",
			want: "meme",
		},
		{
			name: "Not a command",
			text: "hello world",
			want: "",
		},
		{
			name: "Empty string",
			text: "",
			want: "",
		},
		{
			name: "Only slash",
			text: "/",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Message{Text: tt.text}
			if got := m.Command(); got != tt.want {
				t.Errorf("Message.Command() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMessageCommandArguments(t *testing.T) {
	tests := []struct {
		name string
		text string
		want string
	}{
		{
			name: "No arguments",
			text: "/start",
			want: "",
		},
		{
			name: "Single argument",
			text: "/price BTC",
			want: "BTC",
		},
		{
			name: "Multiple arguments",
			text: "/echo hello world",
			want: "hello world",
		},
		{
			name: "Not a command",
			text: "hello",
			want: "",
		},
		{
			name: "Empty string",
			text: "",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Message{Text: tt.text}
			if got := m.CommandArguments(); got != tt.want {
				t.Errorf("Message.CommandArguments() = %v, want %v", got, tt.want)
			}
		})
	}
}
