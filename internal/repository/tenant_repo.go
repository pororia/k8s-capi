package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/k8s-cmp/k8s-cmp/internal/domain"
	apperrors "github.com/k8s-cmp/k8s-cmp/pkg/errors"
	"github.com/k8s-cmp/k8s-cmp/pkg/pagination"
	"gorm.io/gorm"
)

type tenantRepository struct {
	db *gorm.DB
}

// NewTenantRepository는 새로운 TenantRepository를 생성합니다.
func NewTenantRepository(db *gorm.DB) TenantRepository {
	return &tenantRepository{db: db}
}

func (r *tenantRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
	var tenant domain.Tenant
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&tenant).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound("tenant")
		}
		return nil, apperrors.ErrDatabase(err)
	}
	return &tenant, nil
}

func (r *tenantRepository) FindByName(ctx context.Context, name string) (*domain.Tenant, error) {
	var tenant domain.Tenant
	err := r.db.WithContext(ctx).Where("name = ? AND deleted_at IS NULL", name).First(&tenant).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound("tenant")
		}
		return nil, apperrors.ErrDatabase(err)
	}
	return &tenant, nil
}

func (r *tenantRepository) FindAll(ctx context.Context, page *pagination.Params) ([]*domain.Tenant, int64, error) {
	var tenants []*domain.Tenant
	var total int64

	q := r.db.WithContext(ctx).Model(&domain.Tenant{}).Where("deleted_at IS NULL")

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, apperrors.ErrDatabase(err)
	}

	if err := q.Order("created_at DESC").Offset(page.Offset()).Limit(page.PageSize).Find(&tenants).Error; err != nil {
		return nil, 0, apperrors.ErrDatabase(err)
	}

	return tenants, total, nil
}

func (r *tenantRepository) Create(ctx context.Context, tenant *domain.Tenant) error {
	if err := r.db.WithContext(ctx).Create(tenant).Error; err != nil {
		return apperrors.ErrDatabase(err)
	}
	return nil
}

func (r *tenantRepository) Update(ctx context.Context, tenant *domain.Tenant) error {
	if err := r.db.WithContext(ctx).Save(tenant).Error; err != nil {
		return apperrors.ErrDatabase(err)
	}
	return nil
}

func (r *tenantRepository) Delete(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	if err := r.db.WithContext(ctx).Model(&domain.Tenant{}).Where("id = ?", id).Update("deleted_at", now).Error; err != nil {
		return apperrors.ErrDatabase(err)
	}
	return nil
}

func (r *tenantRepository) CountClusters(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.Cluster{}).
		Where("tenant_id = ? AND deleted_at IS NULL AND status != ?", tenantID, domain.ClusterStatusDeleted).
		Count(&count).Error
	if err != nil {
		return 0, apperrors.ErrDatabase(err)
	}
	return count, nil
}
