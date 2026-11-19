package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/youpingyang49-ai/The-Blue-Book/backend/internal/service"
)

type PostHandler struct {
	post *service.PostService
}

func NewPostHandler(post *service.PostService) *PostHandler {
	return &PostHandler{post: post}
}

type createPostReq struct {
	Title     string   `json:"title"`
	Content   string   `json:"content"`
	ImageKeys []string `json:"image_keys"`
}

func (h *PostHandler) Create(c *gin.Context) {
	userID := c.GetUint("userID")
	var req createPostReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	if len(req.ImageKeys) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "至少需要一张图片"})
		return
	}
	post, err := h.post.Create(c.Request.Context(), userID, req.Title, req.Content, req.ImageKeys)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "发布失败"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"post_id": post.ID})
}

func (h *PostHandler) List(c *gin.Context) {
	sort := c.DefaultQuery("sort", "latest")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}
	viewerID := c.GetUint("userID")

	var dtos []service.PostDTO
	var err error
	if sort == "following" {
		if viewerID == 0 {
			c.JSON(http.StatusOK, gin.H{"items": []service.PostDTO{}, "page": page, "page_size": pageSize})
			return
		}
		dtos, err = h.post.FollowingFeed(c.Request.Context(), viewerID, page, pageSize)
	} else {
		dtos, err = h.post.Feed(c.Request.Context(), sort, page, pageSize, viewerID)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": dtos, "page": page, "page_size": pageSize})
}

func (h *PostHandler) Detail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	viewerID := c.GetUint("userID")
	dto, err := h.post.Detail(c.Request.Context(), uint(id), viewerID)
	if err != nil {
		if err == service.ErrPostNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"post": dto})
}

func (h *PostHandler) Like(c *gin.Context) {
	userID := c.GetUint("userID")
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	liked, err := h.post.Like(c.Request.Context(), userID, uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"liked": liked})
}

func (h *PostHandler) Unlike(c *gin.Context) {
	userID := c.GetUint("userID")
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	unliked, err := h.post.Unlike(c.Request.Context(), userID, uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"unliked": unliked})
}

func (h *PostHandler) Search(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	if q == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请输入搜索关键词"})
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}
	viewerID := c.GetUint("userID")

	dtos, err := h.post.Search(c.Request.Context(), q, page, pageSize, viewerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "搜索失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": dtos, "page": page, "page_size": pageSize, "query": q})
}

func (h *PostHandler) MyPosts(c *gin.Context) {
	userID := c.GetUint("userID")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}
	dtos, err := h.post.MyPosts(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": dtos, "page": page, "page_size": pageSize})
}

func (h *PostHandler) MyLikes(c *gin.Context) {
	userID := c.GetUint("userID")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}
	dtos, err := h.post.MyLikes(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": dtos, "page": page, "page_size": pageSize})
}

func (h *PostHandler) Favorite(c *gin.Context) {
	userID := c.GetUint("userID")
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	favorited, err := h.post.Favorite(c.Request.Context(), userID, uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"favorited": favorited})
}

func (h *PostHandler) Unfavorite(c *gin.Context) {
	userID := c.GetUint("userID")
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	unfavorited, err := h.post.Unfavorite(c.Request.Context(), userID, uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"unfavorited": unfavorited})
}

func (h *PostHandler) MyFavorites(c *gin.Context) {
	userID := c.GetUint("userID")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}
	dtos, err := h.post.MyFavorites(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": dtos, "page": page, "page_size": pageSize})
}

func (h *PostHandler) UserPosts(c *gin.Context) {
	targetID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}
	viewerID := c.GetUint("userID")
	dtos, err := h.post.UserPosts(c.Request.Context(), uint(targetID), page, pageSize, viewerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": dtos, "page": page, "page_size": pageSize})
}
