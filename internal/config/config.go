package config

import (
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Auth     AuthConfig     `yaml:"auth"`
	Resend   ResendConfig   `yaml:"resend"`
	UI       UIConfig       `yaml:"ui"`
	Security SecurityConfig `yaml:"security"`
}

type ServerConfig struct {
	Port string `yaml:"port" env:"APP_PORT"`
	Host string `yaml:"host" env:"APP_HOST"`
	Mode string `yaml:"mode" env:"APP_MODE"`
}

type DatabaseConfig struct {
	Driver       string `yaml:"driver" env:"DB_DRIVER"`
	Path         string `yaml:"path" env:"DB_PATH"`
	MaxOpenConns int    `yaml:"max_open_conns" env:"DB_MAX_OPEN_CONNS"`
	MaxIdleConns int    `yaml:"max_idle_conns" env:"DB_MAX_IDLE_CONNS"`
}

type AuthConfig struct {
	MagicCode     string `yaml:"magic_code" env:"MAGIC_CODE"`
	SessionTimeout int   `yaml:"session_timeout" env:"SESSION_TIMEOUT"`
	JWTSecret     string `yaml:"jwt_secret" env:"JWT_SECRET"`
}

type ResendConfig struct {
	APIKey    string `yaml:"api_key" env:"RESEND_API_KEY"`
	FromEmail string `yaml:"from_email" env:"RESEND_FROM_EMAIL"`
	FromName  string `yaml:"from_name" env:"RESEND_FROM_NAME"`
}

type UIConfig struct {
	Title           string `yaml:"title" env:"UI_TITLE"`
	Theme           string `yaml:"theme" env:"UI_THEME"`
	BackgroundColor string `yaml:"background_color" env:"UI_BACKGROUND_COLOR"`
	PrimaryColor    string `yaml:"primary_color" env:"UI_PRIMARY_COLOR"`
	AccentColor     string `yaml:"accent_color" env:"UI_ACCENT_COLOR"`
}

type SecurityConfig struct {
	RateLimit      int      `yaml:"rate_limit" env:"RATE_LIMIT"`
	EnableCORS     bool     `yaml:"enable_cors" env:"ENABLE_CORS"`
	AllowedOrigins []string `yaml:"allowed_origins" env:"ALLOWED_ORIGINS"`
}

var AppConfig *Config

func LoadConfig() error {
	// 加载.env文件
	_ = godotenv.Load()
	
	// 读取YAML配置文件
	configFile := "config/config.yaml"
	if env := os.Getenv("APP_ENV"); env == "production" {
		configFile = "config/config.production.yaml"
	}
	
	data, err := os.ReadFile(configFile)
	if err != nil {
		log.Printf("Warning: Config file %s not found, using defaults: %v", configFile, err)
		AppConfig = &Config{
			Server: ServerConfig{
				Port: "8080",
				Host: "0.0.0.0",
				Mode: "debug",
			},
			Database: DatabaseConfig{
				Driver:       "sqlite3",
				Path:         "./data/auth.db",
				MaxOpenConns: 10,
				MaxIdleConns: 5,
			},
			Auth: AuthConfig{
				MagicCode:     "123456",
				SessionTimeout: 24,
				JWTSecret:     "your-secret-key-change-this",
			},
			UI: UIConfig{
				Title:           "Resend邮箱验证码审核系统",
				Theme:           "dark",
				BackgroundColor: "#121212",
				PrimaryColor:    "#1db954",
				AccentColor:     "#ffffff",
			},
		}
		return nil
	}
	
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return err
	}
	
	AppConfig = &config
	
	// 从环境变量覆盖配置
	overrideFromEnv()
	
	return nil
}

func overrideFromEnv() {
	// 这里可以实现从环境变量覆盖配置的逻辑
	// 简化起见，我们只使用YAML配置
}

func GetConfig() *Config {
	if AppConfig == nil {
		if err := LoadConfig(); err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}
	}
	return AppConfig
}