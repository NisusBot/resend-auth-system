package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"resend-auth-system/internal/config"
	"resend-auth-system/internal/database"
	"resend-auth-system/internal/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	cfg := config.GetConfig()

	// 初始化数据库
	if err := database.InitDB(cfg.Database.Path); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.CloseDB()

	// 设置Gin模式
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建Gin引擎
	r := gin.Default()

	// 配置CORS
	if cfg.Security.EnableCORS {
		r.Use(func(c *gin.Context) {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			if c.Request.Method == "OPTIONS" {
				c.AbortWithStatus(http.StatusNoContent)
				return
			}
			c.Next()
		})
	}

	// 加载 HTML 模板
	r.LoadHTMLGlob("templates/*.html")

	// 设置静态文件服务
	r.Static("/static", "./static")
	r.StaticFile("/favicon.ico", "./static/favicon.ico")

	// 注册路由
	handlers.RegisterRoutes(r, cfg)

	// 启动服务器
	addr := cfg.Server.Host + ":" + cfg.Server.Port
	log.Printf("Starting server on %s in %s mode", addr, cfg.Server.Mode)

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := r.Run(addr); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	<-quit
	log.Println("Shutting down server...")
	log.Println("Server stopped")
}