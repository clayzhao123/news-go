package news

import "time"

type Article struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	URL         string    `json:"url"`
	Source      string    `json:"source"`
	Content     string    `json:"content,omitempty"`
	PublishedAt time.Time `json:"published_at"`
}
