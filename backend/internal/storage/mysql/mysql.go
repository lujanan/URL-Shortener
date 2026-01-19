package mysql

import (
	"context"
	"database/sql"
	"errors"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"url-shortener/backend/internal/model"
	"url-shortener/backend/internal/storage"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) storage.LinkRepository {
	return &Repository{db: db}
}

// InitSchema 创建 short_links 表（如不存在）
func InitSchema(db *sql.DB) error {
	schema := `
CREATE TABLE IF NOT EXISTS short_links (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  code VARCHAR(32) NOT NULL UNIQUE,
  long_url TEXT NOT NULL,
  created_at DATETIME NOT NULL,
  expire_at DATETIME NULL,
  click_count BIGINT NOT NULL DEFAULT 0,
  last_accessed_at DATETIME NULL,
  INDEX idx_expire_at (expire_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
`
	_, err := db.Exec(schema)
	return err
}

func (r *Repository) Create(ctx context.Context, link *model.ShortLink) (*model.ShortLink, error) {
	now := time.Now().UTC()
	if link.CreatedAt.IsZero() {
		link.CreatedAt = now
	}
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO short_links (code, long_url, created_at, expire_at, click_count, last_accessed_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		link.Code,
		link.LongURL,
		link.CreatedAt,
		link.ExpireAt,
		link.ClickCount,
		link.LastAccessedAt,
	)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	link.ID = id
	return link, nil
}

func (r *Repository) GetByCode(ctx context.Context, code string) (*model.ShortLink, error) {
	row := r.db.QueryRowContext(ctx, `
SELECT id, code, long_url, created_at, expire_at, click_count, last_accessed_at
FROM short_links
WHERE code = ? LIMIT 1
`, code)

	var link model.ShortLink
	var expireAt sql.NullTime
	var lastAccessed sql.NullTime

	err := row.Scan(
		&link.ID,
		&link.Code,
		&link.LongURL,
		&link.CreatedAt,
		&expireAt,
		&link.ClickCount,
		&lastAccessed,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if expireAt.Valid {
		link.ExpireAt = &expireAt.Time
	}
	if lastAccessed.Valid {
		link.LastAccessedAt = &lastAccessed.Time
	}

	return &link, nil
}

func (r *Repository) IncrementClick(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `
UPDATE short_links
SET click_count = click_count + 1,
    last_accessed_at = ?
WHERE id = ?
`, time.Now().UTC(), id)
	return err
}

