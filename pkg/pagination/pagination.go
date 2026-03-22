package pagination

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	DefaultPage     = 1
	DefaultPageSize = 20
	MaxPageSize     = 100
)

// Params는 페이지네이션 파라미터입니다.
type Params struct {
	Page     int
	PageSize int
}

// Offset은 DB 쿼리용 오프셋을 계산합니다.
func (p *Params) Offset() int {
	return (p.Page - 1) * p.PageSize
}

// FromContext는 gin Context에서 페이지네이션 파라미터를 추출합니다.
func FromContext(c *gin.Context) *Params {
	page := DefaultPage
	pageSize := DefaultPageSize

	if p := c.Query("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}

	if ps := c.Query("page_size"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 {
			if v > MaxPageSize {
				v = MaxPageSize
			}
			pageSize = v
		}
	}

	return &Params{
		Page:     page,
		PageSize: pageSize,
	}
}
