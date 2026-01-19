package redis

import (
	"context"
	"strconv"
	"time"

	redisv9 "github.com/redis/go-redis/v9"

	"url-shortener/backend/internal/model"
	"url-shortener/backend/internal/storage"
)

// RedisRepository 使用 Redis 存储短链接数据
//
// Key 设计：
// - 全局自增：shortener:next_id (string)
// - 记录：shortener:link:{code} (hash)
//   fields: id, code, long_url, created_at, expire_at, click_count, last_accessed_at
type RedisRepository struct {
	rdb *redisv9.Client
}

func NewRepository(rdb *redisv9.Client) storage.LinkRepository {
	return &RedisRepository{rdb: rdb}
}

func (r *RedisRepository) NextID(ctx context.Context) (int64, error) {
	return r.rdb.Incr(ctx, "shortener:next_id").Result()
}

func (r *RedisRepository) Create(ctx context.Context, link *model.ShortLink) (*model.ShortLink, error) {
	key := "shortener:link:" + link.Code

	// 保证 created_at 非空
	if link.CreatedAt.IsZero() {
		link.CreatedAt = time.Now().UTC()
	}

	createdAt := link.CreatedAt.UTC().Format(time.RFC3339Nano)
	expireAt := ""
	if link.ExpireAt != nil {
		expireAt = link.ExpireAt.UTC().Format(time.RFC3339Nano)
	}
	lastAccessed := ""
	if link.LastAccessedAt != nil {
		lastAccessed = link.LastAccessedAt.UTC().Format(time.RFC3339Nano)
	}

	// 如果 key 已存在则返回冲突（上层按 duplicate 处理）
	exists, err := r.rdb.Exists(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	if exists > 0 {
		return nil, redisv9.TxFailedErr
	}

	pipe := r.rdb.TxPipeline()
	pipe.HSet(ctx, key, map[string]any{
		"id":               link.ID,
		"code":             link.Code,
		"long_url":         link.LongURL,
		"created_at":       createdAt,
		"expire_at":        expireAt,
		"click_count":      link.ClickCount,
		"last_accessed_at": lastAccessed,
	})

	// 设置 TTL：如果有 expire_at，则 key TTL = expire_at - now（过期后自动失效）
	if link.ExpireAt != nil {
		ttl := time.Until(link.ExpireAt.UTC())
		if ttl > 0 {
			pipe.Expire(ctx, key, ttl)
		}
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		return nil, err
	}
	return link, nil
}

func (r *RedisRepository) GetByCode(ctx context.Context, code string) (*model.ShortLink, error) {
	key := "shortener:link:" + code
	m, err := r.rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	if len(m) == 0 {
		return nil, nil
	}

	id, _ := strconv.ParseInt(m["id"], 10, 64)
	clickCount, _ := strconv.ParseInt(m["click_count"], 10, 64)

	var createdAt time.Time
	if m["created_at"] != "" {
		createdAt, _ = time.Parse(time.RFC3339Nano, m["created_at"])
	}

	var expireAt *time.Time
	if m["expire_at"] != "" {
		t, err := time.Parse(time.RFC3339Nano, m["expire_at"])
		if err == nil {
			expireAt = &t
		}
	}

	var lastAccessedAt *time.Time
	if m["last_accessed_at"] != "" {
		t, err := time.Parse(time.RFC3339Nano, m["last_accessed_at"])
		if err == nil {
			lastAccessedAt = &t
		}
	}

	return &model.ShortLink{
		ID:             id,
		Code:           m["code"],
		LongURL:        m["long_url"],
		CreatedAt:      createdAt,
		ExpireAt:       expireAt,
		ClickCount:     clickCount,
		LastAccessedAt: lastAccessedAt,
	}, nil
}

func (r *RedisRepository) IncrementClick(ctx context.Context, code string) error {
	key := "shortener:link:" + code
	now := time.Now().UTC().Format(time.RFC3339Nano)
	pipe := r.rdb.TxPipeline()
	pipe.HIncrBy(ctx, key, "click_count", 1)
	pipe.HSet(ctx, key, "last_accessed_at", now)
	_, err := pipe.Exec(ctx)
	return err
}

