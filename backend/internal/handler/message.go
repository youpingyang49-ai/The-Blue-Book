package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/youpingyang49-ai/The-Blue-Book/backend/internal/service"
)

type MessageHandler struct {
	message *service.MessageService
}

func NewMessageHandler(message *service.MessageService) *MessageHandler {
	return &MessageHandler{message: message}
}

type sendMessageReq struct {
	Content string `json:"content" binding:"required"`
}

func (h *MessageHandler) Send(c *gin.Context) {
	fromUserID := c.GetUint("userID")
	toUserID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	var req sendMessageReq
	if err := c.ShouldBindJSON(&req); err != nil || req.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "消息不能为空"})
		return
	}
	dto, err := h.message.Send(c.Request.Context(), fromUserID, uint(toUserID), req.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "发送失败"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": dto})
}

func (h *MessageHandler) Conversation(c *gin.Context) {
	userID := c.GetUint("userID")
	otherID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	dtos, err := h.message.Conversation(c.Request.Context(), userID, uint(otherID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": dtos})
}

func (h *MessageHandler) Conversations(c *gin.Context) {
	userID := c.GetUint("userID")
	dtos, err := h.message.Conversations(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": dtos})
}
