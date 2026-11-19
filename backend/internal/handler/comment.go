package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/youpingyang49-ai/The-Blue-Book/backend/internal/service"
)

type CommentHandler struct {
	comment *service.CommentService
}

func NewCommentHandler(comment *service.CommentService) *CommentHandler {
	return &CommentHandler{comment: comment}
}

type createCommentReq struct {
	Content  string `json:"content" binding:"required"`
	ParentID uint   `json:"parent_id"`
}

func (h *CommentHandler) Create(c *gin.Context) {
	userID := c.GetUint("userID")
	postID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	var req createCommentReq
	if err := c.ShouldBindJSON(&req); err != nil || req.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "评论内容不能为空"})
		return
	}
	dto, err := h.comment.Create(c.Request.Context(), userID, uint(postID), req.ParentID, req.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "评论失败"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"comment": dto})
}

func (h *CommentHandler) List(c *gin.Context) {
	postID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	dtos, err := h.comment.List(c.Request.Context(), uint(postID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": dtos})
}
