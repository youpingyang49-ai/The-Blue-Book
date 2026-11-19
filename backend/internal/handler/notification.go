package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/youpingyang49-ai/The-Blue-Book/backend/internal/service"
)

type NotificationHandler struct {
	notif *service.NotificationService
}

func NewNotificationHandler(notif *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{notif: notif}
}

func (h *NotificationHandler) List(c *gin.Context) {
	userID := c.GetUint("userID")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}
	dtos, err := h.notif.List(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": dtos, "page": page, "page_size": pageSize})
}

func (h *NotificationHandler) UnreadCount(c *gin.Context) {
	userID := c.GetUint("userID")
	count, err := h.notif.UnreadCount(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"count": count})
}

func (h *NotificationHandler) MarkAllRead(c *gin.Context) {
	userID := c.GetUint("userID")
	if err := h.notif.MarkAllRead(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已全部标记为已读"})
}
