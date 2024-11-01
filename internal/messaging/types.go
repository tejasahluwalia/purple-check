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
	TemplateType string            `json:"template_type"`
	Elements     []PayloadElements `json:"elements"`
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

type PersistentMenuRequestBody struct {
	Platform       string           `json:"platform"`
	PersistentMenu []PersistentMenu `json:"persistent_menu"`
}

type PersistentMenu struct {
	Locale                string         `json:"locale"`
	ComposerInputDisabled bool           `json:"composer_input_disabled"`
	CallToActions         []CallToAction `json:"call_to_actions"`
}

type CallToAction struct {
	Type               string `json:"type"`
	Title              string `json:"title"`
	Payload            string `json:"payload,omitempty"`
	URL                string `json:"url,omitempty"`
	WebviewHeightRatio string `json:"webview_height_ratio,omitempty"`
}
