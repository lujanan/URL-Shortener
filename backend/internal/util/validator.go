package util

import (
	"net/url"
	"strings"
)

// ValidateURL 验证 URL 是否合法（必须是 http 或 https）
func ValidateURL(rawURL string) error {
	if rawURL == "" {
		return ErrEmptyURL
	}
	if len(rawURL) > 2048 {
		return ErrURLTooLong
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ErrInvalidURL
	}
	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return ErrInvalidScheme
	}
	if parsed.Host == "" {
		return ErrInvalidURL
	}
	return nil
}

// ValidateCode 验证自定义短码是否合法（仅允许字母数字，长度限制）
func ValidateCode(code string) error {
	if code == "" {
		return ErrEmptyCode
	}
	if len(code) < 6 {
		return ErrCodeTooShort
	}
	if len(code) > 32 {
		return ErrCodeTooLong
	}
	allDigits := true
	for _, char := range code {
		if !((char >= '0' && char <= '9') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= 'a' && char <= 'z')) {
			return ErrInvalidCode
		}
		if char < '0' || char > '9' {
			allDigits = false
		}
	}
	if allDigits {
		return ErrCodeAllDigits
	}
	return nil
}
