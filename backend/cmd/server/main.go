package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"

	"url-shortener/backend/internal/handler"
	"url-shortener/backend/internal/service"
	"url-shortener/backend/internal/storage/mysql"
)

func main() {
	// 读取环境变量
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbDSN := os.Getenv("DB_DSN")
	if dbDSN == "" {
		dbDSN = "shortuser:shortpass@tcp(mysql:3306)/shorturl?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci"
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	// 初始化数据库连接
	db, err := initDatabase(dbDSN)
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
	defer db.Close()

	// 初始化数据库表结构
	if err := mysql.InitSchema(db); err != nil {
		log.Fatalf("failed to init database schema: %v", err)
	}

	// 初始化 Repository
	repo := mysql.NewRepository(db)

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

// initDatabase 初始化数据库连接，带重试机制
func initDatabase(dsn string) (*sql.DB, error) {
	var db *sql.DB
	var err error

	// 重试连接数据库（等待 MySQL 容器启动）
	maxRetries := 30
	for i := 0; i < maxRetries; i++ {
		db, err = sql.Open("mysql", dsn)
		if err != nil {
			log.Printf("failed to open database (attempt %d/%d): %v", i+1, maxRetries, err)
			time.Sleep(2 * time.Second)
			continue
		}

		// 测试连接
		if err = db.Ping(); err != nil {
			log.Printf("failed to ping database (attempt %d/%d): %v", i+1, maxRetries, err)
			db.Close()
			time.Sleep(2 * time.Second)
			continue
		}

		// 连接成功
		log.Println("Database connection established")
		break
	}

	if err != nil {
		return nil, err
	}

	// 设置连接池参数
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}
