package models

import "time"

type SlideMeta struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Title     string    `json:"title"`
	Caption   string    `json:"caption"`
	MimeType  string    `json:"mime_type"`
	ImageURL  string    `json:"image_url"`
}
