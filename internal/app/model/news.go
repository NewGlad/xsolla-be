package model

import (
	validation "github.com/go-ozzo/ozzo-validation"
)

// News ...
type News struct {
	ID       int    `json:"id"`
	AuthorID int    `json:"author_id"`
	Content  string `json:"content"`
	Likes    string `json:"likes,omitempty"`
}

// Validate ...
func (news *News) Validate() error {
	return validation.ValidateStruct(news,
		validation.Field(&news.Content, validation.Required))
}
