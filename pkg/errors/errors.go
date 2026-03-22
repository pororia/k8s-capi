package errors

import (
	"fmt"
	"net/http"
)

// ErrorCode는 애플리케이션 에러 코드를 나타냅니다.
type ErrorCode string

const (
	// 인증 관련
	ErrCodeUnauthorized   ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden      ErrorCode = "FORBIDDEN"
	ErrCodeTokenExpired   ErrorCode = "TOKEN_EXPIRED"
	ErrCodeInvalidToken   ErrorCode = "INVALID_TOKEN"

	// 리소스 관련
	ErrCodeNotFound      ErrorCode = "NOT_FOUND"
	ErrCodeAlreadyExists ErrorCode = "ALREADY_EXISTS"
	ErrCodeConflict      ErrorCode = "CONFLICT"

	// 입력값 검증
	ErrCodeValidation    ErrorCode = "VALIDATION_ERROR"
	ErrCodeBadRequest    ErrorCode = "BAD_REQUEST"

	// 내부 오류
	ErrCodeInternal      ErrorCode = "INTERNAL_ERROR"
	ErrCodeDatabase      ErrorCode = "DATABASE_ERROR"
	ErrCodeExternal      ErrorCode = "EXTERNAL_SERVICE_ERROR"

	// 쿼터 초과
	ErrCodeQuotaExceeded ErrorCode = "QUOTA_EXCEEDED"

	// 상태 오류
	ErrCodeInvalidStatus ErrorCode = "INVALID_STATUS"
)

// AppError는 애플리케이션 커스텀 에러 타입입니다.
type AppError struct {
	Code       ErrorCode `json:"code"`
	Message    string    `json:"message"`
	HTTPStatus int       `json:"-"`
	Cause      error     `json:"-"`
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Cause
}

// New는 새로운 AppError를 생성합니다.
func New(code ErrorCode, message string, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

// Wrap은 기존 에러를 감싸는 AppError를 생성합니다.
func Wrap(code ErrorCode, message string, httpStatus int, cause error) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
		Cause:      cause,
	}
}

// 자주 사용하는 에러 생성 헬퍼
func ErrNotFound(resource string) *AppError {
	return New(ErrCodeNotFound, fmt.Sprintf("%s not found", resource), http.StatusNotFound)
}

func ErrAlreadyExists(resource string) *AppError {
	return New(ErrCodeAlreadyExists, fmt.Sprintf("%s already exists", resource), http.StatusConflict)
}

func ErrUnauthorized(message string) *AppError {
	return New(ErrCodeUnauthorized, message, http.StatusUnauthorized)
}

func ErrForbidden(message string) *AppError {
	return New(ErrCodeForbidden, message, http.StatusForbidden)
}

func ErrValidation(message string) *AppError {
	return New(ErrCodeValidation, message, http.StatusBadRequest)
}

func ErrBadRequest(message string) *AppError {
	return New(ErrCodeBadRequest, message, http.StatusBadRequest)
}

func ErrInternal(cause error) *AppError {
	return Wrap(ErrCodeInternal, "internal server error", http.StatusInternalServerError, cause)
}

func ErrDatabase(cause error) *AppError {
	return Wrap(ErrCodeDatabase, "database error", http.StatusInternalServerError, cause)
}

func ErrQuotaExceeded(message string) *AppError {
	return New(ErrCodeQuotaExceeded, message, http.StatusForbidden)
}

func ErrInvalidStatus(message string) *AppError {
	return New(ErrCodeInvalidStatus, message, http.StatusConflict)
}

// IsAppError는 주어진 에러가 AppError인지 확인합니다.
func IsAppError(err error) (*AppError, bool) {
	if appErr, ok := err.(*AppError); ok {
		return appErr, true
	}
	return nil, false
}
