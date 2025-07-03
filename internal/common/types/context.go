package types

import (
	"context"
)

// ContextKey 是自定义上下文键的类型，旨在通过 Go 的类型系统避免上下文键冲突。
// 使用空结构体类型作为键的类型，可以确保键的唯一性，因为即使两个包使用相同的字符串常量，
// 只要它们的 ContextKey 类型不同，它们的键就不会冲突。
// 此外，将键定义为未导出（小写开头）的类型，可以强制使用者通过本包提供的辅助函数来操作上下文。
type contextKey string

const (
	// ContextKeyRequestID 用于在上下文中存储请求ID
	ContextKeyRequestID contextKey = "requestID"

	// ContextKeyLogger 用于在上下文中存储一个与请求相关的 Logger 实例
	// 具体的 Logger 类型将在 internal/infrastructure/logger 中定义，这里使用 interface{} 占位
	ContextKeyLogger contextKey = "logger"

	// ContextKeyAuthUserID 用于在上下文中存储认证后的用户ID
	ContextKeyAuthUserID contextKey = "authUserID"

	// 可以在这里添加更多通用的上下文键
)

// GetRequestIDFromContext 从上下文中获取请求ID。
// 如果请求ID不存在或类型不匹配，则返回空字符串。
func GetRequestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	val := ctx.Value(ContextKeyRequestID)
	if id, ok := val.(string); ok {
		return id
	}
	return ""
}

// SetRequestIDToContext 将请求ID设置到上下文中。
func SetRequestIDToContext(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, ContextKeyRequestID, requestID)
}

// GetLoggerFromContext 从上下文中获取与请求相关的 Logger 实例。
// 返回 interface{} 类型，具体的类型转换由调用者根据实际的 Logger 实现进行。
func GetLoggerFromContext(ctx context.Context) interface{} {
	if ctx == nil {
		return nil
	}
	return ctx.Value(ContextKeyLogger)
}

// SetLoggerToContext 将 Logger 实例设置到上下文中。
func SetLoggerToContext(ctx context.Context, logger interface{}) context.Context {
	return context.WithValue(ctx, ContextKeyLogger, logger)
}

// GetAuthUserIDFromContext 从上下文中获取认证后的用户ID。
func GetAuthUserIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	val := ctx.Value(ContextKeyAuthUserID)
	if id, ok := val.(string); ok { // 假设用户ID是字符串
		return id
	}
	return ""
}

// SetAuthUserIDToContext 将用户ID设置到上下文中。
func SetAuthUserIDToContext(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, ContextKeyAuthUserID, userID)
}
