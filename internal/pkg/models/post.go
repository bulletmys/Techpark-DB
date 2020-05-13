package models

import "time"

type Post struct {
	Author   string    `json:"author"`
	Created  time.Time `json:"created"`
	Forum    string    `json:"forum"`
	ID       int64     `json:"id"`
	IsEdited bool      `json:"isEdited"`
	Message  string    `json:"message"`
	Parent   int64     `json:"parent"`
	Thread   int32     `json:"thread"`
}
