package handler

import "time"

type LinkView struct {
	ID        int64
	Slug      string
	TargetURL string
	Title     string
	ExpiresAt *time.Time
	CreatedAt time.Time
	IsActive  bool
	IsExpired bool
	Clicks    int64
}

type AdminPageData struct {
	BaseURL    string
	Links      []LinkView
	LinkCount  int64
	ClickCount int64
	Error      string
}
