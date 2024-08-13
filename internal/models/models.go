package models

import "database/sql"

type Profile struct {
	ID       	string
	Platform 	string
	PlatformUserID 	sql.NullString
	Username 	string
	Status   	string
	Token		sql.NullString
	ExpiresIn	sql.NullInt64 `json:"expires_in"`
	CreatedAt	string `json:"created_at"`
	UpdatedAt	string `json:"updated_at"`
}

type Feedback struct {
	ID      		string
	GiverID   		string
	ReceiverID  	string
	Giver   		Profile
	Receiver 		Profile
	Rating  		int
	Comment 		string
	CreatedAt		string `json:"created_at"`
}
