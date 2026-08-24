package httpx

import "fmt"

type AppError struct {
	Status  int
	Code    string
	Message string
	Cause   error
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

func (e *AppError) Unwrap() error { return e.Cause }

func New(status int, code, msg string) *AppError {
	return &AppError{Status: status, Code: code, Message: msg}
}

func Wrap(status int, code, msg string, err error) *AppError {
	return &AppError{Status: status, Code: code, Message: msg, Cause: err}
}

var (
	ErrBadRequest      = New(400, "bad_request", "请求参数不完整或类型错误")
	ErrUnauthorized    = New(401, "unauthorized", "未登录或凭证无效")
	ErrForbidden       = New(403, "forbidden", "无权访问该资源")
	ErrNotFound        = New(404, "not_found", "资源不存在")
	ErrConflict        = New(409, "conflict", "资源冲突")
	ErrGone            = New(410, "gone", "凭证已过期")
	ErrUnprocessable   = New(422, "unprocessable", "业务校验未通过")
	ErrTooMany         = New(429, "too_many_requests", "请求过于频繁")
	ErrInternal        = New(500, "internal", "服务内部错误")
	ErrTicketTampered  = New(401, "ticket_tampered", "下载凭证签名无效")
	ErrTicketExpired   = New(410, "ticket_expired", "下载凭证已过期")
	ErrTicketRevoked   = New(403, "ticket_revoked", "下载凭证已吊销")
	ErrTicketExhausted = New(429, "ticket_exhausted", "下载次数已用尽")
	ErrSegmentDenied   = New(403, "segment_denied", "切片超出授权范围")
	ErrRefererDenied   = New(403, "referer_denied", "来源不在防盗链白名单")
	ErrCommentWindow   = New(422, "comment_window", "乐评时间点超出可访问区间")
)
