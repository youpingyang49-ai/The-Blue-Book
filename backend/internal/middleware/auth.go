package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/youpingyang49-ai/The-Blue-Book/backend/internal/pkg/jwt"
)

// SessionGuard 检查 access token 所属的 refresh 家族是否已被吊销。
// 检测到 refresh token 重放时家族会被立即吊销，这里让该家族的 access token 同步失效。
type SessionGuard interface {
	IsSessionRevoked(ctx context.Context, familyID string) bool
}

func Auth(jwtMgr *jwt.Manager, guard SessionGuard) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token 格式错误"})
			return
		}
		claims, err := jwtMgr.ParseAccess(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token 无效或已过期"})
			return
		}
		if guard != nil && claims.FamilyID != "" && guard.IsSessionRevoked(c.Request.Context(), claims.FamilyID) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "登录状态已失效，请重新登录"})
			return
		}
		c.Set("userID", claims.UserID)
		c.Next()
	}
}

func OptionalAuth(jwtMgr *jwt.Manager, guard SessionGuard) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.Next()
			return
		}
		claims, err := jwtMgr.ParseAccess(parts[1])
		if err == nil && (guard == nil || claims.FamilyID == "" || !guard.IsSessionRevoked(c.Request.Context(), claims.FamilyID)) {
			c.Set("userID", claims.UserID)
		}
		c.Next()
	}
}
