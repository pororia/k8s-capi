package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/k8s-cmp/k8s-cmp/internal/domain"
	"github.com/k8s-cmp/k8s-cmp/internal/repository"
	"github.com/k8s-cmp/k8s-cmp/pkg/crypto"
	apperrors "github.com/k8s-cmp/k8s-cmp/pkg/errors"
	"github.com/k8s-cmp/k8s-cmp/pkg/pagination"
	"go.uber.org/zap"
)

// CreateTenantRequest는 테넌트 생성 요청입니다.
type CreateTenantRequest struct {
	Name               string                      `json:"name" validate:"required,min=2,max=100"`
	Description        string                      `json:"description"`
	OSAuthURL          string                      `json:"os_auth_url" validate:"required,url"`
	OSProjectID        string                      `json:"os_project_id" validate:"required"`
	OSProjectName      string                      `json:"os_project_name" validate:"required"`
	OSDomainName       string                      `json:"os_domain_name"`
	Credentials        domain.OSCredentialsJSON    `json:"credentials" validate:"required"`
	MaxClusters        int                         `json:"max_clusters"`
	MaxNodesPerCluster int                         `json:"max_nodes_per_cluster"`
}

// UpdateTenantRequest는 테넌트 수정 요청입니다.
type UpdateTenantRequest struct {
	Description        *string                     `json:"description"`
	OSAuthURL          *string                     `json:"os_auth_url"`
	OSProjectID        *string                     `json:"os_project_id"`
	OSProjectName      *string                     `json:"os_project_name"`
	OSDomainName       *string                     `json:"os_domain_name"`
	Credentials        *domain.OSCredentialsJSON   `json:"credentials"`
	MaxClusters        *int                        `json:"max_clusters"`
	MaxNodesPerCluster *int                        `json:"max_nodes_per_cluster"`
	Status             *domain.TenantStatus        `json:"status"`
}

// TenantService는 테넌트 비즈니스 로직을 담당합니다.
type TenantService struct {
	tenantRepo repository.TenantRepository
	crypto     *crypto.AES
	logger     *zap.Logger
}

// NewTenantService는 새로운 TenantService를 생성합니다.
func NewTenantService(tenantRepo repository.TenantRepository, crypto *crypto.AES, logger *zap.Logger) *TenantService {
	return &TenantService{
		tenantRepo: tenantRepo,
		crypto:     crypto,
		logger:     logger,
	}
}

// List는 테넌트 목록을 반환합니다.
func (s *TenantService) List(ctx context.Context, page *pagination.Params) ([]*domain.Tenant, int64, error) {
	return s.tenantRepo.FindAll(ctx, page)
}

// GetByID는 특정 테넌트를 반환합니다.
func (s *TenantService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
	return s.tenantRepo.FindByID(ctx, id)
}

// Create는 새 테넌트를 생성합니다.
func (s *TenantService) Create(ctx context.Context, req *CreateTenantRequest) (*domain.Tenant, error) {
	// 이름 중복 확인
	if _, err := s.tenantRepo.FindByName(ctx, req.Name); err == nil {
		return nil, apperrors.ErrAlreadyExists("tenant")
	}

	// OpenStack 자격증명 암호화
	credsJSON, err := json.Marshal(req.Credentials)
	if err != nil {
		return nil, apperrors.ErrInternal(fmt.Errorf("failed to marshal credentials: %w", err))
	}
	encrypted, err := s.crypto.Encrypt(string(credsJSON))
	if err != nil {
		return nil, apperrors.ErrInternal(fmt.Errorf("failed to encrypt credentials: %w", err))
	}

	maxClusters := 10
	if req.MaxClusters > 0 {
		maxClusters = req.MaxClusters
	}
	maxNodes := 20
	if req.MaxNodesPerCluster > 0 {
		maxNodes = req.MaxNodesPerCluster
	}
	domainName := "Default"
	if req.OSDomainName != "" {
		domainName = req.OSDomainName
	}

	tenant := &domain.Tenant{
		Name:               req.Name,
		Description:        req.Description,
		OSAuthURL:          req.OSAuthURL,
		OSProjectID:        req.OSProjectID,
		OSProjectName:      req.OSProjectName,
		OSDomainName:       domainName,
		OSCredentials:      encrypted,
		MaxClusters:        maxClusters,
		MaxNodesPerCluster: maxNodes,
		Status:             domain.TenantStatusActive,
	}

	if err := s.tenantRepo.Create(ctx, tenant); err != nil {
		return nil, err
	}

	s.logger.Info("tenant created", zap.String("tenant_id", tenant.ID.String()), zap.String("name", tenant.Name))
	return tenant, nil
}

// Update는 테넌트를 수정합니다.
func (s *TenantService) Update(ctx context.Context, id uuid.UUID, req *UpdateTenantRequest) (*domain.Tenant, error) {
	tenant, err := s.tenantRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Description != nil {
		tenant.Description = *req.Description
	}
	if req.OSAuthURL != nil {
		tenant.OSAuthURL = *req.OSAuthURL
	}
	if req.OSProjectID != nil {
		tenant.OSProjectID = *req.OSProjectID
	}
	if req.OSProjectName != nil {
		tenant.OSProjectName = *req.OSProjectName
	}
	if req.OSDomainName != nil {
		tenant.OSDomainName = *req.OSDomainName
	}
	if req.Credentials != nil {
		credsJSON, err := json.Marshal(req.Credentials)
		if err != nil {
			return nil, apperrors.ErrInternal(err)
		}
		encrypted, err := s.crypto.Encrypt(string(credsJSON))
		if err != nil {
			return nil, apperrors.ErrInternal(err)
		}
		tenant.OSCredentials = encrypted
	}
	if req.MaxClusters != nil {
		tenant.MaxClusters = *req.MaxClusters
	}
	if req.MaxNodesPerCluster != nil {
		tenant.MaxNodesPerCluster = *req.MaxNodesPerCluster
	}
	if req.Status != nil {
		tenant.Status = *req.Status
	}

	if err := s.tenantRepo.Update(ctx, tenant); err != nil {
		return nil, err
	}

	return tenant, nil
}

// Delete는 테넌트를 삭제합니다 (soft delete).
func (s *TenantService) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := s.tenantRepo.FindByID(ctx, id); err != nil {
		return err
	}
	return s.tenantRepo.Delete(ctx, id)
}

// DecryptCredentials는 테넌트의 암호화된 자격증명을 복호화합니다.
func (s *TenantService) DecryptCredentials(tenant *domain.Tenant) (*domain.OSCredentialsJSON, error) {
	plaintext, err := s.crypto.Decrypt(tenant.OSCredentials)
	if err != nil {
		return nil, apperrors.ErrInternal(fmt.Errorf("failed to decrypt credentials: %w", err))
	}

	var creds domain.OSCredentialsJSON
	if err := json.Unmarshal([]byte(plaintext), &creds); err != nil {
		return nil, apperrors.ErrInternal(fmt.Errorf("failed to unmarshal credentials: %w", err))
	}

	return &creds, nil
}
