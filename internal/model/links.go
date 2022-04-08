package model

import "time"

type Links struct {
	Id        int64
	Token     string
	Url       string
	CreatedAt time.Time
	Clicks    int64
}

func (l Links) CreateTable() string {
	return `CREATE TABLE IF NOT EXISTS links (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		token TEXT UNIQUE,
		url TEXT,
		created_at DATETIME,
		clicks INTEGER
	)`
}
