package storage

import (
	"context"

	"url-shortener/backend/internal/model"
)

// LinkRepository 定义短链接存储接口，方便未来替换实现（如 Redis/MySQL 等）
type LinkRepository interface {
	// Create 保存新的短链接记录，并返回带 ID 的记录
	Create(ctx context.Context, link *model.ShortLink) (*model.ShortLink, error)
	// GetByCode 根据短码查询
	GetByCode(ctx context.Context, code string) (*model.ShortLink, error)
	// IncrementClick 在访问时增加点击次数并更新 last_accessed_at
	IncrementClick(ctx context.Context, id int64) error
}

