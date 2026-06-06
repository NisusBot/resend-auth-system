package handlers

import (
	"log"
	"net/http"
	"strings"
	"time"

	"resend-auth-system/internal/auth"
	"resend-auth-system/internal/config"
	"resend-auth-system/internal/database"
	"resend-auth-system/internal/models"

	"github.com/gin-gonic/gin"
)

type Handlers struct {
	cfg         *config.Config
	authService *auth.AuthService
}

func NewHandlers(cfg *config.Config) *Handlers {
	return &Handlers{
		cfg:         cfg,
		authService: auth.NewAuthService(cfg),
	}
}

func RegisterRoutes(r *gin.Engine, cfg *config.Config) {
	h := NewHandlers(cfg)

	// API路由组
	api := r.Group("/api")
	{
		// 认证相关
		api.POST("/login", h.handleLogin)
		api.POST("/verify", h.handleVerify)
		api.POST("/logout", h.handleLogout)
		api.GET("/me", h.AuthMiddleware(), h.handleGetCurrentUser)
		
		// 音乐相关
		api.GET("/music", h.AuthMiddleware(), h.handleGetMusicList)
		api.POST("/music", h.AuthMiddleware(), h.handleAddMusic)
		
		// 系统信息
		api.GET("/config", h.handleGetConfig)
		api.GET("/health", h.handleHealthCheck)
	}

	// 页面路由
	r.GET("/", h.handleIndex)
	r.GET("/login", h.handleLoginPage)
	r.GET("/dashboard", h.PageAuthMiddleware(), h.handleDashboardPage)
	r.GET("/music", h.PageAuthMiddleware(), h.handleMusicPage)
}

// AuthMiddleware 认证中间件
func (h *Handlers) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			// 尝试从cookie获取
			tokenString, _ = c.Cookie("auth_token")
		}

		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, models.NewErrorResponse("未授权访问"))
			c.Abort()
			return
		}

		// 移除Bearer前缀
		tokenString = strings.TrimPrefix(tokenString, "Bearer ")

		claims, err := h.authService.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, models.NewErrorResponse("token无效或已过期"))
			c.Abort()
			return
		}

		// 检查用户是否已验证
		if !claims.Verified {
			c.JSON(http.StatusForbidden, models.NewErrorResponse("请先验证邮箱"))
			c.Abort()
			return
		}

		// 将用户信息保存到上下文
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_verified", claims.Verified)

		c.Next()
	}
}

func (h *Handlers) handleLogin(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("请求参数错误"))
		return
	}

	// 使用认证服务登录
	user, token, err := h.authService.Login(req.Email, req.MagicCode)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse(err.Error()))
		return
	}

	// 生成验证码（简化版本，实际应该发送到邮箱）
	code, err := h.authService.CreateVerificationCode(req.Email)
	if err != nil {
		log.Printf("Failed to create verification code: %v", err)
		// 继续执行，使用默认验证码
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(gin.H{
		"user": gin.H{
			"id":        user.ID,
			"email":     user.Email,
			"verified":  user.Verified,
			"last_login": user.LastLogin,
		},
		"token": token,
		"verification_code": code, // 简化版本，实际不应该返回给前端
		"message": "验证码已发送到邮箱",
	}))
}

func (h *Handlers) handleVerify(c *gin.Context) {
	var req models.VerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse("请求参数错误"))
		return
	}

	// 验证验证码
	valid, err := h.authService.CheckVerificationCode(req.Email, req.Code)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewErrorResponse(err.Error()))
		return
	}

	if !valid {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse("验证码错误"))
		return
	}

	// 更新用户验证状态并生成新token
	user, token, err := h.authService.VerifyCode(req.Email, req.Code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewErrorResponse("验证失败"))
		return
	}

	// 设置token到cookie（可选）
	c.SetCookie("auth_token", token, int(time.Hour.Seconds()*24), "/", "", false, true)

	c.JSON(http.StatusOK, models.NewSuccessResponse(gin.H{
		"user": gin.H{
			"id":        user.ID,
			"email":     user.Email,
			"verified":  user.Verified,
			"last_login": user.LastLogin,
		},
		"token": token,
		"message": "验证成功",
	}))
}

func (h *Handlers) handleLogout(c *gin.Context) {
	// 清除cookie
	c.SetCookie("auth_token", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, models.NewSuccessResponse(gin.H{
		"message": "退出成功",
	}))
}

func (h *Handlers) handleGetCurrentUser(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		tokenString, _ = c.Cookie("auth_token")
	}
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	user, err := h.authService.GetCurrentUser(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.NewErrorResponse("获取用户信息失败"))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccessResponse(gin.H{
		"id":        user.ID,
		"email":     user.Email,
		"verified":  user.Verified,
		"last_login": user.LastLogin,
		"created_at": user.CreatedAt,
	}))
}

func (h *Handlers) handleGetMusicList(c *gin.Context) {
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

func (h *Handlers) handleAddMusic(c *gin.Context) {
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

func (h *Handlers) handleGetConfig(c *gin.Context) {
	c.JSON(http.StatusOK, models.NewSuccessResponse(gin.H{
		"ui": h.cfg.UI,
		"auth": gin.H{
			"magic_code_required": h.cfg.Auth.MagicCode != "",
		},
	}))
}

func (h *Handlers) handleHealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"timestamp": time.Now().Unix(),
		"service": "resend-auth-system",
		"version": "1.0.0",
	})
}

// 页面处理器
func (h *Handlers) handleIndex(c *gin.Context) {
	c.Redirect(http.StatusFound, "/login")
}

// PageAuthMiddleware 页面专用认证中间件
// 未授权或未验证时跳转 /login(而不是返回 401 JSON)
func (h *Handlers) PageAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			// 尝试从 cookie 获取
			tokenString, _ = c.Cookie("auth_token")
		}

		if tokenString == "" {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		// 移除 Bearer 前缀
		tokenString = strings.TrimPrefix(tokenString, "Bearer ")

		claims, err := h.authService.ValidateToken(tokenString)
		if err != nil {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		if !claims.Verified {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		// 将用户信息保存到上下文,供模板渲染使用
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_verified", claims.Verified)

		c.Next()
	}
}

func (h *Handlers) handleLoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", gin.H{
		"Title":           h.cfg.UI.Title,
		"Theme":           h.cfg.UI.Theme,
		"BackgroundColor": h.cfg.UI.BackgroundColor,
		"PrimaryColor":    h.cfg.UI.PrimaryColor,
		"AccentColor":     h.cfg.UI.AccentColor,
	})
}

func (h *Handlers) handleDashboardPage(c *gin.Context) {
	userEmail, _ := c.Get("user_email")
	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"Title":           h.cfg.UI.Title,
		"Theme":           h.cfg.UI.Theme,
		"BackgroundColor": h.cfg.UI.BackgroundColor,
		"PrimaryColor":    h.cfg.UI.PrimaryColor,
		"AccentColor":     h.cfg.UI.AccentColor,
		"UserEmail":       userEmail,
	})
}

func (h *Handlers) handleMusicPage(c *gin.Context) {
	userEmail, _ := c.Get("user_email")
	c.HTML(http.StatusOK, "music.html", gin.H{
		"Title":           h.cfg.UI.Title,
		"Theme":           h.cfg.UI.Theme,
		"BackgroundColor": h.cfg.UI.BackgroundColor,
		"PrimaryColor":    h.cfg.UI.PrimaryColor,
		"AccentColor":     h.cfg.UI.AccentColor,
		"UserEmail":       userEmail,
	})
}
