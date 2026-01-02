# Got

Telegram bot written in Go.

## Setup

Create `.env`:
```bash
BOT_TOKEN=your_telegram_bot_token
GROQ_API_KEY=your_groq_api_key  # optional, for AI
ADMIN_PASS=your_secret_password  # optional
```

Run:
```bash
docker compose --profile prod up -d
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
| `/remind <time> <msg>` | Set reminder |
| `/remind list` | List reminders |
| `/meme` | Random meme |
| `/meme add <subreddit>` | Add subreddit |
| `/sticker` | Random sticker |
| `/sticker add` | Add sticker (reply to sticker) |
| `/fact` | Random fact |
| `/fact add <text>` | Add a fact |
| `/roulette` | Daily winner roulette |
| `/roulette stats` | View stats |
| `/lang <code>` | Set language (en, ru, lt, ja) |
| `/admin login <pass>` | Admin login (DM only) |
