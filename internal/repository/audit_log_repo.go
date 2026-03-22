package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/k8s-cmp/k8s-cmp/internal/domain"
	apperrors "github.com/k8s-cmp/k8s-cmp/pkg/errors"
	"github.com/k8s-cmp/k8s-cmp/pkg/pagination"
	"gorm.io/gorm"
)

type auditLogRepository struct {
	db *gorm.DB
}

// NewAuditLogRepository는 새로운 AuditLogRepository를 생성합니다.
func NewAuditLogRepository(db *gorm.DB) AuditLogRepository {
	return &auditLogRepository{db: db}
}

func (r *auditLogRepository) Create(ctx context.Context, log *domain.AuditLog) error {
	if err := r.db.WithContext(ctx).Create(log).Error; err != nil {
		return apperrors.ErrDatabase(err)
	}
	return nil
}

func (r *auditLogRepository) FindAll(ctx context.Context, filter AuditLogFilter, page *pagination.Params) ([]*domain.AuditLog, int64, error) {
	var logs []*domain.AuditLog
	var total int64

	q := r.db.WithContext(ctx).Model(&domain.AuditLog{})

	if filter.TenantID != nil {
		q = q.Where("tenant_id = ?", *filter.TenantID)
	}
	if filter.UserID != nil {
		q = q.Where("user_id = ?", *filter.UserID)
	}
	if filter.Action != "" {
		q = q.Where("action = ?", filter.Action)
	}
	if filter.ResourceType != "" {
		q = q.Where("resource_type = ?", filter.ResourceType)
	}
	if filter.Result != "" {
		q = q.Where("result = ?", filter.Result)
	}
	if filter.Search != "" {
		q = q.Where("resource_name ILIKE ?", "%"+filter.Search+"%")
	}
	if filter.From != nil {
		q = q.Where("created_at >= ?", *filter.From)
	}
	if filter.To != nil {
		q = q.Where("created_at <= ?", *filter.To)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, apperrors.ErrDatabase(err)
	}

	if err := q.Order("created_at DESC").Offset(page.Offset()).Limit(page.PageSize).Find(&logs).Error; err != nil {
		return nil, 0, apperrors.ErrDatabase(err)
	}

	return logs, total, nil
}

func (r *auditLogRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.AuditLog, error) {
	var log domain.AuditLog
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&log).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound("audit log")
		}
		return nil, apperrors.ErrDatabase(err)
	}
	return &log, nil
}
