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

// Following 关注列表
func (h *UserHandler) Following(c *gin.Context) {
	h.listRelations(c, true)
}

// Followers 粉丝列表
func (h *UserHandler) Followers(c *gin.Context) {
	h.listRelations(c, false)
}

func (h *UserHandler) listRelations(c *gin.Context, following bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	page := 1
	if v, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil && v > 0 {
		page = v
	}
	pageSize := 20
	if v, err := strconv.Atoi(c.DefaultQuery("page_size", "20")); err == nil && v > 0 && v <= 50 {
		pageSize = v
	}
	viewerID := c.GetUint("userID")

	var items []service.FollowUserDTO
	var total int64
	if following {
		items, total, err = h.userSvc.Following(c.Request.Context(), uint(id), viewerID, page, pageSize)
	} else {
		items, total, err = h.userSvc.Followers(c.Request.Context(), uint(id), viewerID, page, pageSize)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": total})
}
