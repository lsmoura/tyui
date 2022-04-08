package model

import "time"

type Links struct {
	ID        int64     `json:"-"`
	Token     string    `json:"token"`
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"created_at"`
	Clicks    int64     `json:"clicks"`
}
