package telegram

type Update struct {
	UpdateID int      `json:"update_id"`
	Message  *Message `json:"message"`
}

type Message struct {
	MessageID int    `json:"message_id"`
	From      *User  `json:"from"`
	Chat      *Chat  `json:"chat"`
	Text      string `json:"text"`
}

func (m *Message) Command() string {
	if len(m.Text) > 0 && m.Text[0] == '/' {
		for i, r := range m.Text {
			if r == ' ' {
				return m.Text[1:i]
			}
		}
		return m.Text[1:]
	}
	return ""
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

type User struct {
	ID        int64  `json:"id"`
	IsBot     bool   `json:"is_bot"`
	FirstName string `json:"first_name"`
	UserName  string `json:"username"`
}

type Chat struct {
	ID   int64  `json:"id"`
	Type string `json:"type"`
}

type APIResponse struct {
	Ok          bool     `json:"ok"`
	Result      []Update `json:"result,omitempty"`
	Description string   `json:"description,omitempty"`
}
