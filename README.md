# Got

Telegram bot written in Go.

## Setup

Install tools with [mise](https://mise.jdx.dev):
```bash
mise install
eval "$(mise activate bash)"  # or zsh
```

Create `.env`:
```bash
BOT_TOKEN=your_telegram_bot_token
DB_URL=postgres://postgres:postgres@localhost:5432/got?sslmode=disable
REDIS_ADDR=localhost:6379

# Optional
GROQ_API_KEY=your_groq_api_key
ADMIN_PASS=your_secret_password
ROULETTE_SENTENCES_PATH=roulette_sentences.json
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
| `/gpt model` | List AI models |
| `/gpt model <name\|number>` | Select AI model by name or number |
| `/gpt image <prompt>` | Generate images |
| `/gpt memory` | Export chat history |
| `/gpt clear` | Clear chat history |
| `/tts <text>` | Text to speech |
| `/remind <time> <msg>` | Set a reminder |
| `/remind list` | List pending reminders |
| `/remind delete <id>` | Delete a reminder |
| `/meme` | Random meme |
| `/meme <count>` | Multiple memes (1-5) |
| `/meme add <subreddit>` | Add subreddit |
| `/meme list` | List subreddits |
| `/sticker` | Random sticker |
| `/sticker add` | Add sticker (reply to sticker) |
| `/sticker add <set_name>` | Add entire sticker set |
| `/sticker list` | List sticker sets |
| `/fact` | Random fact |
| `/fact add <text>` | Add a fact |
| `/roulette` | Daily winner roulette |
| `/roulette stats [year]` | View stats |
| `/roulette all` | All-time stats |
| `/lang` | Show current language |
| `/lang <code>` | Set chat language (en, ru, lt, ja) |
| `/admin login <pass>` | Login as admin (DM only) |
| `/admin reset` | Reset today's winner |

## Configuration

Commands can be customized via environment variables:

```bash
CMD_ROULETTE=pidor    # Rename /roulette to /pidor
CMD_GPT=ai            # Rename /gpt to /ai
CMD_ADMIN=sudo        # Rename /admin to /sudo
CMD_LANG=language     # Rename /lang to /language
```

Other configuration:

```bash
ROULETTE_SENTENCES_PATH=custom_sentences.json  # Custom roulette animations
```

## Admin Commands

Admin commands allow you to manage the bot securely:

1. **Login** - DM the bot privately with `/admin login <password>`
2. **Use admin commands** - Once logged in, use `/admin reset` in any chat

Sessions expire after 12 hours.

## API Keys

- [Groq](https://console.groq.com) (optional, for AI chat)
