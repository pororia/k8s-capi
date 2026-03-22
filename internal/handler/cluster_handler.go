package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/k8s-cmp/k8s-cmp/internal/middleware"
	"github.com/k8s-cmp/k8s-cmp/internal/service"
	apperrors "github.com/k8s-cmp/k8s-cmp/pkg/errors"
	"github.com/k8s-cmp/k8s-cmp/pkg/pagination"
	"github.com/k8s-cmp/k8s-cmp/pkg/response"
)

// ClusterHandler는 클러스터 관련 HTTP 핸들러입니다.
type ClusterHandler struct {
	clusterSvc *service.ClusterService
}

// NewClusterHandler는 새로운 ClusterHandler를 생성합니다.
func NewClusterHandler(clusterSvc *service.ClusterService) *ClusterHandler {
	return &ClusterHandler{clusterSvc: clusterSvc}
}

// List godoc
// GET /api/v1/clusters
func (h *ClusterHandler) List(c *gin.Context) {
	tenantID, err := getTenantID(c)
	if err != nil {
		response.Error(c, err)
		return
	}

	page := pagination.FromContext(c)
	filter := service.ClusterListFilter{
		Search:    c.Query("search"),
		SortBy:    c.Query("sort_by"),
		SortOrder: c.Query("sort_order"),
	}
	if statuses := c.QueryArray("status"); len(statuses) > 0 {
		filter.Status = statuses
	}

	clusters, total, err := h.clusterSvc.List(c.Request.Context(), *tenantID, filter, page)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OKPaginated(c, clusters, page.Page, page.PageSize, total)
}

// Get godoc
// GET /api/v1/clusters/:id
func (h *ClusterHandler) Get(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		response.Error(c, err)
		return
	}

	cluster, err := h.clusterSvc.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, cluster)
}

// Create godoc
// POST /api/v1/clusters
func (h *ClusterHandler) Create(c *gin.Context) {
	tenantID, err := getTenantID(c)
	if err != nil {
		response.Error(c, err)
		return
	}

	userID, err := getUserID(c)
	if err != nil {
		response.Error(c, err)
		return
	}

	var req service.CreateClusterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.ErrValidation(err.Error()))
		return
	}

	cluster, err := h.clusterSvc.Create(c.Request.Context(), *tenantID, *userID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Created(c, cluster)
}

// Delete godoc
// DELETE /api/v1/clusters/:id
func (h *ClusterHandler) Delete(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		response.Error(c, err)
		return
	}

	if err := h.clusterSvc.Delete(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}

	response.NoContent(c)
}

// Scale godoc
// POST /api/v1/clusters/:id/scale
func (h *ClusterHandler) Scale(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		response.Error(c, err)
		return
	}

	var req service.ScaleParams
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.ErrValidation(err.Error()))
		return
	}

	if err := h.clusterSvc.Scale(c.Request.Context(), id, req); err != nil {
		response.Error(c, err)
		return
	}

	response.NoContent(c)
}

// Upgrade godoc
// POST /api/v1/clusters/:id/upgrade
func (h *ClusterHandler) Upgrade(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		response.Error(c, err)
		return
	}

	var req service.UpgradeParams
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.ErrValidation(err.Error()))
		return
	}

	if err := h.clusterSvc.Upgrade(c.Request.Context(), id, req); err != nil {
		response.Error(c, err)
		return
	}

	response.NoContent(c)
}

// GetKubeconfig godoc
// GET /api/v1/clusters/:id/kubeconfig
func (h *ClusterHandler) GetKubeconfig(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		response.Error(c, err)
		return
	}

	kubeconfig, err := h.clusterSvc.GetKubeconfig(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}

	c.Header("Content-Disposition", "attachment; filename=kubeconfig.yaml")
	c.Data(200, "application/yaml", kubeconfig)
}

// GetEvents godoc
// GET /api/v1/clusters/:id/events
func (h *ClusterHandler) GetEvents(c *gin.Context) {
	id, err := parseUUID(c, "id")
	if err != nil {
		response.Error(c, err)
		return
	}

	page := pagination.FromContext(c)
	events, total, err := h.clusterSvc.GetEvents(c.Request.Context(), id, page)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OKPaginated(c, events, page.Page, page.PageSize, total)
}

// 헬퍼 함수들
func parseUUID(c *gin.Context, param string) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Param(param))
	if err != nil {
		return uuid.Nil, apperrors.ErrBadRequest("invalid " + param + " format")
	}
	return id, nil
}

func getTenantID(c *gin.Context) (*uuid.UUID, error) {
	claims, exists := c.Get(middleware.ContextKeyClaims)
	if !exists {
		return nil, apperrors.ErrUnauthorized("claims not found")
	}

	svcClaims, ok := claims.(*service.Claims)
	if !ok {
		return nil, apperrors.ErrUnauthorized("invalid claims type")
	}

	// Super Admin은 헤더에서 tenant_id를 받거나 path param 사용
	if svcClaims.TenantID == nil {
		headerTenantID := c.GetHeader("X-Tenant-ID")
		if headerTenantID != "" {
			id, err := uuid.Parse(headerTenantID)
			if err != nil {
				return nil, apperrors.ErrBadRequest("invalid X-Tenant-ID header")
			}
			return &id, nil
		}
		return nil, apperrors.ErrBadRequest("tenant_id required")
	}

	return svcClaims.TenantID, nil
}

func getUserID(c *gin.Context) (*uuid.UUID, error) {
	userIDStr, exists := c.Get(middleware.ContextKeyUserID)
	if !exists {
		return nil, apperrors.ErrUnauthorized("user_id not found")
	}
	id, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		return nil, apperrors.ErrUnauthorized("invalid user_id")
	}
	return &id, nil
}
