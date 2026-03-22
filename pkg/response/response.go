package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	apperrors "github.com/k8s-cmp/k8s-cmp/pkg/errors"
)

// Response는 표준 API 응답 구조입니다.
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorBody  `json:"error,omitempty"`
}

// ErrorBody는 에러 응답 본문입니다.
type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// PaginatedResponse는 페이지네이션이 포함된 응답 구조입니다.
type PaginatedResponse struct {
	Success    bool        `json:"success"`
	Data       interface{} `json:"data"`
	Pagination *Pagination `json:"pagination"`
}

// Pagination은 페이지네이션 메타데이터입니다.
type Pagination struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// OK는 성공 응답을 반환합니다.
func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    data,
	})
}

// Created는 201 Created 응답을 반환합니다.
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{
		Success: true,
		Data:    data,
	})
}

// NoContent는 204 No Content 응답을 반환합니다.
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// OKPaginated는 페이지네이션이 포함된 성공 응답을 반환합니다.
func OKPaginated(c *gin.Context, data interface{}, page, pageSize int, total int64) {
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}
	c.JSON(http.StatusOK, PaginatedResponse{
		Success: true,
		Data:    data,
		Pagination: &Pagination{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}

// Error는 에러 응답을 반환합니다.
func Error(c *gin.Context, err error) {
	if appErr, ok := apperrors.IsAppError(err); ok {
		c.JSON(appErr.HTTPStatus, Response{
			Success: false,
			Error: &ErrorBody{
				Code:    string(appErr.Code),
				Message: appErr.Message,
			},
		})
		return
	}

	// 알 수 없는 에러는 500으로 처리
	c.JSON(http.StatusInternalServerError, Response{
		Success: false,
		Error: &ErrorBody{
			Code:    string(apperrors.ErrCodeInternal),
			Message: "internal server error",
		},
	})
}

// Abort는 에러와 함께 요청을 중단합니다.
func Abort(c *gin.Context, err error) {
	if appErr, ok := apperrors.IsAppError(err); ok {
		c.AbortWithStatusJSON(appErr.HTTPStatus, Response{
			Success: false,
			Error: &ErrorBody{
				Code:    string(appErr.Code),
				Message: appErr.Message,
			},
		})
		return
	}
	c.AbortWithStatusJSON(http.StatusInternalServerError, Response{
		Success: false,
		Error: &ErrorBody{
			Code:    string(apperrors.ErrCodeInternal),
			Message: "internal server error",
		},
	})
}
