package auth

import (
	"errors"
	"fmt"
	"log"
	"time"

	"resend-auth-system/internal/config"
	"resend-auth-system/internal/database"
	"resend-auth-system/internal/models"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenExpired       = errors.New("token expired")
)

type Claims struct {
	Email    string `json:"email"`
	UserID   int64  `json:"user_id"`
	Verified bool   `json:"verified"`
	jwt.RegisteredClaims
}

type AuthService struct {
	cfg *config.Config
}

func NewAuthService(cfg *config.Config) *AuthService {
	return &AuthService{
		cfg: cfg,
	}
}

// Login 用户登录
func (s *AuthService) Login(email, magicCode string) (*models.User, string, error) {
	// 验证魔法验证码
	if magicCode != s.cfg.Auth.MagicCode {
		return nil, "", ErrInvalidCredentials
	}

	db := database.GetDB()
	
	// 查找用户
	var user models.User
	err := db.QueryRow(`
		SELECT id, email, magic_code, verified, last_login, created_at, updated_at 
		FROM users WHERE email = ?
	`, email).Scan(&user.ID, &user.Email, &user.MagicCode, &user.Verified, &user.LastLogin, &user.CreatedAt, &user.UpdatedAt)
	
	if err != nil {
		// 用户不存在，创建新用户
		result, err := db.Exec(`
			INSERT INTO users (email, magic_code, verified, last_login) 
			VALUES (?, ?, ?, ?)
		`, email, magicCode, false, time.Now())
		
		if err != nil {
			return nil, "", fmt.Errorf("failed to create user: %v", err)
		}
		
		id, _ := result.LastInsertId()
		user = models.User{
			ID:        id,
			Email:     email,
			MagicCode: magicCode,
			Verified:  false,
			LastLogin: time.Now(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
	} else {
		// 更新最后登录时间
		_, err := db.Exec(`UPDATE users SET last_login = ? WHERE id = ?`, time.Now(), user.ID)
		if err != nil {
			log.Printf("Failed to update last login: %v", err)
		}
		user.LastLogin = time.Now()
	}
	
	// 生成JWT token（即使未验证也生成，但会标记为未验证）
	token, err := s.GenerateToken(user.ID, user.Email, user.Verified)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %v", err)
	}
	
	return &user, token, nil
}

// VerifyCode 验证邮箱验证码
func (s *AuthService) VerifyCode(email, code string) (*models.User, string, error) {
	db := database.GetDB()
	
	// 在实际项目中，这里应该验证验证码是否匹配且未过期
	// 简化版本：直接接受任何123456验证码
	if code != "123456" {
		return nil, "", ErrInvalidCredentials
	}
	
	// 更新用户验证状态
	_, err := db.Exec(`
		UPDATE users SET verified = TRUE, updated_at = ? WHERE email = ?
	`, time.Now(), email)
	
	if err != nil {
		return nil, "", fmt.Errorf("failed to update user verification: %v", err)
	}
	
	// 获取用户信息
	var user models.User
	err = db.QueryRow(`
		SELECT id, email, magic_code, verified, last_login, created_at, updated_at 
		FROM users WHERE email = ?
	`, email).Scan(&user.ID, &user.Email, &user.MagicCode, &user.Verified, &user.LastLogin, &user.CreatedAt, &user.UpdatedAt)
	
	if err != nil {
		return nil, "", fmt.Errorf("failed to get user: %v", err)
	}
	
	// 生成新的JWT token（已验证）
	token, err := s.GenerateToken(user.ID, user.Email, true)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %v", err)
	}
	
	return &user, token, nil
}

// GenerateToken 生成JWT token
func (s *AuthService) GenerateToken(userID int64, email string, verified bool) (string, error) {
	claims := Claims{
		Email:    email,
		UserID:   userID,
		Verified: verified,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * time.Duration(s.cfg.Auth.SessionTimeout))),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "resend-auth-system",
		},
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.Auth.JWTSecret))
}

// ValidateToken 验证JWT token
func (s *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.cfg.Auth.JWTSecret), nil
	})
	
	if err != nil {
		return nil, err
	}
	
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	
	return nil, ErrInvalidToken
}

// GetCurrentUser 从token获取当前用户
func (s *AuthService) GetCurrentUser(tokenString string) (*models.User, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}
	
	db := database.GetDB()
	var user models.User
	err = db.QueryRow(`
		SELECT id, email, magic_code, verified, last_login, created_at, updated_at 
		FROM users WHERE id = ?
	`, claims.UserID).Scan(&user.ID, &user.Email, &user.MagicCode, &user.Verified, &user.LastLogin, &user.CreatedAt, &user.UpdatedAt)
	
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %v", err)
	}
	
	return &user, nil
}

// CreateVerificationCode 创建验证码（简化版本）
func (s *AuthService) CreateVerificationCode(email string) (string, error) {
	// 在实际项目中，这里应该生成随机验证码并保存到数据库
	// 简化版本：总是返回123456
	code := "123456"
	
	db := database.GetDB()
	_, err := db.Exec(`
		INSERT INTO verification_codes (email, code, expires_at, used) 
		VALUES (?, ?, ?, ?)
	`, email, code, time.Now().Add(10*time.Minute), false)
	
	if err != nil {
		return "", fmt.Errorf("failed to create verification code: %v", err)
	}
	
	return code, nil
}

// CheckVerificationCode 检查验证码
func (s *AuthService) CheckVerificationCode(email, code string) (bool, error) {
	db := database.GetDB()
	
	var (
		storedCode   string
		expiresAt    time.Time
		used         bool
	)
	
	err := db.QueryRow(`
		SELECT code, expires_at, used 
		FROM verification_codes 
		WHERE email = ? AND used = FALSE 
		ORDER BY created_at DESC LIMIT 1
	`, email).Scan(&storedCode, &expiresAt, &used)
	
	if err != nil {
		return false, fmt.Errorf("failed to get verification code: %v", err)
	}
	
	if time.Now().After(expiresAt) {
		return false, errors.New("verification code expired")
	}
	
	if used {
		return false, errors.New("verification code already used")
	}
	
	if storedCode != code {
		return false, nil
	}
	
	// 标记为已使用
	_, err = db.Exec(`UPDATE verification_codes SET used = TRUE WHERE email = ? AND code = ?`, email, code)
	if err != nil {
		return false, fmt.Errorf("failed to update verification code: %v", err)
	}
	
	return true, nil
}