package model

import "time"

// ShortLink 表示一条短链接记录
type ShortLink struct {
	ID             int64      `db:"id"`
	Code           string     `db:"code"`
	LongURL        string     `db:"long_url"`
	CreatedAt      time.Time  `db:"created_at"`
	ExpireAt       *time.Time `db:"expire_at"`
	ClickCount     int64      `db:"click_count"`
	LastAccessedAt *time.Time `db:"last_accessed_at"`
}

