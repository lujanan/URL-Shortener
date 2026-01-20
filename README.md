# 🔗 URL Shortener - 短链接生成服务

一个功能完整、高性能的短链接生成服务，支持将长链接转换为短链接，并提供自动重定向、统计查询等功能。

## 📋 项目简介

本项目实现了一个全栈短链接生成服务，包含：
- **后端 API**：基于 Go + Gin 框架的 RESTful API 服务
- **前端界面**：基于 React + Vite 的现代化 Web 界面
- **数据存储**：使用 Redis 7.4 作为高性能键值存储
- **容器化部署**：支持 Docker Compose 一键启动

## ✨ 功能特性

### 核心功能

- ✅ **长链转短链**：支持将任意 HTTP/HTTPS 链接转换为短链接
- ✅ **随机短码生成**：自动生成 6-32 位随机短码，保证唯一性且不能是纯数字
- ✅ **自定义短码**：支持用户自定义短码（需满足 6-32 位，不能是纯数字）
- ✅ **短链重定向**：访问短链接自动 302 重定向到原始长链接
- ✅ **过期时间管理**：支持为短链接设置过期时间，过期后自动失效
- ✅ **访问统计**：记录每个短链接的访问次数和最后访问时间
- ✅ **短链查询**：支持查询短链接的详细信息（原始链接、创建时间、访问统计等）

### 技术特性

- 🚀 **高性能**：基于 Redis 的高并发读写，单实例可支持数百 QPS
- 🔒 **数据持久化**：Redis AOF 持久化，容器重启数据不丢失
- 🎨 **现代化 UI**：简洁美观的前端界面，支持响应式设计
- 🐳 **容器化部署**：Docker Compose 一键启动，开箱即用
- 🌐 **CORS 支持**：完整的前后端跨域支持

## 🛠 技术栈

### 后端

- **语言**：Go 1.23
- **框架**：Gin
- **存储**：Redis 7.4
- **容器化**：Docker + Docker Compose

### 前端

- **框架**：React 18
- **构建工具**：Vite
- **运行时**：Nginx 1.28-alpine

## 🚀 快速开始

### 前置要求

- Docker & Docker Compose
- Git

### 安装步骤

1. **克隆仓库**

```bash
git clone https://github.com/lujanan/URL-Shortener.git
cd URL-Shortener
```

2. **启动服务**

```bash
docker compose up -d
```

3. **访问服务**

- 前端界面：http://localhost:3000
- 后端 API：http://localhost:8080
- 健康检查：http://localhost:8080/healthz

### 停止服务

```bash
docker compose down
```

如需删除数据卷（清除 Redis 数据）：

```bash
docker compose down -v
```

## 📁 项目结构

```
URL-Shortener/
├── backend/                 # 后端服务
│   ├── cmd/
│   │   └── server/         # 主程序入口
│   ├── internal/
│   │   ├── handler/        # HTTP 处理器
│   │   ├── service/        # 业务逻辑层
│   │   ├── storage/        # 存储抽象层
│   │   │   └── redis/      # Redis 实现
│   │   ├── model/          # 数据模型
│   │   └── util/           # 工具函数
│   ├── Dockerfile          # 后端 Docker 镜像
│   └── go.mod              # Go 依赖管理
├── frontend/               # 前端应用
│   ├── src/
│   │   ├── App.tsx         # 主组件
│   │   ├── main.tsx        # 入口文件
│   │   └── styles.css      # 样式文件
│   ├── Dockerfile          # 前端 Docker 镜像
│   └── package.json        # Node 依赖
├── docker-compose.yml      # Docker Compose 配置
├── DESIGN.md              # 详细设计文档
└── README.md              # 项目说明（本文件）
```

## 📖 API 文档

### 1. 创建短链接

**请求**

```http
POST /api/v1/shorten
Content-Type: application/json

{
  "url": "https://example.com/very/long/url",
  "custom_code": "myalias",           // 可选：自定义短码
  "expire_at": "2026-12-31T23:59:59Z" // 可选：过期时间（ISO8601）
}
```

**响应**

```json
{
  "code": "a3K9mP2x",
  "short_url": "http://localhost:8080/a3K9mP2x",
  "long_url": "https://example.com/very/long/url",
  "expire_at": "2026-12-31T23:59:59Z"
}
```

### 2. 短链重定向

**请求**

```http
GET /{code}
```

**响应**

- 成功：`302 Found`，`Location: <原始长链接>`
- 失败：`404 Not Found`（短码不存在或已过期）

### 3. 查询短链信息

**请求**

```http
GET /api/v1/links/{code}
```

**响应**

```json
{
  "code": "a3K9mP2x",
  "long_url": "https://example.com/very/long/url",
  "created_at": "2026-01-19T10:00:00Z",
  "expire_at": "2026-12-31T23:59:59Z",
  "click_count": 123,
  "last_accessed_at": "2026-01-19T11:00:00Z"
}
```

### 4. 健康检查

**请求**

```http
GET /healthz
```

**响应**

```json
{
  "status": "ok"
}
```

## ⚙️ 配置说明

### 环境变量

#### 后端 (Backend)

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| `PORT` | 服务监听端口 | `8080` |
| `REDIS_ADDR` | Redis 地址 | `redis:6379` |
| `REDIS_PASSWORD` | Redis 密码 | 空 |
| `REDIS_DB` | Redis 数据库编号 | `0` |
| `BASE_URL` | 短链接基础 URL | `http://localhost:8080` |

#### Redis

使用 `redis:7.4` 镜像，默认开启 AOF 持久化。

### 短码规则

- **长度**：6-32 个字符
- **字符集**：`0-9A-Za-z`（Base62）
- **限制**：不能是纯数字，至少包含一个字母

## 🧪 开发指南

### 本地开发

#### 后端开发

```bash
cd backend
go mod download
go run cmd/server/main.go
```

#### 前端开发

```bash
cd frontend
npm install
npm run dev
```

### 构建镜像

```bash
# 构建后端镜像
docker build -t url-shortener-backend ./backend

# 构建前端镜像
docker build -t url-shortener-frontend ./frontend
```

## 📊 数据存储

### Redis 数据结构

- **短链记录**：`shortener:link:{code}` (Hash)
  - `long_url`: 原始长链接
  - `created_at`: 创建时间
  - `expire_at`: 过期时间（可选）
  - `click_count`: 访问次数
  - `last_accessed_at`: 最后访问时间

- **数据持久化**：通过 Redis AOF 和 Docker volume 实现持久化存储

## 🐛 故障排查

### 常见问题

1. **Redis 连接失败**
   - 检查 Redis 容器是否正常运行：`docker ps`
   - 检查环境变量 `REDIS_ADDR` 是否正确

2. **前端无法访问后端**
   - 检查 `BASE_URL` 环境变量配置
   - 确认后端服务已启动并可访问

3. **端口冲突**
   - 修改 `docker-compose.yml` 中的端口映射
   - 确保 3000（前端）和 8080（后端）端口未被占用

## 📝 更新日志

### v1.0.0 (2026-01-19)

- ✨ 初始版本发布
- ✨ 支持长链转短链功能
- ✨ 支持随机短码生成
- ✨ 支持自定义短码
- ✨ 支持短链重定向
- ✨ 支持访问统计
- ✨ 支持过期时间管理
- ✨ 完整的前后端界面

## 📄 许可证

本项目采用 MIT 许可证。

## 👥 贡献

欢迎提交 Issue 和 Pull Request！

## 📧 联系方式

如有问题或建议，请通过 GitHub Issues 联系。

---

**享受使用！** 🎉
