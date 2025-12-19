package telegram

import (
	"context"
	"fmt"
	"got/internal/app"
	"got/internal/app/model"
	"got/internal/groq"
	"got/internal/redis"
	"got/internal/tts"
	"got/pkg/i18n"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const defaultSubreddit = "programmerhumor"

type SubCommand string

const (
	SubCommandList   SubCommand = "list"
	SubCommandAdd    SubCommand = "add"
	SubCommandRemove SubCommand = "remove"
	SubCommandDelete SubCommand = "delete"
	SubCommandModel  SubCommand = "model"
	SubCommandClear  SubCommand = "clear"
	SubCommandForget SubCommand = "forget"
	SubCommandAll    SubCommand = "all"
	SubCommandStats  SubCommand = "stats"
	SubCommandMemory SubCommand = "memory"
	SubCommandImage  SubCommand = "image"
)

const (
	ActionTyping          = "typing"
	ActionUploadPhoto     = "upload_photo"
	ActionRecordVoice     = "record_voice"
	ActionUploadVoice     = "upload_voice"
	ActionUploadDocument  = "upload_document"
	ActionUploadVideo     = "upload_video"
	ActionRecordVideoNote = "record_video_note"
	ActionUploadVideoNote = "upload_video_note"
)

type BotHandlers struct {
	client  *Client
	service *app.Service
	gpt     *groq.Client
	cache   *redis.Client
	t       *i18n.Translator
	tts     *tts.Client
}

func NewBotHandlers(client *Client, service *app.Service, gpt *groq.Client, cache *redis.Client, t *i18n.Translator, tts *tts.Client) *BotHandlers {
	return &BotHandlers{
		client:  client,
		service: service,
		gpt:     gpt,
		cache:   cache,
		t:       t,
		tts:     tts,
	}
}

func (h *BotHandlers) HandleStart(ctx context.Context, update *Update) error {
	chatID := update.Message.Chat.ID

	if update.Message.Chat != nil {
		_ = h.service.RegisterChat(ctx, &model.Chat{
			ChatID:   chatID,
			ChatName: update.Message.Chat.Title,
		})
	}

	if update.Message.From != nil {
		_ = h.service.RegisterUser(ctx, &model.User{
			UserID:   update.Message.From.ID,
			Username: update.Message.From.UserName,
		}, chatID)
	}

	return h.client.SendMessage(chatID, h.t.Get(i18n.KeyWelcome))
}

func (h *BotHandlers) HandleHelp(ctx context.Context, update *Update) error {
	chatID := update.Message.Chat.ID

	commands := []struct {
		cmd  string
		desc i18n.Key
	}{
		{"start", i18n.KeyCmdStart},
		{"help", i18n.KeyCmdHelp},
		{"gpt", i18n.KeyCmdGpt},
		{"remind", i18n.KeyCmdRemind},
		{"meme", i18n.KeyCmdMeme},
		{"sticker", i18n.KeyCmdSticker},
		{"fact", i18n.KeyCmdFact},
		{"stats", i18n.KeyCmdStats},
	}

	var sb strings.Builder
	sb.WriteString(h.t.Get(i18n.KeyHelpHeader))
	for _, c := range commands {
		sb.WriteString(fmt.Sprintf("/%s — %s\n", c.cmd, h.t.Get(c.desc)))
	}

	return h.client.SendMessage(chatID, sb.String())
}

func (h *BotHandlers) HandleFact(ctx context.Context, update *Update) error {
	chatID := update.Message.Chat.ID
	args := strings.TrimSpace(update.Message.CommandArguments())
	parts := strings.SplitN(args, " ", 2)

	if len(parts) > 0 && SubCommand(parts[0]) == SubCommandAdd {
		if len(parts) < 2 {
			return h.client.SendMessage(chatID, h.t.Get(i18n.KeyFactUsage))
		}
		if err := h.service.AddFact(ctx, parts[1], chatID); err != nil {
			return h.client.SendMessage(chatID, h.t.Get(i18n.KeyFactError))
		}
		return h.client.SendMessage(chatID, h.t.Get(i18n.KeyFactAdded))
	}

	fact, err := h.service.GetRandomFact(ctx, chatID)
	if err != nil {
		return h.client.SendMessage(chatID, h.t.Get(i18n.KeyFactError))
	}
	if fact == nil {
		return h.client.SendMessage(chatID, h.t.Get(i18n.KeyNoFacts))
	}
	return h.client.SendMessage(chatID, fmt.Sprintf(h.t.Get(i18n.KeyFactFormat), fact.Comment))
}

func (h *BotHandlers) HandleSticker(ctx context.Context, update *Update) error {
	chatID := update.Message.Chat.ID
	args := strings.TrimSpace(update.Message.CommandArguments())

	switch SubCommand(args) {
	case SubCommandAdd:
		if update.Message.ReplyToMessage == nil || update.Message.ReplyToMessage.Sticker == nil {
			return h.client.SendMessage(chatID, h.t.Get(i18n.KeyStickerUsage))
		}
		fileID := update.Message.ReplyToMessage.Sticker.FileID
		if err := h.service.AddSticker(ctx, fileID, chatID); err != nil {
			return h.client.SendMessage(chatID, h.t.Get(i18n.KeyStickerError))
		}
		return h.client.SendMessage(chatID, h.t.Get(i18n.KeyStickerAdded))

	case SubCommandRemove:
		if update.Message.ReplyToMessage == nil || update.Message.ReplyToMessage.Sticker == nil {
			return h.client.SendMessage(chatID, h.t.Get(i18n.KeyStickerRemoveUsage))
		}
		fileID := update.Message.ReplyToMessage.Sticker.FileID
		if err := h.service.RemoveSticker(ctx, fileID, chatID); err != nil {
			return h.client.SendMessage(chatID, h.t.Get(i18n.KeyStickerError))
		}
		return h.client.SendMessage(chatID, h.t.Get(i18n.KeyStickerRemoved))

	case SubCommandList:
		stickers, err := h.service.ListStickers(ctx, chatID)
		if err != nil {
			return h.client.SendMessage(chatID, h.t.Get(i18n.KeyStickerError))
		}
		return h.client.SendMessage(chatID, fmt.Sprintf(h.t.Get(i18n.KeyStickerListHeader), len(stickers)))

	default:
		sticker, err := h.service.GetRandomSticker(ctx, chatID)
		if err != nil {
			return h.client.SendMessage(chatID, h.t.Get(i18n.KeyStickerError))
		}
		if sticker == nil {
			return h.client.SendMessage(chatID, h.t.Get(i18n.KeyNoStickers))
		}
		return h.client.SendSticker(chatID, sticker.FileID)
	}
}

func (h *BotHandlers) HandleMeme(ctx context.Context, update *Update) error {
	chatID := update.Message.Chat.ID
	args := strings.TrimSpace(update.Message.CommandArguments())
	parts := strings.Fields(args)

	if len(parts) > 0 {
		switch SubCommand(parts[0]) {
		case SubCommandAdd:
			if len(parts) < 2 {
				return h.client.SendMessage(chatID, h.t.Get(i18n.KeyMemeUsage))
			}
			if err := h.service.AddSubreddit(ctx, parts[1], chatID); err != nil {
				return h.client.SendMessage(chatID, h.t.Get(i18n.KeySubredditError))
			}
			return h.client.SendMessage(chatID, fmt.Sprintf(h.t.Get(i18n.KeyMemeAdded), parts[1]))

		case SubCommandRemove:
			if len(parts) < 2 {
				return h.client.SendMessage(chatID, h.t.Get(i18n.KeyMemeUsage))
			}
			if err := h.service.RemoveSubreddit(ctx, parts[1], chatID); err != nil {
				return h.client.SendMessage(chatID, h.t.Get(i18n.KeySubredditError))
			}
			return h.client.SendMessage(chatID, fmt.Sprintf(h.t.Get(i18n.KeyMemeRemoved), parts[1]))

		case SubCommandList:
			subs, err := h.service.ListSubreddits(ctx, chatID)
			if err != nil {
				return h.client.SendMessage(chatID, h.t.Get(i18n.KeySubredditError))
			}
			return h.client.SendMessage(chatID, h.formatSubredditList(subs))
		}
	}

	count := 1
	var explicitSubreddit string

	if len(parts) > 0 {
		if n, err := strconv.Atoi(parts[0]); err == nil {
			if n < 1 || n > 5 {
				return h.client.SendMessage(chatID, h.t.Get(i18n.KeyMemeCountInvalid))
			}
			count = n
			if len(parts) > 1 {
				explicitSubreddit = parts[1]
			}
		} else {
			explicitSubreddit = parts[0]
		}
	}

	var subName string
	if explicitSubreddit != "" {
		subName = explicitSubreddit
	} else {
		sub, err := h.service.GetRandomSubreddit(ctx, chatID)
		if err != nil {
			return h.client.SendMessage(chatID, h.t.Get(i18n.KeySubredditError))
		}
		if sub != nil {
			subName = sub.Name
		} else {
			subName = defaultSubreddit
		}
	}

	_ = h.client.SendChatAction(chatID, ActionUploadPhoto)

	memes, err := h.fetchMemes(ctx, subName, count)
	if err != nil || len(memes) == 0 {
		return h.client.SendMessage(chatID, h.t.Get(i18n.KeyMemeError)+subName)
	}

	return h.sendMemes(chatID, memes)
}

func (h *BotHandlers) HandleGPT(ctx context.Context, update *Update) error {
	chatID := update.Message.Chat.ID

	if h.gpt == nil {
		return h.client.SendMessage(chatID, h.t.Get(i18n.KeyGptNoKey))
	}

	args := update.Message.CommandArguments()
	if args == "" {
		return h.client.SendMessage(chatID, h.t.Get(i18n.KeyGptUsage))
	}

	parts := strings.SplitN(args, " ", 2)
	subCmd := SubCommand(parts[0])

	switch subCmd {
	case SubCommandModel:
		if len(parts) > 1 && strings.TrimSpace(parts[1]) != "" {
			return h.handleGPTSetModel(ctx, chatID, strings.TrimSpace(parts[1]))
		}
		return h.handleGPTModels(ctx, chatID)
	case SubCommandClear, SubCommandForget:
		return h.handleGPTClear(ctx, chatID)
	case SubCommandMemory:
		return h.handleGPTMemory(ctx, chatID)
	case SubCommandImage:
		return h.handleGPTImage(ctx, chatID, parts)
	default:
		username := ""
		if update.Message.From != nil {
			username = update.Message.From.UserName
		}
		return h.handleGPTChat(ctx, chatID, username, args)
	}
}

func (h *BotHandlers) HandleRemind(ctx context.Context, update *Update) error {
	chatID := update.Message.Chat.ID
	args := update.Message.CommandArguments()
	parts := strings.SplitN(args, " ", 2)

	if len(parts) == 0 {
		return h.client.SendMessage(chatID, h.t.Get(i18n.KeyRemindUsage))
	}

	switch SubCommand(parts[0]) {
	case SubCommandList:
		return h.handleRemindList(ctx, chatID)
	case SubCommandDelete:
		return h.handleRemindDelete(ctx, chatID, parts)
	default:
		return h.handleRemindAdd(ctx, chatID, update.Message.From.ID, parts)
	}
}

func (h *BotHandlers) handleRemindList(ctx context.Context, chatID int64) error {
	reminders, err := h.service.GetPendingReminders(ctx, chatID)
	if err != nil {
		return h.client.SendMessage(chatID, h.t.Get(i18n.KeyRemindListError))
	}
	if len(reminders) == 0 {
		return h.client.SendMessage(chatID, h.t.Get(i18n.KeyRemindNoPending))
	}

	return h.client.SendMessage(chatID, h.formatReminders(reminders))
}

func (h *BotHandlers) handleRemindDelete(ctx context.Context, chatID int64, parts []string) error {
	if len(parts) < 2 {
		return h.client.SendMessage(chatID, h.t.Get(i18n.KeyRemindDeleteUsage))
	}

	reminderID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return h.client.SendMessage(chatID, h.t.Get(i18n.KeyRemindDeleteUsage))
	}

	if err := h.service.DeleteReminder(ctx, reminderID, chatID); err != nil {
		return h.client.SendMessage(chatID, h.t.Get(i18n.KeyRemindDeleteError))
	}

	return h.client.SendMessage(chatID, h.t.Get(i18n.KeyRemindDeleted))
}

func (h *BotHandlers) handleRemindAdd(ctx context.Context, chatID int64, userID int64, parts []string) error {
	if len(parts) < 2 {
		return h.client.SendMessage(chatID, h.t.Get(i18n.KeyRemindUsage))
	}

	duration, err := ParseDuration(parts[0])
	if err != nil {
		return h.client.SendMessage(chatID, h.t.Get(i18n.KeyRemindInvalid))
	}

	if err := h.service.AddReminder(ctx, chatID, userID, parts[1], duration); err != nil {
		return h.client.SendMessage(chatID, fmt.Sprintf("Error: %v", err))
	}

	return h.client.SendMessage(chatID, fmt.Sprintf(h.t.Get(i18n.KeyRemindSuccess), duration))
}

func (h *BotHandlers) formatReminders(reminders []*model.Reminder) string {
	var sb strings.Builder
	sb.WriteString(h.t.Get(i18n.KeyRemindHeader))
	for _, r := range reminders {
		sb.WriteString(fmt.Sprintf(h.t.Get(i18n.KeyRemindFormat), r.ReminderID, r.Message, r.RemindAt.Format(time.RFC822)))
	}
	return sb.String()
}

func (h *BotHandlers) HandleStats(ctx context.Context, update *Update) error {
	chatID := update.Message.Chat.ID
	userID := update.Message.From.ID
	currentYear := time.Now().Year()

	_ = h.service.RegisterUser(ctx, &model.User{
		UserID:   userID,
		Username: update.Message.From.UserName,
	}, chatID)

	_, _ = h.service.GetOrCreateStat(ctx, userID, chatID, currentYear)

	args := strings.TrimSpace(update.Message.CommandArguments())
	parts := strings.Fields(args)

	if len(parts) == 0 {
		return h.handleStatsRoulette(ctx, chatID, currentYear)
	}

	switch SubCommand(parts[0]) {
	case SubCommandAll:
		return h.handleStatsAll(ctx, chatID)
	case SubCommandStats:
		year := currentYear
		if len(parts) > 1 {
			if y, err := strconv.Atoi(parts[1]); err == nil {
				year = y
			}
		}
		return h.handleStatsYear(ctx, chatID, year)
	default:
		if year, err := strconv.Atoi(parts[0]); err == nil {
			return h.handleStatsYear(ctx, chatID, year)
		}
		return h.client.SendMessage(chatID, h.t.Get(i18n.KeyStatsUsage))
	}
}

func (h *BotHandlers) handleStatsRoulette(ctx context.Context, chatID int64, year int) error {
	alias := h.t.Get(i18n.KeyStatsAlias)

	winner, err := h.service.GetTodayWinner(ctx, chatID, year)
	if err != nil {
		return h.client.SendMessage(chatID, h.t.Get(i18n.KeyStatsNoStats))
	}

	if winner != nil {
		msg := fmt.Sprintf(h.t.Get(i18n.KeyStatsWinnerExists), alias, h.formatUser(winner.User), winner.Score)
		return h.client.SendMessage(chatID, msg)
	}

	winner, err = h.service.SelectRandomWinner(ctx, chatID, year)
	if err != nil || winner == nil {
		return h.client.SendMessage(chatID, h.t.Get(i18n.KeyStatsNoUsers))
	}

	msg := fmt.Sprintf(h.t.Get(i18n.KeyStatsWinnerNew), alias, h.formatUser(winner.User))
	return h.client.SendMessage(chatID, msg)
}

func (h *BotHandlers) handleStatsYear(ctx context.Context, chatID int64, year int) error {
	stats, err := h.service.GetStatsByYear(ctx, chatID, year)
	if err != nil || len(stats) == 0 {
		return h.client.SendMessage(chatID, h.t.Get(i18n.KeyStatsNoStats))
	}

	return h.client.SendMessage(chatID, h.formatStats(stats, fmt.Sprintf(h.t.Get(i18n.KeyStatsHeader), year)))
}

func (h *BotHandlers) handleStatsAll(ctx context.Context, chatID int64) error {
	stats, err := h.service.GetAllStats(ctx, chatID)
	if err != nil || len(stats) == 0 {
		return h.client.SendMessage(chatID, h.t.Get(i18n.KeyStatsNoStats))
	}

	aggregated := h.aggregateStats(stats)
	return h.client.SendMessage(chatID, h.formatStats(aggregated, h.t.Get(i18n.KeyStatsHeaderAll)))
}

func (h *BotHandlers) handleGPTModels(ctx context.Context, chatID int64) error {
	currentModel := h.getChatModel(ctx, chatID)
	models := h.fetchModelsWithFallback(ctx)

	var sb strings.Builder
	sb.WriteString(h.t.Get(i18n.KeyGptModelsHeader))
	for _, m := range models {
		prefix := "  "
		if m == currentModel {
			prefix = "→ "
		}
		sb.WriteString(prefix + m + "\n")
	}
	return h.client.SendMessage(chatID, sb.String())
}

func (h *BotHandlers) fetchModelsWithFallback(ctx context.Context) []string {
	models, err := h.gpt.FetchModels(ctx)
	if err != nil || len(models) == 0 {
		return h.gpt.ListModels()
	}
	return models
}

func (h *BotHandlers) handleGPTSetModel(ctx context.Context, chatID int64, modelName string) error {
	if err := h.gpt.ValidateModel(modelName); err != nil {
		var sb strings.Builder
		sb.WriteString(h.t.Get(i18n.KeyGptModelInvalid))
		for _, m := range h.gpt.ListModels() {
			sb.WriteString("- " + m + "\n")
		}
		return h.client.SendMessage(chatID, sb.String())
	}

	if h.cache != nil {
		_ = h.cache.SetModel(ctx, chatID, modelName)
	}

	return h.client.SendMessage(chatID, fmt.Sprintf(h.t.Get(i18n.KeyGptModelSet), modelName))
}

func (h *BotHandlers) getChatModel(ctx context.Context, chatID int64) string {
	if h.cache == nil {
		return ""
	}
	model, _ := h.cache.GetModel(ctx, chatID)
	return model
}

func (h *BotHandlers) startTyping(ctx context.Context, chatID int64, action string) *TypingIndicator {
	indicator := NewTypingIndicator(h.client, chatID)
	indicator.Start(ctx, action)
	return indicator
}

func (h *BotHandlers) handleGPTClear(ctx context.Context, chatID int64) error {
	if h.cache != nil {
		_ = h.cache.ClearHistory(ctx, chatID)
	}
	return h.client.SendMessage(chatID, h.t.Get(i18n.KeyGptCleared))
}

func (h *BotHandlers) handleGPTMemory(ctx context.Context, chatID int64) error {
	if h.cache == nil {
		return h.client.SendMessage(chatID, h.t.Get(i18n.KeyGptMemoryNoRedis))
	}

	history, err := h.cache.GetHistory(ctx, chatID)
	if err != nil || len(history) == 0 {
		return h.client.SendMessage(chatID, h.t.Get(i18n.KeyGptMemoryEmpty))
	}

	_ = h.client.SendChatAction(chatID, ActionUploadDocument)

	content := formatHistoryAsText(history)
	filename := fmt.Sprintf("chat_history_%d.txt", chatID)
	caption := h.t.Get(i18n.KeyGptMemoryCaption)

	return h.client.SendDocument(chatID, []byte(content), filename, caption)
}

func formatHistoryAsText(history []groq.Message) string {
	var sb strings.Builder
	for _, msg := range history {
		sb.WriteString(msg.Role)
		sb.WriteString(": ")
		sb.WriteString(msg.Content)
		sb.WriteString("\n")
	}
	return sb.String()
}

func (h *BotHandlers) handleGPTImage(ctx context.Context, chatID int64, parts []string) error {
	if len(parts) < 2 || strings.TrimSpace(parts[1]) == "" {
		return h.client.SendMessage(chatID, h.t.Get(i18n.KeyGptImageUsage))
	}

	typing := h.startTyping(ctx, chatID, ActionUploadPhoto)
	defer typing.Stop()

	prompt := strings.TrimSpace(parts[1])
	imageURL := buildImageURL(prompt)

	return h.client.SendPhoto(chatID, imageURL, prompt)
}

func buildImageURL(prompt string) string {
	seed := time.Now().UnixNano() % 1000000
	return fmt.Sprintf(
		"https://image.pollinations.ai/prompt/%s?width=1024&height=1024&seed=%d&nologo=true&enhance=true",
		url.QueryEscape(prompt),
		seed,
	)
}

func (h *BotHandlers) handleGPTChat(ctx context.Context, chatID int64, username string, prompt string) error {
	typing := h.startTyping(ctx, chatID, ActionTyping)
	defer typing.Stop()

	formattedPrompt := formatPromptWithUsername(username, prompt)
	model := h.getChatModel(ctx, chatID)

	var history []groq.Message
	if h.cache != nil {
		history, _ = h.cache.GetHistory(ctx, chatID)
	}

	response, err := h.gpt.ChatWithModel(ctx, formattedPrompt, history, model)
	if err != nil {
		return h.client.SendMessage(chatID, h.t.Get(i18n.KeyGptError))
	}

	if h.cache != nil {
		history = append(history, groq.Message{Role: "user", Content: formattedPrompt})
		history = append(history, groq.Message{Role: "assistant", Content: response})
		_ = h.cache.SaveHistory(ctx, chatID, history)
	}

	return h.client.SendMessage(chatID, response)
}

func formatPromptWithUsername(username string, prompt string) string {
	if strings.TrimSpace(username) == "" {
		return prompt
	}
	return username + ": " + prompt
}

func (h *BotHandlers) formatUser(user *model.User) string {
	return fmt.Sprintf("[%s](tg://user?id=%d)", user.Username, user.UserID)
}

func (h *BotHandlers) formatStats(stats []*model.Stat, header string) string {
	var sb strings.Builder
	sb.WriteString("*" + header + "*\n\n")

	for i, stat := range stats {
		username := stat.User.Username
		if stat.IsWinner {
			username = "👑 " + username
		}
		sb.WriteString(fmt.Sprintf(h.t.Get(i18n.KeyStatsUser), i+1, username, stat.Score))
		sb.WriteString("\n")
	}

	sb.WriteString(fmt.Sprintf("\n*%s*", fmt.Sprintf(h.t.Get(i18n.KeyStatsFooter), len(stats))))
	return sb.String()
}

func (h *BotHandlers) formatSubredditList(subs []*model.Subreddit) string {
	var sb strings.Builder
	sb.WriteString(h.t.Get(i18n.KeyMemeListHeader))
	for _, s := range subs {
		sb.WriteString("- r/" + s.Name + "\n")
	}
	return sb.String()
}

func (h *BotHandlers) aggregateStats(stats []*model.Stat) []*model.Stat {
	userStats := make(map[int64]*model.Stat)

	for _, s := range stats {
		if existing, ok := userStats[s.User.UserID]; ok {
			existing.Score += s.Score
			if s.IsWinner {
				existing.IsWinner = true
			}
		} else {
			userStats[s.User.UserID] = &model.Stat{
				User:     s.User,
				Chat:     s.Chat,
				Score:    s.Score,
				IsWinner: s.IsWinner,
			}
		}
	}

	result := make([]*model.Stat, 0, len(userStats))
	for _, s := range userStats {
		result = append(result, s)
	}

	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if result[j].Score > result[i].Score {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	return result
}

func (h *BotHandlers) HandleTTS(ctx context.Context, update *Update) error {
	chatID := update.Message.Chat.ID
	text := update.Message.CommandArguments()

	if text == "" {
		return h.client.SendMessage(chatID, h.t.Get(i18n.KeyTtsUsage))
	}

	if h.tts == nil {
		return h.client.SendMessage(chatID, h.t.Get(i18n.KeyTtsError))
	}

	typing := h.startTyping(ctx, chatID, ActionRecordVoice)
	defer typing.Stop()

	audioData, err := h.tts.GenerateSpeech(ctx, text)
	if err != nil {
		return h.client.SendMessage(chatID, h.t.Get(i18n.KeyTtsError))
	}

	return h.client.SendVoice(chatID, audioData, "speech.mp3")
}

func (h *BotHandlers) sendMemes(chatID int64, memes []model.RedditMeme) error {
	var photos []model.RedditMeme
	var gifs []model.RedditMeme

	for _, meme := range memes {
		if isAnimatedURL(meme.URL) {
			gifs = append(gifs, meme)
		} else {
			photos = append(photos, meme)
		}
	}

	if len(photos) > 1 {
		media := make([]InputMediaPhoto, len(photos))
		for i, meme := range photos {
			media[i] = InputMediaPhoto{
				Type:    "photo",
				Media:   meme.URL,
				Caption: meme.Title,
			}
		}
		if err := h.client.SendMediaGroup(chatID, media); err != nil {
			return err
		}
	} else if len(photos) == 1 {
		if err := h.client.SendPhoto(chatID, photos[0].URL, photos[0].Title); err != nil {
			return err
		}
	}

	for _, gif := range gifs {
		if err := h.client.SendAnimation(chatID, gif.URL, gif.Title); err != nil {
			return err
		}
	}

	return nil
}

func isAnimatedURL(url string) bool {
	lowerURL := strings.ToLower(url)
	return strings.HasSuffix(lowerURL, ".gif") ||
		strings.Contains(lowerURL, ".gif?") ||
		strings.HasSuffix(lowerURL, ".mp4") ||
		strings.Contains(lowerURL, ".mp4?")
}
