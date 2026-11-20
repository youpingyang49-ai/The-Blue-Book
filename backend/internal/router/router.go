package router

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/youpingyang49-ai/The-Blue-Book/backend/internal/handler"
	"github.com/youpingyang49-ai/The-Blue-Book/backend/internal/middleware"
	"github.com/youpingyang49-ai/The-Blue-Book/backend/internal/pkg/jwt"
	"github.com/youpingyang49-ai/The-Blue-Book/backend/internal/pkg/storage"
	"github.com/youpingyang49-ai/The-Blue-Book/backend/internal/repository"
	"github.com/youpingyang49-ai/The-Blue-Book/backend/internal/service"
)

func New(db *gorm.DB, rdb *redis.Client, jwtMgr *jwt.Manager, st storage.Storage, notifSvc *service.NotificationService, appEnv string) *gin.Engine {
	r := gin.Default()
	r.Use(cors())

	health := handler.NewHealthHandler(db)

	userRepo := repository.NewUserRepo(db)
	followRepo := repository.NewFollowRepo(db)
	topicRepo := repository.NewTopicRepo(db)
	userSvc := service.NewUserService(userRepo, followRepo)

	authSvc := service.NewAuthService(userRepo, rdb, jwtMgr)
	authHandler := handler.NewAuthHandler(authSvc, appEnv)
	userHandler := handler.NewUserHandler(userRepo, userSvc)

	postRepo := repository.NewPostRepo(db)
	rankRepo := repository.NewRankRepo(rdb)
	rankingSvc := service.NewRankingService(postRepo, topicRepo, rankRepo)
	// 首次启动用 MySQL 回填榜单，之后每 5 分钟纠偏一次；Redis 故障时自动降级为 SQL。
	go rankingSvc.Start(context.Background())

	postSvc := service.NewPostService(postRepo, followRepo, topicRepo, userRepo, notifSvc, st, rankingSvc)
	postHandler := handler.NewPostHandler(postSvc)
	uploadHandler := handler.NewUploadHandler(st)

	commentRepo := repository.NewCommentRepo(db)
	commentSvc := service.NewCommentService(commentRepo, postRepo, userRepo, notifSvc, rankingSvc)
	commentHandler := handler.NewCommentHandler(commentSvc)

	messageRepo := repository.NewMessageRepo(db)
	messageSvc := service.NewMessageService(messageRepo)
	messageHandler := handler.NewMessageHandler(messageSvc)

	followSvc := service.NewFollowService(followRepo, notifSvc)
	followHandler := handler.NewFollowHandler(followSvc)
	notifHandler := handler.NewNotificationHandler(notifSvc)

	api := r.Group("/api/v1")
	{
		api.GET("/health", health.Check)

		auth := api.Group("/auth")
		{
			auth.POST("/send-code", authHandler.SendCode)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.Refresh)
			auth.POST("/logout", authHandler.Logout)
		}

		api.GET("/rankings", middleware.OptionalAuth(jwtMgr, authSvc), postHandler.Rankings)
		api.GET("/posts", middleware.OptionalAuth(jwtMgr, authSvc), postHandler.List)
		api.GET("/posts/:id", middleware.OptionalAuth(jwtMgr, authSvc), postHandler.Detail)
		api.GET("/search", middleware.OptionalAuth(jwtMgr, authSvc), postHandler.Search)
		api.GET("/search/users", middleware.OptionalAuth(jwtMgr, authSvc), userHandler.Search)
		api.GET("/posts/:id/comments", middleware.OptionalAuth(jwtMgr, authSvc), commentHandler.List)
		api.GET("/users/:id", middleware.OptionalAuth(jwtMgr, authSvc), userHandler.Get)
		api.GET("/users/:id/posts", middleware.OptionalAuth(jwtMgr, authSvc), postHandler.UserPosts)
		api.GET("/users/:id/following", middleware.OptionalAuth(jwtMgr, authSvc), userHandler.Following)
		api.GET("/users/:id/followers", middleware.OptionalAuth(jwtMgr, authSvc), userHandler.Followers)

		protected := api.Group("")
		protected.Use(middleware.Auth(jwtMgr, authSvc))
		{
			protected.GET("/users/me", userHandler.Me)
			protected.GET("/users/me/posts", postHandler.MyPosts)
			protected.GET("/users/me/likes", postHandler.MyLikes)
			protected.GET("/users/me/favorites", postHandler.MyFavorites)
			protected.POST("/upload", uploadHandler.Upload)
			protected.POST("/posts", postHandler.Create)
			protected.PUT("/posts/:id", postHandler.Update)
			protected.DELETE("/posts/:id", postHandler.Delete)
			protected.POST("/posts/:id/like", postHandler.Like)
			protected.DELETE("/posts/:id/like", postHandler.Unlike)
			protected.POST("/posts/:id/favorite", postHandler.Favorite)
			protected.DELETE("/posts/:id/favorite", postHandler.Unfavorite)
			protected.POST("/posts/:id/comments", commentHandler.Create)
			protected.POST("/users/:id/follow", followHandler.Follow)
			protected.DELETE("/users/:id/follow", followHandler.Unfollow)
			protected.GET("/notifications", notifHandler.List)
			protected.GET("/notifications/unread-count", notifHandler.UnreadCount)
			protected.POST("/notifications/read", notifHandler.MarkAllRead)
			protected.GET("/conversations", messageHandler.Conversations)
			protected.GET("/messages/:id", messageHandler.Conversation)
			protected.POST("/messages/:id", messageHandler.Send)
			protected.POST("/messages/:id/unread", messageHandler.MarkUnread)
			protected.DELETE("/messages/:id", messageHandler.Delete)
			protected.POST("/conversations/:id/hide", messageHandler.Hide)
		}
	}

	return r
}

func cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
