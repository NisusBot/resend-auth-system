package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDB(dbPath string) error {
	// 确保数据目录存在
	dataDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %v", err)
	}

	// 打开数据库
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}

	// 设置连接池参数
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	// 创建表
	if err := createTables(db); err != nil {
		return fmt.Errorf("failed to create tables: %v", err)
	}

	DB = db
	log.Println("Database initialized successfully")
	return nil
}

func createTables(db *sql.DB) error {
	// 创建用户表
	userTable := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT UNIQUE NOT NULL,
		magic_code TEXT NOT NULL,
		verified BOOLEAN DEFAULT FALSE,
		last_login TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)
	`

	// 创建验证码表
	verificationTable := `
	CREATE TABLE IF NOT EXISTS verification_codes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT NOT NULL,
		code TEXT NOT NULL,
		expires_at TIMESTAMP NOT NULL,
		used BOOLEAN DEFAULT FALSE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)
	`

	// 创建音乐表
	musicTable := `
	CREATE TABLE IF NOT EXISTS music (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		artist TEXT NOT NULL,
		url TEXT NOT NULL,
		cover_url TEXT,
		duration INTEGER DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)
	`

	// 创建触发器：更新updated_at
	trigger := `
	CREATE TRIGGER IF NOT EXISTS update_users_updated_at
	AFTER UPDATE ON users
	FOR EACH ROW
	BEGIN
		UPDATE users SET updated_at = CURRENT_TIMESTAMP WHERE id = OLD.id;
	END;
	`

	// 执行SQL
	tables := []string{userTable, verificationTable, musicTable, trigger}
	for _, tableSQL := range tables {
		_, err := db.Exec(tableSQL)
		if err != nil {
			return fmt.Errorf("failed to create table: %v\nSQL: %s", err, tableSQL)
		}
	}

	// 插入示例音乐数据
	insertSampleMusic := `
	INSERT OR IGNORE INTO music (title, artist, url, cover_url, duration) VALUES
	('Blinding Lights', 'The Weeknd', 'https://example.com/music1.mp3', 'https://example.com/cover1.jpg', 200),
	('Shape of You', 'Ed Sheeran', 'https://example.com/music2.mp3', 'https://example.com/cover2.jpg', 233),
	('Bad Guy', 'Billie Eilish', 'https://example.com/music3.mp3', 'https://example.com/cover3.jpg', 194),
	('Watermelon Sugar', 'Harry Styles', 'https://example.com/music4.mp3', 'https://example.com/cover4.jpg', 174),
	('Levitating', 'Dua Lipa', 'https://example.com/music5.mp3', 'https://example.com/cover5.jpg', 203)
	`
	_, err := db.Exec(insertSampleMusic)
	if err != nil {
		log.Printf("Warning: Failed to insert sample music data: %v", err)
	}

	return nil
}

func CloseDB() {
	if DB != nil {
		DB.Close()
	}
}

func GetDB() *sql.DB {
	return DB
}