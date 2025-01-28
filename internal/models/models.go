package models

type Feedback struct {
	ID           string
	Giver        string
	Receiver     string
	Rating       string
	GiverRole    string `json:"giver_role"`
	ReceiverRole string `json:"receiver_role"`
	Comment      string
	CreatedAt    string `json:"created_at"`
}
