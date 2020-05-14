package models

type User struct {
	Nickname string `json:"nickname" sql:"nick"`
	FullName string `json:"fullname" sql:"name"`
	Email    string `json:"email" sql:"email"`
	About    string `json:"about" sql:"about"`
}
