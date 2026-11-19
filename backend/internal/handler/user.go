package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/youpingyang49-ai/The-Blue-Book/backend/internal/repository"
	"github.com/youpingyang49-ai/The-Blue-Book/backend/internal/service"
)

type UserHandler struct {
	userRepo *repository.UserRepo
	userSvc  *service.UserService
}

func NewUserHandler(userRepo *repository.UserRepo, userSvc *service.UserService) *UserHandler {
	return &UserHandler{userRepo: userRepo, userSvc: userSvc}
}

func (h *UserHandler) Me(c *gin.Context) {
	userID := c.GetUint("userID")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}
	user, err := h.userRepo.FindByID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user})
}

func (h *UserHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	viewerID := c.GetUint("userID")
	profile, err := h.userSvc.Profile(c.Request.Context(), uint(id), viewerID)
	if err != nil {
		if err == service.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": profile})
}
