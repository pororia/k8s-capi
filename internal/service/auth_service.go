package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/k8s-cmp/k8s-cmp/internal/domain"
	"github.com/k8s-cmp/k8s-cmp/internal/repository"
	apperrors "github.com/k8s-cmp/k8s-cmp/pkg/errors"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// TokenPair는 액세스 토큰과 리프레시 토큰 쌍입니다.
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// Claims는 JWT 클레임입니다.
type Claims struct {
	UserID   uuid.UUID        `json:"sub"`
	TenantID *uuid.UUID       `json:"tenant_id,omitempty"`
	Role     domain.UserRole  `json:"role"`
	jwt.RegisteredClaims
}

// AuthService는 인증/인가 서비스입니다.
type AuthService struct {
	userRepo         repository.UserRepository
	redis            *redis.Client
	logger           *zap.Logger
	jwtSecret        []byte
	accessTokenTTL   time.Duration
	refreshTokenTTL  time.Duration
}

// NewAuthService는 새로운 AuthService를 생성합니다.
func NewAuthService(
	userRepo repository.UserRepository,
	redis *redis.Client,
	logger *zap.Logger,
	jwtSecret string,
	accessTokenTTL time.Duration,
	refreshTokenTTL time.Duration,
) *AuthService {
	return &AuthService{
		userRepo:        userRepo,
		redis:           redis,
		logger:          logger,
		jwtSecret:       []byte(jwtSecret),
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
	}
}

// Login은 이메일/비밀번호로 로그인하고 토큰 쌍을 반환합니다.
func (s *AuthService) Login(ctx context.Context, email, password string) (*TokenPair, *domain.User, error) {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, nil, apperrors.ErrUnauthorized("invalid credentials")
	}

	if !user.IsActive() {
		return nil, nil, apperrors.ErrUnauthorized("account is not active")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, nil, apperrors.ErrUnauthorized("invalid credentials")
	}

	tokens, err := s.generateTokenPair(user)
	if err != nil {
		return nil, nil, err
	}

	// 리프레시 토큰을 Redis에 저장
	key := s.refreshTokenKey(user.ID)
	if err := s.redis.Set(ctx, key, tokens.RefreshToken, s.refreshTokenTTL).Err(); err != nil {
		s.logger.Error("failed to store refresh token", zap.Error(err))
	}

	// 마지막 로그인 시간 업데이트
	now := time.Now()
	if err := s.userRepo.UpdateLastLogin(ctx, user.ID, now); err != nil {
		s.logger.Warn("failed to update last login", zap.Error(err))
	}

	return tokens, user, nil
}

// Refresh는 리프레시 토큰으로 새로운 토큰 쌍을 발급합니다.
func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*TokenPair, error) {
	claims, err := s.validateToken(refreshToken)
	if err != nil {
		return nil, apperrors.ErrUnauthorized("invalid refresh token")
	}

	// Redis에서 유효한 리프레시 토큰인지 확인
	key := s.refreshTokenKey(claims.UserID)
	stored, err := s.redis.Get(ctx, key).Result()
	if err != nil || stored != refreshToken {
		return nil, apperrors.ErrUnauthorized("refresh token expired or revoked")
	}

	user, err := s.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, apperrors.ErrUnauthorized("user not found")
	}

	if !user.IsActive() {
		return nil, apperrors.ErrUnauthorized("account is not active")
	}

	tokens, err := s.generateTokenPair(user)
	if err != nil {
		return nil, err
	}

	// 새 리프레시 토큰으로 교체
	if err := s.redis.Set(ctx, key, tokens.RefreshToken, s.refreshTokenTTL).Err(); err != nil {
		s.logger.Error("failed to update refresh token", zap.Error(err))
	}

	return tokens, nil
}

// Logout은 리프레시 토큰을 무효화합니다.
func (s *AuthService) Logout(ctx context.Context, userID uuid.UUID) error {
	key := s.refreshTokenKey(userID)
	if err := s.redis.Del(ctx, key).Err(); err != nil {
		s.logger.Error("failed to delete refresh token", zap.Error(err))
	}
	return nil
}

// ValidateAccessToken은 액세스 토큰을 검증하고 클레임을 반환합니다.
func (s *AuthService) ValidateAccessToken(tokenStr string) (*Claims, error) {
	return s.validateToken(tokenStr)
}

func (s *AuthService) generateTokenPair(user *domain.User) (*TokenPair, error) {
	now := time.Now()

	// 액세스 토큰
	accessClaims := &Claims{
		UserID:   user.ID,
		TenantID: user.TenantID,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString(s.jwtSecret)
	if err != nil {
		return nil, apperrors.ErrInternal(fmt.Errorf("failed to sign access token: %w", err))
	}

	// 리프레시 토큰 (클레임 포함하여 검증 가능)
	refreshClaims := &Claims{
		UserID: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.refreshTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString(s.jwtSecret)
	if err != nil {
		return nil, apperrors.ErrInternal(fmt.Errorf("failed to sign refresh token: %w", err))
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) validateToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

func (s *AuthService) refreshTokenKey(userID uuid.UUID) string {
	return fmt.Sprintf("refresh_token:%s", userID.String())
}

// HashPassword는 비밀번호를 bcrypt로 해싱합니다.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// MarshalClaims는 Claims를 JSON으로 직렬화합니다 (테스트/디버깅용).
func MarshalClaims(claims *Claims) string {
	b, _ := json.Marshal(claims)
	return string(b)
}
