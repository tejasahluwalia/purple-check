package models

type Feedback struct {
	ID        string
	Giver     string
	Receiver  string
	Rating    string
	Comment   string
	CreatedAt string `json:"created_at"`
}
