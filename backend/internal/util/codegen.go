package util

import (
	"strings"
)

const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// EncodeBase62 将数字 ID 编码为 Base62 短码
func EncodeBase62(id int64) string {
	if id == 0 {
		return string(base62Chars[0])
	}
	var result strings.Builder
	for id > 0 {
		result.WriteByte(base62Chars[id%62])
		id /= 62
	}
	// 反转字符串
	runes := []rune(result.String())
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// DecodeBase62 将 Base62 短码解码为数字 ID（可选，用于验证）
func DecodeBase62(code string) int64 {
	var result int64
	for _, char := range code {
		idx := strings.IndexRune(base62Chars, char)
		if idx == -1 {
			return -1 // 非法字符
		}
		result = result*62 + int64(idx)
	}
	return result
}
