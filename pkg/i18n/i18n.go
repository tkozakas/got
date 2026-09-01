package i18n

import (
	"encoding/json"
	"log/slog"
	"os"
)

const (
	defaultLang     = "en"
	defaultFilePath = "translations/template_message.json"
)

const (
	KeyWelcome            Key = "welcome"
	KeyFactError          Key = "fact_error"
	KeyNoFacts            Key = "no_facts"
	KeyStickerError       Key = "sticker_error"
	KeyNoStickers         Key = "no_stickers"
	KeySubredditError     Key = "subreddit_error"
	KeyGptUsage           Key = "gpt_usage"
	KeyGptModelsHeader    Key = "gpt_models_header"
	KeyGptCleared         Key = "gpt_cleared"
	KeyGptError           Key = "gpt_error"
	KeyGptNoKey           Key = "gpt_no_key"
	KeyRemindListError    Key = "remind_list_error"
	KeyRemindNoPending    Key = "remind_no_pending"
	KeyRemindUsage        Key = "remind_usage"
	KeyRemindInvalid      Key = "remind_invalid_time"
	KeyRemindSuccess      Key = "remind_success"
	KeyRemindHeader       Key = "remind_header"
	KeyRemindFormat       Key = "remind_format"
	KeyRemindDeleted      Key = "remind_deleted"
	KeyRemindDeleteUsage  Key = "remind_delete_usage"
	KeyRemindDeleteError  Key = "remind_delete_error"
	KeyMemeError          Key = "meme_error"
	KeyMemeUsage          Key = "meme_usage"
	KeyMemeAdded          Key = "meme_added"
	KeyMemeRemoved        Key = "meme_removed"
	KeyMemeListHeader     Key = "meme_list_header"
	KeyMemeCountInvalid   Key = "meme_count_invalid"
	KeyFactFormat         Key = "fact_format"
	KeyFactUsage          Key = "fact_usage"
	KeyFactAdded          Key = "fact_added"
	KeyStickerUsage       Key = "sticker_usage"
	KeyStickerAdded       Key = "sticker_added"
	KeyStickerListHeader  Key = "sticker_list_header"
	KeyStickerRemoved     Key = "sticker_removed"
	KeyStickerRemoveUsage Key = "sticker_remove_usage"
	KeyStickerCount       Key = "sticker_count"
	KeyStickerSetNotFound Key = "sticker_set_not_found"
	KeyStickerSetAdded    Key = "sticker_set_added"
	KeyStickerSetRemoved  Key = "sticker_set_removed"
	KeyReminderNotify     Key = "reminder_notify"

	KeyCmdStart    Key = "cmd_start"
	KeyCmdHelp     Key = "cmd_help"
	KeyCmdGpt      Key = "cmd_gpt"
	KeyCmdRemind   Key = "cmd_remind"
	KeyCmdMeme     Key = "cmd_meme"
	KeyCmdSticker  Key = "cmd_sticker"
	KeyCmdFact     Key = "cmd_fact"
	KeyCmdRoulette Key = "cmd_roulette"

	KeyHelpHeader Key = "help_header"

	KeyRouletteAlias        Key = "roulette_alias"
	KeyRouletteNoStats      Key = "roulette_no_stats"
	KeyRouletteHeader       Key = "roulette_header"
	KeyRouletteHeaderAll    Key = "roulette_header_all"
	KeyRouletteFooter       Key = "roulette_footer"
	KeyRouletteUser         Key = "roulette_user"
	KeyRouletteWinnerExists Key = "roulette_winner_exists"
	KeyRouletteWinnerNew    Key = "roulette_winner_new"
	KeyRouletteAutoWinner   Key = "roulette_auto_winner"
	KeyRouletteNoUsers      Key = "roulette_no_users"
	KeyRouletteUsage        Key = "roulette_usage"

	KeyCmdTts   Key = "cmd_tts"
	KeyTtsUsage Key = "tts_usage"
	KeyTtsError Key = "tts_error"

	KeyGptImageUsage Key = "gpt_image_usage"
	KeyGptImageError Key = "gpt_image_error"

	KeyGptMemoryHeader  Key = "gpt_memory_header"
	KeyGptMemoryStats   Key = "gpt_memory_stats"
	KeyGptMemoryEmpty   Key = "gpt_memory_empty"
	KeyGptMemoryNoRedis Key = "gpt_memory_no_redis"
	KeyGptMemoryCaption Key = "gpt_memory_caption"
	KeyGptModelSet      Key = "gpt_model_set"
	KeyGptModelInvalid  Key = "gpt_model_invalid"

	KeyAdminUnauthorized Key = "admin_unauthorized"
	KeyAdminUsage        Key = "admin_usage"
	KeyAdminLoginSuccess Key = "admin_login_success"
	KeyAdminNotLoggedIn  Key = "admin_not_logged_in"
	KeyAdminResetSuccess Key = "admin_reset_success"
	KeyAdminResetError   Key = "admin_reset_error"
	KeyAdminNoPass       Key = "admin_no_pass"
	KeyAdminDMOnly       Key = "admin_dm_only"

	KeyCmdLang     Key = "cmd_lang"
	KeyLangUsage   Key = "lang_usage"
	KeyLangSet     Key = "lang_set"
	KeyLangCurrent Key = "lang_current"
	KeyLangList    Key = "lang_list"
)

type Key string

type Translator struct {
	lang         string
	translations map[string]string
	fallback     map[string]string
}

func New(lang string) *Translator {
	all := loadTranslationsFile(defaultFilePath)

	translations := all[lang]
	if translations == nil {
		slog.Warn("language not found, using default", "lang", lang, "default", defaultLang)
		translations = all[defaultLang]
	}

	return &Translator{
		lang:         lang,
		translations: translations,
		fallback:     all[defaultLang],
	}
}

func NewWithTranslations(lang string, translations map[string]string) *Translator {
	return &Translator{
		lang:         lang,
		translations: translations,
		fallback:     translations,
	}
}

func (t *Translator) Get(key Key) string {
	if msg, ok := t.translations[string(key)]; ok {
		return msg
	}
	if msg, ok := t.fallback[string(key)]; ok {
		return msg
	}
	return string(key)
}

func (t *Translator) Lang() string {
	return t.lang
}

func loadTranslationsFile(path string) map[string]map[string]string {
	data, err := os.ReadFile(path)
	if err != nil {
		slog.Error("failed to read translations file", "path", path, "error", err)
		return make(map[string]map[string]string)
	}

	var translations map[string]map[string]string
	if err := json.Unmarshal(data, &translations); err != nil {
		slog.Error("failed to parse translations file", "path", path, "error", err)
		return make(map[string]map[string]string)
	}

	return translations
}
