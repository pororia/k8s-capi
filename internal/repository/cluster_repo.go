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

type clusterRepository struct {
	db *gorm.DB
}

// NewClusterRepository는 새로운 ClusterRepository를 생성합니다.
func NewClusterRepository(db *gorm.DB) ClusterRepository {
	return &clusterRepository{db: db}
}

func (r *clusterRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Cluster, error) {
	var cluster domain.Cluster
	err := r.db.WithContext(ctx).
		Preload("Nodes").
		Where("id = ? AND deleted_at IS NULL", id).
		First(&cluster).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound("cluster")
		}
		return nil, apperrors.ErrDatabase(err)
	}
	return &cluster, nil
}

func (r *clusterRepository) FindByTenantID(ctx context.Context, tenantID uuid.UUID, filter ClusterFilter, page *pagination.Params) ([]*domain.Cluster, int64, error) {
	filter.TenantID = &tenantID
	return r.FindAll(ctx, filter, page)
}

func (r *clusterRepository) FindAll(ctx context.Context, filter ClusterFilter, page *pagination.Params) ([]*domain.Cluster, int64, error) {
	var clusters []*domain.Cluster
	var total int64

	q := r.db.WithContext(ctx).Model(&domain.Cluster{}).Where("deleted_at IS NULL")

	if filter.TenantID != nil {
		q = q.Where("tenant_id = ?", *filter.TenantID)
	}
	if len(filter.Status) > 0 {
		q = q.Where("status IN ?", filter.Status)
	}
	if filter.Search != "" {
		q = q.Where("name ILIKE ?", "%"+filter.Search+"%")
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, apperrors.ErrDatabase(err)
	}

	// 정렬
	sortBy := "created_at"
	sortOrder := "DESC"
	if filter.SortBy != "" {
		sortBy = filter.SortBy
	}
	if filter.SortOrder == "asc" {
		sortOrder = "ASC"
	}

	if err := q.Order(sortBy + " " + sortOrder).
		Offset(page.Offset()).
		Limit(page.PageSize).
		Find(&clusters).Error; err != nil {
		return nil, 0, apperrors.ErrDatabase(err)
	}

	return clusters, total, nil
}

func (r *clusterRepository) Create(ctx context.Context, cluster *domain.Cluster) error {
	if err := r.db.WithContext(ctx).Create(cluster).Error; err != nil {
		return apperrors.ErrDatabase(err)
	}
	return nil
}

func (r *clusterRepository) Update(ctx context.Context, cluster *domain.Cluster) error {
	if err := r.db.WithContext(ctx).Save(cluster).Error; err != nil {
		return apperrors.ErrDatabase(err)
	}
	return nil
}

func (r *clusterRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.ClusterStatus, errMsg string) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}
	if errMsg != "" {
		updates["error_message"] = errMsg
	}
	if err := r.db.WithContext(ctx).Model(&domain.Cluster{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return apperrors.ErrDatabase(err)
	}
	return nil
}

func (r *clusterRepository) UpdateAPIEndpoint(ctx context.Context, id uuid.UUID, endpoint string) error {
	if err := r.db.WithContext(ctx).Model(&domain.Cluster{}).Where("id = ?", id).Update("api_endpoint", endpoint).Error; err != nil {
		return apperrors.ErrDatabase(err)
	}
	return nil
}

func (r *clusterRepository) Delete(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	if err := r.db.WithContext(ctx).Model(&domain.Cluster{}).Where("id = ?", id).Updates(map[string]interface{}{
		"deleted_at": now,
		"status":     domain.ClusterStatusDeleted,
	}).Error; err != nil {
		return apperrors.ErrDatabase(err)
	}
	return nil
}
