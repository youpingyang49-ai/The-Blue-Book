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
	typ := c.Query("type")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}
	dtos, err := h.notif.List(c.Request.Context(), userID, typ, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": dtos, "page": page, "page_size": pageSize})
}

func (h *NotificationHandler) UnreadCount(c *gin.Context) {
	userID := c.GetUint("userID")
	stats, err := h.notif.Stats(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, stats)
}

func (h *NotificationHandler) MarkAllRead(c *gin.Context) {
	userID := c.GetUint("userID")
	var req struct {
		Types []string `json:"types"`
	}
	// 允许空 body：不传 types 表示全部标记
	_ = c.ShouldBindJSON(&req)
	if err := h.notif.MarkRead(c.Request.Context(), userID, req.Types); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已标记为已读"})
}
