# Resend邮箱验证码审核页系统

基于Resend邮箱验证码服务的审核页系统，提供邮箱验证码注册和魔法验证码机制，仿照道理鱼音乐登录页设计。

## 功能特性

- ✅ 邮箱验证码注册登录
- ✅ 魔法验证码机制
- ✅ 深色主题UI（#121ÿ212背景）
- ✅ 响应式设计
- ✅ 可配置系统标题、主题、验证码
- ✅ 支持Docker一键部署
- ✅ SQLite数据库
- ✅ RESTful API

## 技术栈

- **后端**: Go 1.22 + Gin框架
- **数据库**: SQLite3
- **前端**: 原生HTML/CSS/JavaScript
- **部署**: Docker + Docker Compose
- **配置**: YAML + 环境变量

## 快速开始

### 1. 环境要求

- Go 1.20+
- SQLite3
- Docker (可选)

### 2. 克隆项目

```bash
git clone https://github.com/yourusername/resend-auth-system.git
cd resend-auth-system
```

### 3. 配置环境

```bash
# 复制环境变量模板
cp .env.example .env

# 编辑配置文件
vim config/config.yaml
```

### 4. 安装依赖

```bash
make setup
```

### 5. 运行项目

```bash
# 开发模式
make dev

# 或者直接运行
go run ./cmd/server
```

### 6. 访问系统

打开浏览器访问：http://localhost:8080

## Docker部署

### 使用Docker Compose

```bash
# 启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down
```

### 使用Docker直接运行

```bash
# 构建镜像
docker build -t resend-auth-system .

# 运行容器
docker run -p 8080:8080 \
  -e RESEND_API_KEY=your_api_key \
  -e MAGIC_CODE=your_magic_code \
  -v ./data:/app/data \
  resend-auth-system
```

## 配置文件

### 环境变量 (.env)

```bash
# 应用配置
APP_ENV=development
APP_PORT=8080

# 数据库配置
DB_DRIVER=sqlite3
DB_PATH=./data/auth.db

# 认证配置
MAGIC_CODE=123456
JWT_SECRET=your-secret-key-change-this-in-production

# Resend邮箱服务
RESEND_API_KEY=your-resend-api-key
RESEND_FROM_EMAIL=noreply@example.com
RESEND_FROM_NAME=Resend Auth System

# UI配置
UI_TITLE=Resend邮箱验证码审核系统
UI_THEME=dark
UI_BACKGROUND_COLOR=#121212
UI_PRIMARY_COLOR=#1db954
```

### YAML配置 (config/config.yaml)

```yaml
server:
  port: 8080
  host: "0.0.0.0"
  mode: "debug"

database:
  driver: "sqlite3"
  path: "./data/auth.db"
  max_open_conns: 10
  max_idle_conns: 5

auth:
  magic_code: "123456"
  session_timeout: 24
  jwt_secret: "your-secret-key-change-this"

resend:
  api_key: ""
  from_email: "noreply@example.com"
  from_name: "Resend Auth System"

ui:
  title: "Resend邮箱验证码审核系统"
  theme: "dark"
  background_color: "#121212"
  primary_color: "#1db954"
  accent_color: "#ffffff"
```

## API接口

### 认证接口

#### 1. 登录/发送验证码
```
POST /api/login
Content-Type: application/json

{
  "email": "user@example.com",
  "magic_code": "123456"
}
```

响应：
```json
{
  "success": true,
  "message": "验证码已发送到邮箱",
  "data": {
    "email": "user@example.com",
    "code": "123456"
  }
}
```

#### 2. 验证验证码
```
POST /api/verify
Content-Type: application/json

{
  "email": "user@example.com",
  "code": "123456"
}
```

响应：
```json
{
  "success": true,
  "message": "操作成功",
  "data": {
    "email": "user@example.com",
    "verified": true,
    "token": "jwt_token_placeholder"
  }
}
```

### 音乐接口

#### 1. 获取音乐列表
```
GET /api/music
```

#### 2. 添加音乐
```
POST /api/music
Content-Type: application/json

{
  "title": "歌曲标题",
  "artist": "歌手",
  "url": "音乐URL",
  "cover_url": "封面URL",
  "duration": 200
}
```

### 系统接口

#### 1. 健康检查
```
GET /api/health
```

#### 2. 获取配置
```
GET /api/config
```

## 项目结构

```
resend-auth-system/
├── cmd/
│   └── server/
│       └── main.go          # 主程序入口
├── internal/
│   ├── config/
│   │   └── config.go        # 配置管理
│   ├── database/
│   │   └── database.go      # 数据库初始化
│   ├── handlers/
│   │   └── handlers.go      # HTTP处理器
│   └── models/
│       └── models.go        # 数据模型
├── static/                  # 静态文件
├── templates/               # HTML模板
│   └── login.html
├── config/
│   └── config.yaml          # 配置文件
├── data/                    # 数据库文件
├── migrations/              # 数据库迁移
├── .env.example            # 环境变量模板
├── Dockerfile              # Docker构建文件
├── docker-compose.yml      # Docker Compose配置
├── Makefile                # 构建脚本
├── go.mod                  # Go模块定义
└── README.md               # 项目文档
```

## 开发指南

### 添加新功能

1. 在 `internal/models/models.go` 中添加数据模型
2. 在 `internal/database/database.go` 中添加数据库表
3. 在 `internal/handlers/handlers.go` 中添加处理器
4. 在 `cmd/server/main.go` 中注册路由

### 添加新页面

1. 在 `templates/` 目录下创建HTML模板
2. 在 `handlers.go` 中添加页面处理器
3. 在 `main.go` 中注册路由

### 测试

```bash
# 运行所有测试
make test

# 运行特定包测试
go test ./internal/handlers -v
```

## 部署到生产环境

### 1. 配置生产环境变量

```bash
cp .env.example .env.production
vim .env.production
```

### 2. 构建生产镜像

```bash
docker build -t resend-auth-system:production .
```

### 3. 使用Docker Compose部署

```bash
docker-compose -f docker-compose.production.yml up -d
```

### 4. 配置反向代理 (Nginx)

```nginx
server {
    listen 80;
    server_name your-domain.com;
    
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## 故障排除

### 常见问题

1. **数据库连接失败**
   - 检查 `data/` 目录权限
   - 确保SQLite3已安装

2. **验证码发送失败**
   - 检查Resend API密钥配置
   - 验证邮箱格式

3. **端口被占用**
   - 修改 `config/config.yaml` 中的端口号
   - 检查是否有其他服务占用8080端口

### 日志查看

```bash
# 查看应用日志
tail -f logs/app.log

# 查看Docker容器日志
docker-compose logs -f
```

## 许可证

MIT License

## 贡献

欢迎提交Issue和Pull Request！

## 联系方式

如有问题，请通过Issue或邮件联系。