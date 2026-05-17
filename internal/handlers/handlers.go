package handlers

import (
	"log"
	"net/http"
	"time"

	"resend-auth-system/internal/config"
	"resend-auth-system/internal/database"
	"resend-auth-system/internal/models"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, cfg *config.Config) {
	// API路由组
	api := r.Group("/api")
	{
		// 认证相关
		api.POST("/login", handleLogin(cfg))
		api.POST("/verify", handleVerify)
		api.POST("/logout", handleLogout)
		api.GET("/me", handleGetCurrentUser)
		
		// 音乐相关
		api.GET("/music", handleGetMusicList)
		api.POST("/music", handleAddMusic)
		
		// 系统信息
		api.GET("/config", handleGetConfig(cfg))
		api.GET("/health", handleHealthCheck)
	}

	// 页面路由
	r.GET("/", handleIndex(cfg))
	r.GET("/login", handleLoginPage(cfg))
	r.GET("/dashboard", handleDashboardPage(cfg))
	r.GET("/music", handleMusicPage(cfg))
}

func handleLogin(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, models.NewErrorResponse("请求参数错误"))
			return
		}

		// 检查魔法验证码
		if req.MagicCode != cfg.Auth.MagicCode {
			c.JSON(http.StatusUnauthorized, models.NewErrorResponse("魔法验证码错误"))
			return
		}

		// 这里可以添加发送验证码到邮箱的逻辑（使用Resend API）
		log.Printf("User %s logged in with magic code", req.Email)
		
		// 创建验证码（简化版本，实际应该发送到邮箱）
		code := "123456" // 实际应该生成随机验证码
		
		c.JSON(http.StatusOK, models.NewSuccessResponse(gin.H{
			"email": req.Email,
			"code": code,
			"message": "验证码已发送到邮箱",
		}))
	}
}

func handleVerify(c *gin.Context) {
	var req models.VerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("请求参数错误"))
		return
	}

	// 验证验证码（简化版本）
	if req.Code != "123456" {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse("验证码错误"))
		return
	}

	// 创建或更新用户
	db := database.GetDB()
	_, err := db.Exec(`
		INSERT OR REPLACE INTO users (email, magic_code, verified, last_login, updated_at) 
		VALUES (?, ?, ?, ?, ?)
	`, req.Email, "magic_code_placeholder", true, time.Now(), time.Now())
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("用户创建失败"))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(gin.H{
		"email": req.Email,
		"verified": true,
		"token": "jwt_token_placeholder", // 实际应该生成JWT token
	}))
}

func handleLogout(c *gin.Context) {
	c.JSON(http.StatusOK, models.NewSuccessResponse(nil))
}

func handleGetCurrentUser(c *gin.Context) {
	// 从JWT token获取用户信息（简化版本）
	c.JSON(http.StatusOK, models.NewSuccessResponse(gin.H{
		"email": "user@example.com",
		"verified": true,
	}))
}

func handleGetMusicList(c *gin.Context) {
	db := database.GetDB()
	rows, err := db.Query("SELECT id, title, artist, url, cover_url, duration FROM music ORDER BY created_at DESC")
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("获取音乐列表失败"))
		return
	}
	defer rows.Close()

	var musicList []models.Music
	for rows.Next() {
		var music models.Music
		if err := rows.Scan(&music.ID, &music.Title, &music.Artist, &music.URL, &music.CoverURL, &music.Duration); err != nil {
			continue
		}
		musicList = append(musicList, music)
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(musicList))
}

func handleAddMusic(c *gin.Context) {
	var music models.Music
	if err := c.ShouldBindJSON(&music); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("请求参数错误"))
		return
	}

	db := database.GetDB()
	result, err := db.Exec(`
		INSERT INTO music (title, artist, url, cover_url, duration, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, music.Title, music.Artist, music.URL, music.CoverURL, music.Duration, time.Now())

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("添加音乐失败"))
		return
	}

	id, _ := result.LastInsertId()
	music.ID = id

	c.JSON(http.StatusOK, models.NewSuccessResponse(music))
}

func handleGetConfig(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, models.NewSuccessResponse(gin.H{
			"ui": cfg.UI,
			"auth": gin.H{
				"magic_code_required": cfg.Auth.MagicCode != "",
			},
		}))
	}
}

func handleHealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"timestamp": time.Now().Unix(),
	})
}

// 页面处理器
func handleIndex(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/login")
	}
}

func handleLoginPage(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 这里应该渲染HTML模板
		c.JSON(http.StatusOK, gin.H{
			"page": "login",
			"config": cfg.UI,
		})
	}
}

func handleDashboardPage(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"page": "dashboard",
			"config": cfg.UI,
		})
	}
}

func handleMusicPage(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"page": "music",
			"config": cfg.UI,
		})
	}
}