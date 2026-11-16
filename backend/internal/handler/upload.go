package handler

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/youpingyang49-ai/The-Blue-Book/backend/internal/pkg/storage"
)

type UploadHandler struct {
	storage *storage.Storage
}

func NewUploadHandler(st *storage.Storage) *UploadHandler {
	return &UploadHandler{storage: st}
}

func (h *UploadHandler) Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少文件"})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "打开文件失败"})
		return
	}
	defer src.Close()

	ext := strings.ToLower(filepath.Ext(file.Filename))
	key := storage.NewKey(ext)
	contentType := file.Header.Get("Content-Type")

	if err := h.storage.Upload(c.Request.Context(), key, src, file.Size, contentType); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "上传失败"})
		return
	}

	url, _ := h.storage.PresignedGetURL(c.Request.Context(), key)
	c.JSON(http.StatusOK, gin.H{"key": key, "url": url})
}
