package utils

import (
	"fmt"
	"time"
)

// 定义项目统一使用的时间格式。
// RFC3339 是一个常用的 ISO 8601 变体，包含时区信息，适合机器解析和跨系统交换。
// 示例: "2006-01-02T15:04:05Z07:00"
// 更简洁的 UTC 格式，不带毫秒和时区偏移，例如用于数据库字段的纯 UTC 时间戳：
// const UTCLayout = "2006-01-02 15:04:05" // "YYYY-MM-DD HH:MM:SS"
const (
	// DefaultTimeLayout 定义了项目默认的时间格式布局
	// 使用 time.RFC3339 以确保时区信息被正确处理
	DefaultTimeLayout = time.RFC3339
)

// FormatTime 将 time.Time 对象格式化为项目统一的字符串表示。
// 默认会将时间转换为 UTC 时间，然后按 DefaultTimeLayout 格式化。
func FormatTime(t time.Time) string {
	// 如果传入的是零值时间，返回空字符串或特定的零值表示，取决于业务需求。
	// 这里选择返回格式化后的零值，即 "0001-01-01T00:00:00Z"。
	// 如果希望是空字符串，可以加 if t.IsZero() { return "" }
	return t.UTC().Format(DefaultTimeLayout)
}

// ParseTime 将字符串解析为 time.Time 对象，使用项目统一的时间格式。
// 它期望输入字符串符合 DefaultTimeLayout 格式。
func ParseTime(s string) (time.Time, error) {
	parsedTime, err := time.Parse(DefaultTimeLayout, s)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse time string '%s' with layout '%s': %w", s, DefaultTimeLayout, err)
	}
	return parsedTime, nil
}

// UnixMilliToTime 将 Unix 毫秒时间戳转换为 time.Time 对象。
func UnixMilliToTime(unixMilli int64) time.Time {
	return time.Unix(0, unixMilli*int64(time.Millisecond))
}

// TimeToUnixMilli 将 time.Time 对象转换为 Unix 毫秒时间戳。
// 转换为UTC再获取UnixNano，确保一致性
func TimeToUnixMilli(t time.Time) int64 {
	return t.UTC().UnixNano() / int64(time.Millisecond)
}

// DurationToString 将 time.Duration 格式化为易读的字符串，如 "1h30m5s"。
func DurationToString(d time.Duration) string {
	return d.String()
}

// StringToDuration 将字符串解析为 time.Duration，如 "1h30m5s"。
func StringToDuration(s string) (time.Duration, error) {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("failed to parse duration string '%s': %w", s, err)
	}
	return d, nil
}
