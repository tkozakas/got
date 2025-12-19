# Got

Telegram bot written in Go.

## Setup

Install [Task](https://taskfile.dev):
```bash
go install github.com/go-task/task/v3/cmd/task@latest
```

Create `.env`:
```bash
BOT_TOKEN=your_telegram_bot_token
DB_URL=postgres://postgres:postgres@localhost:5432/got?sslmode=disable
REDIS_ADDR=localhost:6379

# Optional
GROQ_API_KEY=your_groq_api_key
```

Run in production:
```bash
task prod
```

Local development:
```bash
task dev
```

## Commands

| Command | Description |
|---------|-------------|
| `/gpt <prompt>` | Chat with AI |
| `/gpt model` | List/select AI models |
| `/gpt image <prompt>` | Generate images |
| `/gpt memory` | Export chat history |
| `/gpt clear` | Clear chat history |
| `/tts <text>` | Text to speech |
| `/remind <time> <msg>` | Set a reminder |
| `/remind list` | List pending reminders |
| `/meme` | Random meme |
| `/sticker` | Random sticker |
| `/fact` | Random fact |
| `/stats` | Daily winner game |

## API Keys

- [Groq](https://console.groq.com) (optional, for AI chat)
