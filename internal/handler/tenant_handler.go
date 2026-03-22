package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/k8s-cmp/k8s-cmp/internal/service"
	apperrors "github.com/k8s-cmp/k8s-cmp/pkg/errors"
	"github.com/k8s-cmp/k8s-cmp/pkg/pagination"
	"github.com/k8s-cmp/k8s-cmp/pkg/response"
)

// TenantHandler는 테넌트 관련 HTTP 핸들러입니다.
type TenantHandler struct {
	tenantSvc *service.TenantService
}

// NewTenantHandler는 새로운 TenantHandler를 생성합니다.
func NewTenantHandler(tenantSvc *service.TenantService) *TenantHandler {
	return &TenantHandler{tenantSvc: tenantSvc}
}

// List godoc
// GET /api/v1/tenants
func (h *TenantHandler) List(c *gin.Context) {
	page := pagination.FromContext(c)
	tenants, total, err := h.tenantSvc.List(c.Request.Context(), page)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OKPaginated(c, tenants, page.Page, page.PageSize, total)
}

// Get godoc
// GET /api/v1/tenants/:id
func (h *TenantHandler) Get(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		response.Error(c, err)
		return
	}

	tenant, err := h.tenantSvc.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, tenant)
}

// Create godoc
// POST /api/v1/tenants
func (h *TenantHandler) Create(c *gin.Context) {
	var req service.CreateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.ErrValidation(err.Error()))
		return
	}

	tenant, err := h.tenantSvc.Create(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Created(c, tenant)
}

// Update godoc
// PUT /api/v1/tenants/:id
func (h *TenantHandler) Update(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		response.Error(c, err)
		return
	}

	var req service.UpdateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.ErrValidation(err.Error()))
		return
	}

	tenant, err := h.tenantSvc.Update(c.Request.Context(), id, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, tenant)
}

// Delete godoc
// DELETE /api/v1/tenants/:id
func (h *TenantHandler) Delete(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		response.Error(c, err)
		return
	}

	if err := h.tenantSvc.Delete(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}

	response.NoContent(c)
}
