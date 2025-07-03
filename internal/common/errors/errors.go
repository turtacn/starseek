package errors

import (
	"fmt"
)

// ErrorCode 定义错误码类型
type ErrorCode int

// 定义项目使用的错误码常量
// 错误码分类：
// 0: 成功
// 1000-1999: 通用客户端请求错误 (HTTP 4xx 级别)
// 2000-2999: 数据库/基础设施错误
// 3000-3999: 业务逻辑错误 (特定于应用)
// 4000-4999: 内部系统错误 (HTTP 5xx 级别)
const (
	ErrCodeSuccess         ErrorCode = 0    // 操作成功
	ErrCodeUnknown         ErrorCode = 1000 // 未知错误，通常不应该出现
	ErrCodeInvalidParam    ErrorCode = 1001 // 参数校验失败或非法参数
	ErrCodeNotFound        ErrorCode = 1002 // 资源未找到
	ErrCodeUnauthorized    ErrorCode = 1003 // 未授权访问
	ErrCodeForbidden       ErrorCode = 1004 // 禁止访问（权限不足）
	ErrCodeTooManyRequests ErrorCode = 1005 // 请求频率过高

	ErrCodeDatabaseError      ErrorCode = 2001 // 数据库操作失败
	ErrCodeDatabaseConnection ErrorCode = 2002 // 数据库连接失败
	ErrCodeConfigError        ErrorCode = 2003 // 配置错误

	ErrCodeIndexNotRegistered ErrorCode = 3001 // 索引未注册或不存在
	ErrCodeTokenizerNotFound  ErrorCode = 3002 // 指定分词器未找到
	ErrCodeRankingFailed      ErrorCode = 3003 // 召回/排序失败
	ErrCodeOperationConflict  ErrorCode = 3004 // 操作冲突（如重复创建）

	ErrCodeInternalError ErrorCode = 4001 // 内部服务器错误，通常是未捕获的异常或逻辑错误
)

// AppError 结构体定义了项目内部的统一错误类型
type AppError struct {
	Code    ErrorCode `json:"code"`    // 错误码
	Message string    `json:"message"` // 用户友好的错误信息
	Detail  string    `json:"detail"`  // 详细的内部错误信息或原始错误字符串，用于调试和日志
}

// Error 实现 error 接口
func (e *AppError) Error() string {
	// 仅返回用户友好的 Message，不暴露内部 Detail
	return e.Message
}

// ToAppError 将一个 error 转换为 AppError。
// 如果传入的 error 已经是 *AppError 类型，则直接返回；否则，封装为 InternalError。
func ToAppError(err error) *AppError {
	if err == nil {
		return nil
	}
	if appErr, ok := err.(*AppError); ok {
		return appErr
	}
	// 对于普通 error，统一封装为内部错误，并记录原始错误信息
	return NewInternalError("Internal server error", err)
}

// NewAppError 创建一个新的 AppError 实例
// msg: 用户友好的错误信息
// detailErr: 原始错误或更详细的内部描述
func NewAppError(code ErrorCode, msg string, detailErr error) *AppError {
	detail := ""
	if detailErr != nil {
		detail = detailErr.Error()
	}
	return &AppError{
		Code:    code,
		Message: msg,
		Detail:  detail,
	}
}

// NewInvalidParamError 创建一个参数非法错误
func NewInvalidParamError(msg string, detailErr error) *AppError {
	return NewAppError(ErrCodeInvalidParam, msg, detailErr)
}

// NewNotFoundError 创建一个资源未找到错误
func NewNotFoundError(msg string, detailErr error) *AppError {
	return NewAppError(ErrCodeNotFound, msg, detailErr)
}

// NewDatabaseError 创建一个数据库操作错误
func NewDatabaseError(msg string, detailErr error) *AppError {
	return NewAppError(ErrCodeDatabaseError, msg, detailErr)
}

// NewInternalError 创建一个内部服务器错误
func NewInternalError(msg string, detailErr error) *AppError {
	// 如果 msg 为空，提供一个默认的内部错误消息
	if msg == "" {
		msg = "An unexpected error occurred"
	}
	return NewAppError(ErrCodeInternalError, msg, detailErr)
}

// NewIndexNotRegisteredError 创建一个索引未注册错误
func NewIndexNotRegisteredError(indexName string, detailErr error) *AppError {
	msg := fmt.Sprintf("Index '%s' is not registered or does not exist", indexName)
	return NewAppError(ErrCodeIndexNotRegistered, msg, detailErr)
}

// NewTokenizerNotFoundError 创建一个分词器未找到错误
func NewTokenizerNotFoundError(tokenizerName string, detailErr error) *AppError {
	msg := fmt.Sprintf("Tokenizer '%s' not found", tokenizerName)
	return NewAppError(ErrCodeTokenizerNotFound, msg, detailErr)
}

// NewRankingFailedError 创建一个排序失败错误
func NewRankingFailedError(msg string, detailErr error) *AppError {
	return NewAppError(ErrCodeRankingFailed, msg, detailErr)
}

// IsAppErrorType 判断一个 error 是否是特定错误码的 AppError
func IsAppErrorType(err error, code ErrorCode) bool {
	appErr, ok := err.(*AppError)
	return ok && appErr.Code == code
}

// IsInternalError 判断一个 error 是否是内部错误
func IsInternalError(err error) bool {
	return IsAppErrorType(err, ErrCodeInternalError)
}

// IsInvalidParamError 判断一个 error 是否是参数非法错误
func IsInvalidParamError(err error) bool {
	return IsAppErrorType(err, ErrCodeInvalidParam)
}
