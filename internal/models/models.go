package models

import (
	"time"
)

type User struct {
	ID           int64     `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	MagicCode    string    `json:"magic_code" db:"magic_code"`
	Verified     bool      `json:"verified" db:"verified"`
	LastLogin    time.Time `json:"last_login" db:"last_login"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type VerificationCode struct {
	ID        int64     `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	Code      string    `json:"code" db:"code"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	Used      bool      `json:"used" db:"used"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type Music struct {
	ID        int64     `json:"id" db:"id"`
	Title     string    `json:"title" db:"title"`
	Artist    string    `json:"artist" db:"artist"`
	URL       string    `json:"url" db:"url"`
	CoverURL  string    `json:"cover_url" db:"cover_url"`
	Duration  int       `json:"duration" db:"duration"` // 秒
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type LoginRequest struct {
	Email     string `json:"email" binding:"required,email"`
	MagicCode string `json:"magic_code" binding:"required"`
}

type VerifyRequest struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code" binding:"required"`
}

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func NewSuccessResponse(data interface{}) APIResponse {
	return APIResponse{
		Success: true,
		Message: "操作成功",
		Data:    data,
	}
}

func NewErrorResponse(message string) APIResponse {
	return APIResponse{
		Success: false,
		Message: message,
		Error:   message,
	}
}