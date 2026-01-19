package util

import (
	"crypto/rand"
	"math/big"
	"strings"
)

const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
const lettersChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz" // 只包含字母，用于确保非纯数字

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

// GenerateCodeFromID 基于自增 ID 生成符合规则的短码：
// - 长度在 6 ~ 32 之间
// - 不能是纯数字（至少包含一个字母）
func GenerateCodeFromID(id int64) string {
	code := EncodeBase62(id)

	// 保证至少 6 位
	for len(code) < 6 {
		code = "a" + code
	}
	// 限制最大 32 位（截取右侧更具区分度）
	if len(code) > 32 {
		code = code[len(code)-32:]
	}

	// 如果全是数字，前面加一个字母前缀
	allDigits := true
	for _, ch := range code {
		if ch < '0' || ch > '9' {
			allDigits = false
			break
		}
	}
	if allDigits {
		code = "a" + code
		if len(code) > 32 {
			code = code[len(code)-32:]
		}
	}

	return code
}

// GenerateRandomCode 生成随机短码（6~32位，不能是纯数字）
// 默认长度为 8 位，确保至少包含一个字母
func GenerateRandomCode(length int) (string, error) {
	if length < 6 {
		length = 6
	}
	if length > 32 {
		length = 32
	}

	// 确保至少包含一个字母（不能是纯数字）
	code := make([]byte, length)
	hasLetter := false

	// 随机生成字符
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(base62Chars))))
		if err != nil {
			return "", err
		}
		code[i] = base62Chars[n.Int64()]
		if (code[i] >= 'A' && code[i] <= 'Z') || (code[i] >= 'a' && code[i] <= 'z') {
			hasLetter = true
		}
	}

	// 如果全是数字，随机替换一个位置为字母
	if !hasLetter {
		// 随机选择一个位置替换为字母
		pos, err := rand.Int(rand.Reader, big.NewInt(int64(length)))
		if err != nil {
			return "", err
		}
		letterPos, err := rand.Int(rand.Reader, big.NewInt(int64(len(lettersChars))))
		if err != nil {
			return "", err
		}
		code[pos.Int64()] = lettersChars[letterPos.Int64()]
	}

	return string(code), nil
}
