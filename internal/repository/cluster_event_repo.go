package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/k8s-cmp/k8s-cmp/internal/domain"
	apperrors "github.com/k8s-cmp/k8s-cmp/pkg/errors"
	"github.com/k8s-cmp/k8s-cmp/pkg/pagination"
	"gorm.io/gorm"
)

type clusterEventRepository struct {
	db *gorm.DB
}

// NewClusterEventRepository는 새로운 ClusterEventRepository를 생성합니다.
func NewClusterEventRepository(db *gorm.DB) ClusterEventRepository {
	return &clusterEventRepository{db: db}
}

func (r *clusterEventRepository) FindByClusterID(ctx context.Context, clusterID uuid.UUID, page *pagination.Params) ([]*domain.ClusterEvent, int64, error) {
	var events []*domain.ClusterEvent
	var total int64

	q := r.db.WithContext(ctx).Model(&domain.ClusterEvent{}).Where("cluster_id = ?", clusterID)

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, apperrors.ErrDatabase(err)
	}

	if err := q.Order("created_at DESC").Offset(page.Offset()).Limit(page.PageSize).Find(&events).Error; err != nil {
		return nil, 0, apperrors.ErrDatabase(err)
	}

	return events, total, nil
}

func (r *clusterEventRepository) Create(ctx context.Context, event *domain.ClusterEvent) error {
	if err := r.db.WithContext(ctx).Create(event).Error; err != nil {
		return apperrors.ErrDatabase(err)
	}
	return nil
}
