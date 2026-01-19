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
	repo    storage.LinkRepository
	baseURL string
	codeGen func(int64) string
}

func NewLinkService(repo storage.LinkRepository, baseURL string) *LinkService {
	return &LinkService{
		repo:    repo,
		baseURL: baseURL,
		codeGen: util.GenerateCodeFromID,
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
		// 生成随机短码，确保唯一性
		code, err = s.generateUniqueRandomCode(ctx)
		if err != nil {
			return nil, &ServiceError{Type: "internal_error", Message: "failed to generate code"}
		}
	}

	// 检查过期时间是否有效
	if req.ExpireAt != nil && req.ExpireAt.Before(time.Now()) {
		return nil, &ServiceError{Type: "invalid_request", Message: "expire_at must be in the future"}
	}

	link := &model.ShortLink{
		ID:        0,
		Code:      code,
		LongURL:   req.URL,
		CreatedAt: time.Now().UTC(),
		ExpireAt:  req.ExpireAt,
	}

	// 随机生成的短码不需要设置 ID（Redis 存储中 ID 字段可选）

	created, err := s.repo.Create(ctx, link)
	if err != nil {
		// Redis 场景下：只要写入失败且短码已存在，统一返回冲突（更友好）
		existing, _ := s.repo.GetByCode(ctx, code)
		if existing != nil {
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
		_ = s.repo.IncrementClick(context.Background(), code)
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

// generateUniqueRandomCode 生成唯一的随机短码（重试机制）
func (s *LinkService) generateUniqueRandomCode(ctx context.Context) (string, error) {
	maxRetries := 10
	defaultLength := 8 // 默认 8 位随机码

	for i := 0; i < maxRetries; i++ {
		code, err := util.GenerateRandomCode(defaultLength)
		if err != nil {
			return "", err
		}

		// 检查是否已存在
		existing, err := s.repo.GetByCode(ctx, code)
		if err != nil {
			return "", err
		}
		if existing == nil {
			return code, nil
		}
		// 冲突则重试
	}

	return "", errors.New("failed to generate unique code after retries")
}
