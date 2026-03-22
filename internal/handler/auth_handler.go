package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/k8s-cmp/k8s-cmp/internal/middleware"
	"github.com/k8s-cmp/k8s-cmp/internal/service"
	apperrors "github.com/k8s-cmp/k8s-cmp/pkg/errors"
	"github.com/k8s-cmp/k8s-cmp/pkg/response"
)

// AuthHandler는 인증 관련 HTTP 핸들러입니다.
type AuthHandler struct {
	authSvc *service.AuthService
}

// NewAuthHandler는 새로운 AuthHandler를 생성합니다.
func NewAuthHandler(authSvc *service.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// Login godoc
// POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.ErrValidation(err.Error()))
		return
	}

	tokens, user, err := h.authSvc.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, gin.H{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"user":          user,
	})
}

// Refresh godoc
// POST /api/v1/auth/refresh
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.ErrValidation(err.Error()))
		return
	}

	tokens, err := h.authSvc.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, tokens)
}

// Logout godoc
// POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	userIDStr, _ := c.Get(middleware.ContextKeyUserID)
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		response.Error(c, apperrors.ErrBadRequest("invalid user id"))
		return
	}

	if err := h.authSvc.Logout(c.Request.Context(), userID); err != nil {
		response.Error(c, err)
		return
	}

	response.NoContent(c)
}

// Me godoc
// GET /api/v1/auth/me
func (h *AuthHandler) Me(c *gin.Context) {
	claims, _ := c.Get(middleware.ContextKeyClaims)
	response.OK(c, claims)
}
