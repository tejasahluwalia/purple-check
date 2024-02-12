package models

import "database/sql"

type Profile struct {
	ID       	string
	Platform 	string
	PlatformUserID 	sql.NullString
	Username 	string
	Status   	string
	Token		sql.NullString
	CreatedAt	string `json:"created_at"`
	UpdatedAt	string `json:"updated_at"`
}

type Feedback struct {
	ID      		string
	GiverID   		string
	ReceiverID  	string
	Comment 		string
	GiverRole   	string
	ReceiverRole  	string
	CreatedAt		string `json:"created_at"`
}
