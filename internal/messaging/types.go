package messaging

type MessageRecipient struct {
	ID string `json:"id"`
}

type MessageText struct {
	Text string `json:"text"`
}

type ElementButton struct {
	Type    string `json:"type"`
	Title   string `json:"title"`
	Payload string `json:"payload,omitempty"`
	URL     string `json:"url,omitempty"`
}

type PayloadElements struct {
	Title   string          `json:"title"`
	Buttons []ElementButton `json:"buttons"`
}

type AttachmentPayload struct {
	TemplateType string          `json:"template_type"`
	Text         string          `json:"text,omitempty"`
	Buttons      []ElementButton `json:"buttons,omitempty"`
}

type MessageAttachment struct {
	Type    string            `json:"type"`
	Payload AttachmentPayload `json:"payload"`
}

type MessageButtons struct {
	Attachment MessageAttachment `json:"attachment"`
}

type Message interface {
	MessageText | MessageButtons
}

type MessageRequestBody[T Message] struct {
	Recipient MessageRecipient `json:"recipient"`
	Message   T                `json:"message"`
}

type MessengerProfileRequestBody struct {
	Platform       string           `json:"platform"`
	PersistentMenu []PersistentMenu `json:"persistent_menu"`
	IceBreakers    []IceBreakers    `json:"ice_breakers"`
}

type PersistentMenu struct {
	Locale                string                       `json:"locale"`
	ComposerInputDisabled bool                         `json:"composer_input_disabled"`
	CallToActions         []PersistentMenuCallToAction `json:"call_to_actions"`
}

type PersistentMenuCallToAction struct {
	Type               string `json:"type"`
	Title              string `json:"title"`
	Payload            string `json:"payload,omitempty"`
	URL                string `json:"url,omitempty"`
	WebviewHeightRatio string `json:"webview_height_ratio,omitempty"`
}

type IceBreakerCallToAction struct {
	Question string `json:"question"`
	Payload  string `json:"payload"`
}

type IceBreakers struct {
	Locale        string                   `json:"locale"`
	CallToActions []IceBreakerCallToAction `json:"call_to_actions"`
}
