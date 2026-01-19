package service

import (
	"context"
	"errors"
	"time"

	"url-shortener/backend/internal/model"
	"url-shortener/backend/internal/storage"
	"url-shortener/backend/internal/util"
)

type LinkService struct {
	repo     storage.LinkRepository
	baseURL  string
	codeGen  func(int64) string
}

func NewLinkService(repo storage.LinkRepository, baseURL string) *LinkService {
	return &LinkService{
		repo:    repo,
		baseURL: baseURL,
		codeGen: util.EncodeBase62,
	}
}

type CreateRequest struct {
	URL        string     `json:"url" binding:"required"`
	CustomCode string     `json:"custom_code,omitempty"`
	ExpireAt   *time.Time `json:"expire_at,omitempty"`
}

type CreateResponse struct {
	Code     string     `json:"code"`
	ShortURL string     `json:"short_url"`
	LongURL  string     `json:"long_url"`
	ExpireAt *time.Time `json:"expire_at,omitempty"`
}

type LinkInfoResponse struct {
	Code           string     `json:"code"`
	LongURL        string     `json:"long_url"`
	CreatedAt      time.Time  `json:"created_at"`
	ExpireAt       *time.Time `json:"expire_at,omitempty"`
	ClickCount     int64      `json:"click_count"`
	LastAccessedAt *time.Time `json:"last_accessed_at,omitempty"`
}

// CreateShortLink 创建短链接
func (s *LinkService) CreateShortLink(ctx context.Context, req *CreateRequest) (*CreateResponse, error) {
	// 验证 URL
	if err := util.ValidateURL(req.URL); err != nil {
		return nil, &ServiceError{Type: "invalid_request", Message: err.Error()}
	}

	var code string
	var err error

	if req.CustomCode != "" {
		// 使用自定义短码
		if err := util.ValidateCode(req.CustomCode); err != nil {
			return nil, &ServiceError{Type: "invalid_request", Message: err.Error()}
		}
		code = req.CustomCode
		// 检查是否已存在
		existing, err := s.repo.GetByCode(ctx, code)
		if err != nil {
			return nil, &ServiceError{Type: "internal_error", Message: "failed to check code availability"}
		}
		if existing != nil {
			return nil, &ServiceError{Type: "conflict", Message: "custom code already exists"}
		}
	} else {
		// 自动生成短码：先插入一条记录获取 ID，然后用 ID 生成短码
		// 为了简化，我们先生成一个临时短码，如果冲突则重试
		// 实际生产环境可以使用更复杂的策略（如预分配 ID 段）
		code, err = s.generateUniqueCode(ctx)
		if err != nil {
			return nil, &ServiceError{Type: "internal_error", Message: "failed to generate code"}
		}
	}

	// 检查过期时间是否有效
	if req.ExpireAt != nil && req.ExpireAt.Before(time.Now()) {
		return nil, &ServiceError{Type: "invalid_request", Message: "expire_at must be in the future"}
	}

	link := &model.ShortLink{
		Code:    code,
		LongURL: req.URL,
		ExpireAt: req.ExpireAt,
	}

	created, err := s.repo.Create(ctx, link)
	if err != nil {
		// 检查是否是唯一约束冲突
		if isDuplicateKeyError(err) {
			return nil, &ServiceError{Type: "conflict", Message: "code already exists"}
		}
		return nil, &ServiceError{Type: "internal_error", Message: "failed to create short link"}
	}

	shortURL := s.baseURL + "/" + created.Code
	return &CreateResponse{
		Code:     created.Code,
		ShortURL: shortURL,
		LongURL:  created.LongURL,
		ExpireAt: created.ExpireAt,
	}, nil
}

// generateUniqueCode 生成唯一的短码（简化版：使用时间戳 + 随机数，实际可用 ID 生成）
func (s *LinkService) generateUniqueCode(ctx context.Context) (string, error) {
	// 简化实现：使用时间戳的 Base62 编码 + 随机后缀
	// 实际生产环境建议使用数据库自增 ID + Base62
	timestamp := time.Now().UnixNano()
	code := util.EncodeBase62(timestamp)
	// 确保长度合理（取前 8 位，如果不够则补零）
	if len(code) < 6 {
		code = code + "000000"[:6-len(code)]
	}
	if len(code) > 8 {
		code = code[:8]
	}

	// 检查是否冲突，如果冲突则追加随机字符
	for i := 0; i < 10; i++ {
		existing, err := s.repo.GetByCode(ctx, code)
		if err != nil {
			return "", err
		}
		if existing == nil {
			return code, nil
		}
		// 冲突则追加字符
		code = code + util.EncodeBase62(int64(i))[:1]
	}
	return code, errors.New("failed to generate unique code after retries")
}

// GetLongURL 根据短码获取长链接（用于重定向）
func (s *LinkService) GetLongURL(ctx context.Context, code string) (string, error) {
	link, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		return "", &ServiceError{Type: "internal_error", Message: "failed to query link"}
	}
	if link == nil {
		return "", &ServiceError{Type: "not_found", Message: "short link not found"}
	}

	// 检查是否过期
	if link.ExpireAt != nil && link.ExpireAt.Before(time.Now()) {
		return "", &ServiceError{Type: "not_found", Message: "short link expired"}
	}

	// 异步更新点击次数（可选：可以放到 goroutine 中）
	go func() {
		_ = s.repo.IncrementClick(context.Background(), link.ID)
	}()

	return link.LongURL, nil
}

// GetLinkInfo 获取短链接详细信息
func (s *LinkService) GetLinkInfo(ctx context.Context, code string) (*LinkInfoResponse, error) {
	link, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		return nil, &ServiceError{Type: "internal_error", Message: "failed to query link"}
	}
	if link == nil {
		return nil, &ServiceError{Type: "not_found", Message: "short link not found"}
	}

	return &LinkInfoResponse{
		Code:           link.Code,
		LongURL:        link.LongURL,
		CreatedAt:      link.CreatedAt,
		ExpireAt:       link.ExpireAt,
		ClickCount:     link.ClickCount,
		LastAccessedAt: link.LastAccessedAt,
	}, nil
}

// ServiceError 业务错误
type ServiceError struct {
	Type    string
	Message string
}

func (e *ServiceError) Error() string {
	return e.Message
}

// isDuplicateKeyError 检查是否是唯一约束冲突（MySQL 错误码 1062）
func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return contains(errStr, "Duplicate entry") || contains(errStr, "1062")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || indexOf(s, substr) >= 0)
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
