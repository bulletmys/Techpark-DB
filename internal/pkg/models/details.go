package models

type Details struct {
	Post   Post    `json:"post"`
	Author *User   `json:"author,omitempty"`
	Forum  *Forum  `json:"forum,omitempty"`
	Thread *Thread `json:"thread,omitempty"`
}
