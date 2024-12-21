package ievents

type LarkCardEvent struct {
	// 用户在飞书的 OpenID
	OpenID string `json:"open_id"`
	// 卡片触发 token, 可以用于修改 卡片内容
	Token string `json:"token"`
	// 消息ID
	OpenMessageID string `json:"open_message_id"`
	// 群组ID
	OpenChatId string `json:"open_chat_id"`
	// 回调上下文
	Action *struct {
		Value      map[string]interface{} `json:"value"`
		Tag        string                 `json:"tag"`
		Option     string                 `json:"option"`
		Timezone   string                 `json:"timezone"`
		Name       string                 `json:"name"`
		FormValue  map[string]interface{} `json:"form_value"`
		InputValue string                 `json:"input_value"`
		Options    []string               `json:"options"`
		Checked    bool                   `json:"checked"`
	} `json:"action"`
}

func (x *LarkCardEvent) Topic() string {
	return "lark_card_event"
}
