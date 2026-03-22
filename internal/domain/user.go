package domain

import (
	"time"

	"github.com/google/uuid"
)

// UserRole은 사용자 역할을 나타냅니다.
type UserRole string

const (
	RoleSuperAdmin  UserRole = "super_admin"
	RoleTenantAdmin UserRole = "tenant_admin"
	RoleOperator    UserRole = "operator"
	RoleViewer      UserRole = "viewer"
)

// UserStatus는 사용자 상태를 나타냅니다.
type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusInactive  UserStatus = "inactive"
	UserStatusSuspended UserStatus = "suspended"
)

// User는 사용자 도메인 모델입니다.
type User struct {
	ID           uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	TenantID     *uuid.UUID `json:"tenant_id,omitempty" gorm:"type:uuid;index"`
	Email        string     `json:"email" gorm:"size:255;uniqueIndex;not null"`
	PasswordHash string     `json:"-" gorm:"size:255;not null"`
	Name         string     `json:"name" gorm:"size:100;not null"`
	Role         UserRole   `json:"role" gorm:"size:20;not null;default:viewer"`
	Status       UserStatus `json:"status" gorm:"size:20;not null;default:active"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty" gorm:"index"`

	// 연관 관계
	Tenant *Tenant `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`
}

// HasRole은 사용자가 특정 역할 이상의 권한을 가지는지 확인합니다.
// 역할 우선순위: super_admin > tenant_admin > operator > viewer
func (u *User) HasRole(minRole UserRole) bool {
	roleLevel := map[UserRole]int{
		RoleViewer:      1,
		RoleOperator:    2,
		RoleTenantAdmin: 3,
		RoleSuperAdmin:  4,
	}
	return roleLevel[u.Role] >= roleLevel[minRole]
}

// IsSuperAdmin은 Super Admin 역할인지 확인합니다.
func (u *User) IsSuperAdmin() bool {
	return u.Role == RoleSuperAdmin
}

// BelongsToTenant는 사용자가 특정 테넌트에 속하는지 확인합니다.
func (u *User) BelongsToTenant(tenantID uuid.UUID) bool {
	if u.TenantID == nil {
		return false
	}
	return *u.TenantID == tenantID
}

// IsActive는 사용자가 활성 상태인지 확인합니다.
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive && u.DeletedAt == nil
}
