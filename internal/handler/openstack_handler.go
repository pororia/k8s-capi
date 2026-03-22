package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/k8s-cmp/k8s-cmp/internal/service"
	"github.com/k8s-cmp/k8s-cmp/pkg/response"
)

// OpenStackHandler는 OpenStack 리소스 조회 핸들러입니다.
type OpenStackHandler struct {
	osSvc *service.OpenStackService
}

// NewOpenStackHandler는 새로운 OpenStackHandler를 생성합니다.
func NewOpenStackHandler(osSvc *service.OpenStackService) *OpenStackHandler {
	return &OpenStackHandler{osSvc: osSvc}
}

// GetFlavors godoc
// GET /api/v1/openstack/flavors
func (h *OpenStackHandler) GetFlavors(c *gin.Context) {
	tenantID, err := getTenantID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	flavors, err := h.osSvc.GetFlavors(c.Request.Context(), *tenantID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, flavors)
}

// GetImages godoc
// GET /api/v1/openstack/images
func (h *OpenStackHandler) GetImages(c *gin.Context) {
	tenantID, err := getTenantID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	images, err := h.osSvc.GetImages(c.Request.Context(), *tenantID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, images)
}

// GetNetworks godoc
// GET /api/v1/openstack/networks
func (h *OpenStackHandler) GetNetworks(c *gin.Context) {
	tenantID, err := getTenantID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	networks, err := h.osSvc.GetNetworks(c.Request.Context(), *tenantID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, networks)
}

// GetSubnets godoc
// GET /api/v1/openstack/networks/:id/subnets
func (h *OpenStackHandler) GetSubnets(c *gin.Context) {
	tenantID, err := getTenantID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	networkID := c.Param("id")
	subnets, err := h.osSvc.GetSubnets(c.Request.Context(), *tenantID, networkID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, subnets)
}

// GetSecurityGroups godoc
// GET /api/v1/openstack/security-groups
func (h *OpenStackHandler) GetSecurityGroups(c *gin.Context) {
	tenantID, err := getTenantID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	sgs, err := h.osSvc.GetSecurityGroups(c.Request.Context(), *tenantID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, sgs)
}

// GetKeypairs godoc
// GET /api/v1/openstack/keypairs
func (h *OpenStackHandler) GetKeypairs(c *gin.Context) {
	tenantID, err := getTenantID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	kps, err := h.osSvc.GetKeypairs(c.Request.Context(), *tenantID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, kps)
}

// GetExternalNetworks godoc
// GET /api/v1/openstack/external-networks
func (h *OpenStackHandler) GetExternalNetworks(c *gin.Context) {
	tenantID, err := getTenantID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	nets, err := h.osSvc.GetExternalNetworks(c.Request.Context(), *tenantID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, nets)
}

// GetQuotas godoc
// GET /api/v1/openstack/quotas
func (h *OpenStackHandler) GetQuotas(c *gin.Context) {
	tenantID, err := getTenantID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	quotas, err := h.osSvc.GetQuotas(c.Request.Context(), *tenantID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, quotas)
}
