package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/youpingyang49-ai/The-Blue-Book/backend/internal/service"
)

type AuthHandler struct {
	auth *service.AuthService
	env  string
}

func NewAuthHandler(auth *service.AuthService, env string) *AuthHandler {
	return &AuthHandler{auth: auth, env: env}
}

type sendCodeReq struct {
	Phone string `json:"phone" binding:"required"`
}

func (h *AuthHandler) SendCode(c *gin.Context) {
	var req sendCodeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "手机号不能为空"})
		return
	}
	if !validPhone(req.Phone) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "手机号格式不正确"})
		return
	}

	code, err := h.auth.SendCode(c.Request.Context(), req.Phone)
	if err != nil {
		if err == service.ErrRateLimited {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "发送失败"})
		return
	}

	resp := gin.H{"message": "验证码已发送"}
	if h.env == "development" {
		resp["debug_code"] = code
	}
	c.JSON(http.StatusOK, resp)
}

type loginReq struct {
	Phone string `json:"phone" binding:"required"`
	Code  string `json:"code" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "手机号和验证码不能为空"})
		return
	}

	access, refresh, user, err := h.auth.Login(c.Request.Context(), req.Phone, req.Code)
	if err != nil {
		switch err {
		case service.ErrCodeExpired, service.ErrCodeInvalid:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "登录失败"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  access,
		"refresh_token": refresh,
		"token_type":    "Bearer",
		"user":          user,
	})
}

type refreshReq struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// Refresh 使用 refresh token 静默续期。
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req refreshReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 refresh_token"})
		return
	}
	access, refresh, user, err := h.auth.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		if err == service.ErrSessionExpired {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "登录已过期，请重新登录"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "刷新失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"access_token":  access,
		"refresh_token": refresh,
		"token_type":    "Bearer",
		"user":          user,
	})
}

func validPhone(phone string) bool {
	if len(phone) != 11 {
		return false
	}
	if phone[0] != '1' {
		return false
	}
	if phone[1] < '3' || phone[1] > '9' {
		return false
	}
	for i := 2; i < 11; i++ {
		if phone[i] < '0' || phone[i] > '9' {
			return false
		}
	}
	return true
}
