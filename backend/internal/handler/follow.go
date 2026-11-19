package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/youpingyang49-ai/The-Blue-Book/backend/internal/service"
)

type FollowHandler struct {
	follow *service.FollowService
}

func NewFollowHandler(follow *service.FollowService) *FollowHandler {
	return &FollowHandler{follow: follow}
}

func (h *FollowHandler) Follow(c *gin.Context) {
	followerID := c.GetUint("userID")
	followeeID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	followed, err := h.follow.Follow(c.Request.Context(), followerID, uint(followeeID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"followed": followed})
}

func (h *FollowHandler) Unfollow(c *gin.Context) {
	followerID := c.GetUint("userID")
	followeeID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	unfollowed, err := h.follow.Unfollow(c.Request.Context(), followerID, uint(followeeID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"unfollowed": unfollowed})
}
