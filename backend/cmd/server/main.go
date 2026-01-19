package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	redisv9 "github.com/redis/go-redis/v9"

	"url-shortener/backend/internal/handler"
	"url-shortener/backend/internal/service"
	storageredis "url-shortener/backend/internal/storage/redis"
)

func main() {
	// 读取环境变量
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "redis:6379"
	}
	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisDB := 0

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	// 初始化 Redis
	rdb, err := initRedis(redisAddr, redisPassword, redisDB)
	if err != nil {
		log.Fatalf("failed to initialize redis: %v", err)
	}
	defer func() { _ = rdb.Close() }()

	// 初始化 Repository
	repo := storageredis.NewRepository(rdb)

	// 初始化 Service
	linkService := service.NewLinkService(repo, baseURL)

	// 初始化 Handler
	linkHandler := handler.NewLinkHandler(linkService)

	// 设置 Gin 模式
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建路由
	r := gin.Default()

	// 添加 CORS 中间件（允许前端访问）
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Health check
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// API 路由
	api := r.Group("/api/v1")
	{
		api.POST("/shorten", linkHandler.Shorten)
		api.GET("/links/:code", linkHandler.GetLinkInfo)
	}

	// 短链接重定向路由（必须在最后，避免与其他路由冲突）
	r.GET("/:code", linkHandler.Redirect)

	// 启动服务器
	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

// initRedis 初始化 Redis 连接，带重试机制
func initRedis(addr, password string, db int) (*redisv9.Client, error) {
	var rdb *redisv9.Client
	var err error

	// 重试连接 Redis（等待 Redis 容器启动）
	maxRetries := 30
	for i := 0; i < maxRetries; i++ {
		rdb = redisv9.NewClient(&redisv9.Options{
			Addr:     addr,
			Password: password,
			DB:       db,
		})

		_, err = rdb.Ping(context.Background()).Result()
		if err != nil {
			log.Printf("failed to ping redis (attempt %d/%d): %v", i+1, maxRetries, err)
			_ = rdb.Close()
			time.Sleep(2 * time.Second)
			continue
		}
		log.Println("Redis connection established")
		break
	}

	if err != nil {
		return nil, err
	}

	return rdb, nil
}
