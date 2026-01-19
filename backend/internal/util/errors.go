package util

import "errors"

var (
	ErrEmptyURL      = errors.New("url is required")
	ErrURLTooLong    = errors.New("url exceeds maximum length of 2048 characters")
	ErrInvalidURL    = errors.New("invalid url format, must be http or https")
	ErrInvalidScheme = errors.New("url scheme must be http or https")
	ErrEmptyCode     = errors.New("code is required")
	ErrCodeTooLong   = errors.New("code exceeds maximum length of 32 characters")
	ErrInvalidCode   = errors.New("code contains invalid characters, only alphanumeric allowed")
)
