package model

import "time"

type Link struct {
	ID        int64
	Slug      string
	TargetURL string
	Title     string
	ExpiresAt *time.Time
	CreatedAt time.Time
	IsActive  bool
}

type Click struct {
	ID        int64
	LinkID    int64
	ClickedAt time.Time
	Referrer  string
	UserAgent string
}

type ClickStat struct {
	Date  string
	Count int64
}
