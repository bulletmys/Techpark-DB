package models

type Forum struct {
	Slug    string `json:"slug"`
	User    string `json:"user"`
	Title   string `json:"title"`
	Threads int32  `json:"threads"`
	Posts   int64  `json:"posts"`
}
