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

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository는 새로운 UserRepository를 생성합니다.
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound("user")
		}
		return nil, apperrors.ErrDatabase(err)
	}
	return &user, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).Where("email = ? AND deleted_at IS NULL", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound("user")
		}
		return nil, apperrors.ErrDatabase(err)
	}
	return &user, nil
}

func (r *userRepository) FindByTenantID(ctx context.Context, tenantID uuid.UUID, page *pagination.Params) ([]*domain.User, int64, error) {
	var users []*domain.User
	var total int64

	q := r.db.WithContext(ctx).Model(&domain.User{}).Where("tenant_id = ? AND deleted_at IS NULL", tenantID)

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, apperrors.ErrDatabase(err)
	}

	if err := q.Offset(page.Offset()).Limit(page.PageSize).Find(&users).Error; err != nil {
		return nil, 0, apperrors.ErrDatabase(err)
	}

	return users, total, nil
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return apperrors.ErrDatabase(err)
	}
	return nil
}

func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	if err := r.db.WithContext(ctx).Save(user).Error; err != nil {
		return apperrors.ErrDatabase(err)
	}
	return nil
}

func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	if err := r.db.WithContext(ctx).Model(&domain.User{}).Where("id = ?", id).Update("deleted_at", now).Error; err != nil {
		return apperrors.ErrDatabase(err)
	}
	return nil
}

func (r *userRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID, t time.Time) error {
	if err := r.db.WithContext(ctx).Model(&domain.User{}).Where("id = ?", id).Update("last_login_at", t).Error; err != nil {
		return apperrors.ErrDatabase(err)
	}
	return nil
}
