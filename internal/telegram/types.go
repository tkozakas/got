package telegram

type Update struct {
	UpdateID int      `json:"update_id"`
	Message  *Message `json:"message"`
}

type Message struct {
	MessageID      int      `json:"message_id"`
	From           *User    `json:"from"`
	Chat           *Chat    `json:"chat"`
	Text           string   `json:"text"`
	ReplyToMessage *Message `json:"reply_to_message"`
	Sticker        *Sticker `json:"sticker"`
}

type User struct {
	ID        int64  `json:"id"`
	IsBot     bool   `json:"is_bot"`
	FirstName string `json:"first_name"`
	UserName  string `json:"username"`
}

type Chat struct {
	ID    int64  `json:"id"`
	Type  string `json:"type"`
	Title string `json:"title"`
}

type Sticker struct {
	FileID       string `json:"file_id"`
	FileUniqueID string `json:"file_unique_id"`
	SetName      string `json:"set_name"`
}

type APIResponse struct {
	Ok          bool     `json:"ok"`
	Result      []Update `json:"result,omitempty"`
	Description string   `json:"description,omitempty"`
}

func (m *Message) Command() string {
	if len(m.Text) == 0 || m.Text[0] != '/' {
		return ""
	}

	cmd := m.Text[1:]
	for i, r := range cmd {
		if r == ' ' {
			cmd = cmd[:i]
			break
		}
	}

	return stripBotMention(cmd)
}

func (m *Message) CommandArguments() string {
	if len(m.Text) > 0 && m.Text[0] == '/' {
		for i, r := range m.Text {
			if r == ' ' {
				return m.Text[i+1:]
			}
		}
	}
	return ""
}

func stripBotMention(cmd string) string {
	for i, r := range cmd {
		if r == '@' {
			return cmd[:i]
		}
	}
	return cmd
}
