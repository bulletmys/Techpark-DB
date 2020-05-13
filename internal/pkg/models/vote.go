package models

type Vote struct {
	Nick  string `json:"nickname"`
	Voice int8   `json:"voice"`
}
