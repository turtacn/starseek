package utils

import (
	"strings"
	"unicode"
)

// Capitalize 将字符串的第一个字符转换为大写，其余字符保持不变。
// 如果字符串为空，则返回空字符串。
func Capitalize(s string) string {
	if len(s) == 0 {
		return ""
	}
	r := []rune(s)
	return string(unicode.ToUpper(r[0])) + string(r[1:])
}

// SnakeToCamel 将蛇形命名 (snake_case) 字符串转换为大驼峰命名 (PascalCase)。
// 例如: "user_name" -> "UserName", "api_key_id" -> "ApiKeyID"
func SnakeToCamel(s string) string {
	if len(s) == 0 {
		return ""
	}

	parts := strings.Split(s, "_")
	var result strings.Builder
	result.Grow(len(s)) // 预估容量，避免多次内存重新分配

	for _, part := range parts {
		if len(part) > 0 {
			result.WriteString(Capitalize(part))
		}
	}
	return result.String()
}

// CamelToSnake 将驼峰命名 (camelCase 或 PascalCase) 字符串转换为蛇形命名 (snake_case)。
// 例如: "UserName" -> "user_name", "apiKeyID" -> "api_key_id"
// 注意: 此函数对于连续的大写字母（如缩写：APIKey）的处理可能不完全符合所有约定
// (例如 "APIKey" -> "a_p_i_key" 而不是 "api_key")。
// 如果需要更复杂的缩写处理，可能需要一个更复杂的映射规则或正则表达式。
func CamelToSnake(s string) string {
	if len(s) == 0 {
		return ""
	}

	var result strings.Builder
	result.Grow(len(s) + 5) // 预估容量，考虑可能增加的下划线

	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 { // 如果不是第一个字符且是大写字母，则添加下划线
				// 检查前一个字符是否为小写字母，或当前字母后面有小写字母
				// 避免 ID -> i_d, URL -> u_r_l 这样的过度转换，而是 id, url
				if (i > 0 && unicode.IsLower(rune(s[i-1]))) || (i+1 < len(s) && unicode.IsLower(rune(s[i+1]))) {
					result.WriteByte('_')
				}
			}
			result.WriteRune(unicode.ToLower(r))
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}
