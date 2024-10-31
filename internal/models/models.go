package models

type Feedback struct {
	ID        string
	Giver     string
	Receiver  string
	Rating    int
	Comment   string
	CreatedAt string `json:"created_at"`
}
