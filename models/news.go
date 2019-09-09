package models

type News struct {
	Title string `json:"title"`
	Date  string `json:"date,omitempty"`
}

type NewsUpdate struct {
	TitleOld string `json:"title_old"`
	TitleNew string `json:"title_new"`
}
