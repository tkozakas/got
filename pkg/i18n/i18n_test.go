package i18n

import (
	"os"
	"testing"
)

func TestTranslatorGet(t *testing.T) {
	setupTestTranslations(t)

	tests := []struct {
		name string
		lang string
		key  Key
		want string
	}{
		{
			name: "English welcome",
			lang: "en",
			key:  KeyWelcome,
			want: "Welcome! I am ready.",
		},
		{
			name: "Russian welcome",
			lang: "ru",
			key:  KeyWelcome,
			want: "Добро пожаловать! Я готов.",
		},
		{
			name: "Japanese welcome",
			lang: "ja",
			key:  KeyWelcome,
			want: "ようこそ！準備完了です。",
		},
		{
			name: "Lithuanian welcome",
			lang: "lt",
			key:  KeyWelcome,
			want: "Sveiki! Aš pasiruošęs.",
		},
		{
			name: "Unknown language falls back to English",
			lang: "xx",
			key:  KeyWelcome,
			want: "Welcome! I am ready.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := New(tt.lang)
			got := tr.Get(tt.key)
			assertEqual(t, got, tt.want)
		})
	}
}

func TestTranslatorLang(t *testing.T) {
	setupTestTranslations(t)

	tr := New("ru")
	assertEqual(t, tr.Lang(), "ru")
}

func TestTranslatorFallback(t *testing.T) {
	setupTestTranslations(t)

	tr := New("en")
	got := tr.Get("nonexistent_key")
	assertEqual(t, got, "nonexistent_key")
}

func TestNewWithTranslations(t *testing.T) {
	tests := []struct {
		name         string
		lang         string
		translations map[string]string
		key          Key
		want         string
	}{
		{
			name: "simpleTranslation",
			lang: "custom",
			translations: map[string]string{
				"welcome": "Custom Welcome",
			},
			key:  KeyWelcome,
			want: "Custom Welcome",
		},
		{
			name:         "emptyTranslations",
			lang:         "empty",
			translations: map[string]string{},
			key:          KeyWelcome,
			want:         "welcome",
		},
		{
			name: "multipleTranslations",
			lang: "multi",
			translations: map[string]string{
				"welcome":    "Hello",
				"fact_error": "Error occurred",
			},
			key:  KeyFactError,
			want: "Error occurred",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := NewWithTranslations(tt.lang, tt.translations)
			got := tr.Get(tt.key)
			assertEqual(t, got, tt.want)
			assertEqual(t, tr.Lang(), tt.lang)
		})
	}
}

func TestGetFallbackToDefaultLanguage(t *testing.T) {
	tests := []struct {
		name         string
		lang         string
		translations map[string]string
		fallback     map[string]string
		key          Key
		want         string
	}{
		{
			name:         "fallbackWhenKeyMissingInPrimary",
			lang:         "fr",
			translations: map[string]string{},
			fallback: map[string]string{
				"welcome": "Welcome from fallback",
			},
			key:  KeyWelcome,
			want: "Welcome from fallback",
		},
		{
			name: "primaryOverridesFallback",
			lang: "de",
			translations: map[string]string{
				"welcome": "Willkommen",
			},
			fallback: map[string]string{
				"welcome": "Welcome from fallback",
			},
			key:  KeyWelcome,
			want: "Willkommen",
		},
		{
			name: "fallbackUsedForMissingKey",
			lang: "es",
			translations: map[string]string{
				"fact_error": "Error de hecho",
			},
			fallback: map[string]string{
				"welcome": "Welcome fallback",
			},
			key:  KeyWelcome,
			want: "Welcome fallback",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &Translator{
				lang:         tt.lang,
				translations: tt.translations,
				fallback:     tt.fallback,
			}
			got := tr.Get(tt.key)
			assertEqual(t, got, tt.want)
		})
	}
}

func TestGetReturnsKeyWhenNotFound(t *testing.T) {
	tests := []struct {
		name         string
		translations map[string]string
		fallback     map[string]string
		key          Key
		want         string
	}{
		{
			name:         "emptyMaps",
			translations: map[string]string{},
			fallback:     map[string]string{},
			key:          KeyWelcome,
			want:         "welcome",
		},
		{
			name: "keyNotInEitherMap",
			translations: map[string]string{
				"other_key": "other value",
			},
			fallback: map[string]string{
				"another_key": "another value",
			},
			key:  KeyGptError,
			want: "gpt_error",
		},
		{
			name:         "nilMaps",
			translations: nil,
			fallback:     nil,
			key:          KeyFactError,
			want:         "fact_error",
		},
		{
			name:         "customKeyNotFound",
			translations: map[string]string{},
			fallback:     map[string]string{},
			key:          Key("custom_missing_key"),
			want:         "custom_missing_key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &Translator{
				lang:         "test",
				translations: tt.translations,
				fallback:     tt.fallback,
			}
			got := tr.Get(tt.key)
			assertEqual(t, got, tt.want)
		})
	}
}

func TestTranslatorAllKeys(t *testing.T) {
	setupTestTranslations(t)

	keys := getAllKeys()
	tr := New("en")

	for _, key := range keys {
		got := tr.Get(key)
		if got == string(key) {
			t.Errorf("Key %s not found in translations", key)
		}
	}
}

func getAllKeys() []Key {
	return []Key{
		KeyWelcome,
		KeyFactError,
		KeyNoFacts,
		KeyStickerError,
		KeyNoStickers,
		KeySubredditError,
		KeyGptUsage,
		KeyGptModelsHeader,
		KeyGptCleared,
		KeyGptError,
		KeyGptNoKey,
		KeyRemindListError,
		KeyRemindNoPending,
		KeyRemindUsage,
		KeyRemindInvalid,
		KeyRemindSuccess,
		KeyRemindHeader,
		KeyRemindFormat,
		KeyMemeError,
		KeyFactFormat,
		KeyReminderNotify,
	}
}

func assertEqual(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func setupTestTranslations(t *testing.T) {
	t.Helper()

	content := getTestTranslationsJSON()

	if err := os.MkdirAll("translations", 0755); err != nil {
		t.Fatalf("failed to create translations directory: %v", err)
	}

	if err := os.WriteFile(defaultFilePath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test translations file: %v", err)
	}

	t.Cleanup(func() {
		_ = os.Remove(defaultFilePath)
		_ = os.Remove("translations")
	})
}

func getTestTranslationsJSON() string {
	return `{
		"en": {
			"welcome": "Welcome! I am ready.",
			"fact_error": "Failed to fetch a fact.",
			"no_facts": "No facts available.",
			"sticker_error": "Failed to fetch a sticker.",
			"no_stickers": "No stickers available.",
			"subreddit_error": "Failed to fetch a subreddit.",
			"gpt_usage": "Usage: /gpt <prompt>",
			"gpt_models_header": "Available models:\n",
			"gpt_cleared": "Conversation history cleared.",
			"gpt_error": "Failed to get AI response.",
			"gpt_no_key": "GPT is not configured.",
			"remind_list_error": "Failed to list reminders.",
			"remind_no_pending": "No pending reminders.",
			"remind_usage": "Usage: /remind 5s Hello",
			"remind_invalid_time": "Invalid duration format.",
			"remind_success": "Reminder set for %s",
			"remind_header": "Pending reminders:\n",
			"remind_format": "- %s (at %s)\n",
			"meme_error": "Failed to fetch meme from r/",
			"fact_format": "Fact: %s",
			"reminder_notify": "REMINDER: %s"
		},
		"ru": {
			"welcome": "Добро пожаловать! Я готов.",
			"fact_error": "Не удалось получить факт.",
			"no_facts": "Нет доступных фактов.",
			"sticker_error": "Не удалось получить стикер.",
			"no_stickers": "Нет доступных стикеров.",
			"subreddit_error": "Не удалось получить сабреддит.",
			"gpt_usage": "Использование: /gpt <запрос>",
			"gpt_models_header": "Доступные модели:\n",
			"gpt_cleared": "История разговора очищена.",
			"gpt_error": "Не удалось получить ответ ИИ.",
			"gpt_no_key": "GPT не настроен.",
			"remind_list_error": "Не удалось получить список напоминаний.",
			"remind_no_pending": "Нет ожидающих напоминаний.",
			"remind_usage": "Использование: /remind 5s Привет",
			"remind_invalid_time": "Неверный формат времени.",
			"remind_success": "Напоминание установлено на %s",
			"remind_header": "Ожидающие напоминания:\n",
			"remind_format": "- %s (в %s)\n",
			"meme_error": "Не удалось получить мем из r/",
			"fact_format": "Факт: %s",
			"reminder_notify": "НАПОМИНАНИЕ: %s"
		},
		"lt": {
			"welcome": "Sveiki! Aš pasiruošęs.",
			"fact_error": "Nepavyko gauti fakto.",
			"no_facts": "Nėra faktų.",
			"sticker_error": "Nepavyko gauti lipduko.",
			"no_stickers": "Nėra lipdukų.",
			"subreddit_error": "Nepavyko gauti subreddit.",
			"gpt_usage": "Naudojimas: /gpt <užklausa>",
			"gpt_models_header": "Galimi modeliai:\n",
			"gpt_cleared": "Pokalbių istorija išvalyta.",
			"gpt_error": "Nepavyko gauti AI atsakymo.",
			"gpt_no_key": "GPT nesukonfigūruotas.",
			"remind_list_error": "Nepavyko gauti priminimų sąrašo.",
			"remind_no_pending": "Nėra laukiančių priminimų.",
			"remind_usage": "Naudojimas: /remind 5s Labas",
			"remind_invalid_time": "Neteisingas laiko formatas.",
			"remind_success": "Priminimas nustatytas po %s",
			"remind_header": "Laukiantys priminimai:\n",
			"remind_format": "- %s (%s)\n",
			"meme_error": "Nepavyko gauti memo iš r/",
			"fact_format": "Faktas: %s",
			"reminder_notify": "PRIMINIMAS: %s"
		},
		"ja": {
			"welcome": "ようこそ！準備完了です。",
			"fact_error": "ファクトの取得に失敗しました。",
			"no_facts": "ファクトがありません。",
			"sticker_error": "スティッカーの取得に失敗しました。",
			"no_stickers": "スティッカーがありません。",
			"subreddit_error": "サブレディットの取得に失敗しました。",
			"gpt_usage": "使用方法: /gpt <プロンプト>",
			"gpt_models_header": "利用可能なモデル:\n",
			"gpt_cleared": "会話履歴をクリアしました。",
			"gpt_error": "AI応答の取得に失敗しました。",
			"gpt_no_key": "GPTが設定されていません。",
			"remind_list_error": "リマインダー一覧の取得に失敗しました。",
			"remind_no_pending": "保留中のリマインダーはありません。",
			"remind_usage": "使用方法: /remind 5s こんにちは",
			"remind_invalid_time": "無効な時間形式です。",
			"remind_success": "%s後にリマインダーを設定しました",
			"remind_header": "保留中のリマインダー:\n",
			"remind_format": "- %s (%s)\n",
			"meme_error": "r/からミームを取得できませんでした",
			"fact_format": "ファクト: %s",
			"reminder_notify": "リマインダー: %s"
		}
	}`
}
