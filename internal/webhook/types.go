package webhook

type MessageEvent struct {
	Sender struct {
		Id string `json:"id"`
	} `json:"sender"`
	Recipient struct {
		Id string `json:"id"`
	} `json:"recipient"`
	Timestamp int `json:"timestamp"`
	Message   struct {
		Mid            string `json:"mid"`
		Text           string `json:"text"`
		Is_echo        bool   `json:"is_echo"`
		Is_deleted     bool   `json:"is_deleted"`
		Is_unsupported bool   `json:"is_unsupported"`
		Attachments    []struct {
			Type    string `json:"type"`
			Payload struct {
				url string
			} `json:"payload"`
		} `json:"attachments"`
		Quick_reply struct {
			Payload string `json:"payload"`
		} `json:"quick_reply"`
		Referral *Refferal `json:"referral,omitempty"`
		Reply_to struct {
			Mid   string `json:"mid"`
			Story struct {
				Id  string `json:"id"`
				Url string `json:"url"`
			} `json:"story"`
		} `json:"reply_to"`
	} `json:"message"`
	Reaction struct {
		Mid      string `json:"mid"`
		Action   string `json:"action"`
		Reaction string `json:"reaction"`
		Emoji    string `json:"emoji"`
	} `json:"reaction"`
	Postback *struct {
		Mid      string    `json:"mid"`
		Title    string    `json:"title"`
		Payload  string    `json:"payload"`
		Refferal *Refferal `json:"refferal,omitempty"`
	} `json:"postback"`
	Read struct {
		Mid string `json:"mid"`
	} `json:"read"`
	Refferal *Refferal `json:"refferal,omitempty"`
}

type Refferal struct {
	Ref    string `json:"ref"`
	Source string `json:"source"`
	Type   string `json:"type"`
}

type InstagramWebhook struct {
	Object string `json:"object"`
	Entry  []struct {
		Id        string         `json:"id"`
		Time      int            `json:"time"`
		Messaging []MessageEvent `json:"messaging"`
	} `json:"entry"`
}

type ErrorResponse struct {
	Error struct {
		Message      string `json:"message"`
		Type         string `json:"type"`
		Code         string `json:"code"`
		ErrorSubcode string `json:"error_subcode"`
		FBTraceID    string `json:"fbtrace_id"`
	} `json:"error"`
}
