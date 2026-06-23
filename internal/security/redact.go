package security

import (
	"net/url"
	"regexp"
	"strings"
)

const redacted = "[REDACTED]"

// 敏感信息脱敏使用的匹配规则。
var (
	sensitiveKeyValuePattern = regexp.MustCompile(`(?i)\b([A-Za-z0-9_.-]*(?:api[_-]?key|token|password|passwd|secret|authorization)[A-Za-z0-9_.-]*)\b\s*([:=])\s*([^&\s,;]+)`)
	bearerTokenPattern       = regexp.MustCompile(`(?i)\bBearer\s+[A-Za-z0-9._~+/=-]+`)
	mysqlDSNPasswordPattern  = regexp.MustCompile(`([A-Za-z0-9_.%+-]+):([^@\s]+)@tcp\(`)
)

// RedactString 在写日志前移除字符串中的常见敏感值。
func RedactString(value string) string {
	value = bearerTokenPattern.ReplaceAllString(value, `Bearer `+redacted)
	value = sensitiveKeyValuePattern.ReplaceAllString(value, `${1}${2}`+redacted)
	value = mysqlDSNPasswordPattern.ReplaceAllString(value, `${1}:`+redacted+`@tcp(`)

	return value
}

// RedactURL 脱敏请求 URI 中的敏感查询参数，同时保留非敏感参数便于排查。
func RedactURL(requestURI string) string {
	parsed, err := url.ParseRequestURI(requestURI)
	if err != nil {
		return RedactString(requestURI)
	}

	query := parsed.Query()
	for key := range query {
		if isSensitiveKey(key) {
			query.Set(key, redacted)
		}
	}
	parsed.RawQuery = query.Encode()

	return RedactString(parsed.RequestURI())
}

// isSensitiveKey 判断参数名是否属于敏感字段。
func isSensitiveKey(key string) bool {
	normalized := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(key, "-", "_"), ".", "_"))
	return strings.Contains(normalized, "api_key") ||
		strings.Contains(normalized, "token") ||
		strings.Contains(normalized, "password") ||
		strings.Contains(normalized, "passwd") ||
		strings.Contains(normalized, "secret") ||
		strings.Contains(normalized, "authorization")
}
