### 项目设计文档：纯后端短链接生成服务（URL Shortener）

---

## 一、项目概述

- **项目目标**：实现一个将长链接转换为短链接的后端 REST API 服务，支持访问短链接时自动 302 重定向到原始长链接。
- **使用场景**：营销活动链接缩短、社媒分享、内部工具链接统一管理等。
- **交付要求对齐**：满足 README 中题目 B 要求，服务与数据库通过 Docker 化交付，并可通过 `docker compose up` 一键启动（根目录保留 `frontend/` 与 `backend/` 的结构，核心逻辑在 `backend/`）。

---

## 二、功能需求

- **核心功能**
  - **长链转短链**
    - 接口：`POST /api/v1/shorten`
    - 输入长链接 URL，可选自定义短码、可选过期时间。
    - 返回短链接完整地址和短码。
  - **短链重定向**
    - 接口：`GET /{code}`
    - 根据短码查找原始 URL，返回 `302 Found` 重定向到长链接。
  - **短链查询（管理用，可选）**
    - 接口：`GET /api/v1/links/{code}`
    - 返回该短链的原始 URL、创建时间、访问次数、过期时间等信息。

- **业务规则**
  - 支持校验输入 URL（协议必须为 `http` 或 `https`）。
  - 支持短链过期时间（过期后访问返回 404 或 410）。
  - 短码唯一，不允许与已有记录冲突。
  - 记录每次访问的访问次数（可选记录最后访问时间）。

---

## 三、非功能需求

- **性能**：单实例下每秒至少数百请求（QPS）无明显性能瓶颈。
- **可用性**：重启后数据不丢失（依赖持久化数据库）。
- **可扩展性**：短码生成算法和存储层可替换（如后续迁移到 Redis / MySQL / 分布式方案）。
- **安全性**
  - 限制 URL 长度（例如不超过 2048 字符）。
  - 简单防滥用能力（可通过 IP 级别的限流中间件预留扩展点）。
- **可观测性**：基本日志（请求/响应、错误日志）和健康检查。

---

## 四、技术选型

- **编程语言**：Go
  - 原因：与当前工作路径一致（`/home/go/src`），生态成熟，适合高并发短服务。
- **Web 框架**：Gin（或其他轻量框架）
  - 理由：路由直观、中间件生态完善、开发效率高。
- **数据库**：Redis
  - 使用 Redis 作为 KV 存储，短码 -> 记录（Hash）映射，读写路径简单、延迟低。
  - 通过 `redis-server --appendonly yes` 开启 AOF，配合 Docker volume 持久化数据。
- **缓存（可选扩展）**：Redis
  - 本项目已直接使用 Redis 作为主存储；如后续引入其他数据库，可将 Redis 降级为缓存层。
- **配置管理**：环境变量（`PORT`、`REDIS_ADDR`、`REDIS_PASSWORD`、`REDIS_DB`、`BASE_URL` 等）。
- **容器化**：Docker + docker-compose
  - `backend` 服务打包为一个镜像，通过 MySQL 持久化数据。
  - `frontend` 使用 React + Vite 构建，采用 Nginx 镜像对外提供静态资源。

---

## 五、系统架构设计

- **整体架构**
  - **API 服务（Go）**
    - HTTP 服务器，暴露 REST 接口。
    - 负责参数校验、路由、业务逻辑编排。
  - **存储层（Redis）**
    - 负责保存短链记录、访问统计等（Hash + TTL）。

- **调用流程**
  - **创建短链**
    1. 客户端调用 `POST /api/v1/shorten`。
    2. API 校验请求参数。
    3. 生成短码（见后文算法）。
    4. 写入 Redis。
    5. 返回短码和完整短链接。
  - **访问短链**
    1. 用户访问 `GET /{code}`。
    2. API 从 Redis 查出原始 URL。
    3. 校验是否过期，如已过期返回 404/410。
    4. 记录访问次数。
    5. 返回 302 重定向到原始 URL。

---

## 六、模块设计

- **HTTP 层（handler）**
  - `ShortenHandler`：处理创建短链请求。
  - `RedirectHandler`：处理短码重定向。
  - `InfoHandler`：查询短链信息。

- **Service 层**
  - `LinkService`
    - `CreateShortLink(longURL, customCode, expireAt)`：业务校验 + 生成短码 + 调用存储。
    - `GetLongURL(code)`：根据短码取回 URL，校验过期，更新计数。
    - `GetLinkInfo(code)`：查询元数据。

- **Repository 层**
  - `LinkRepository`（面向接口编程）
    - `Save(link)`、`FindByCode(code)`、`IncrementClick(code)` 等。

- **Utility & Infra**
  - 短码生成器（`CodeGenerator`）：基于自增 ID + Base62 编码。
  - 时间工具：统一时区与时间格式。
  - 错误与响应封装：统一 JSON 错误结构。

---

## 七、数据库设计（Redis）

- **Key / Hash 设计**
  - **自增 ID**：`shortener:next_id`（string，使用 `INCR`）
  - **短链记录**：`shortener:link:{code}`（hash）
    - `id`：自增 ID（int）
    - `code`：短码（string）
    - `long_url`：原始链接（string）
    - `created_at`：创建时间（RFC3339Nano 字符串）
    - `expire_at`：过期时间（RFC3339Nano 字符串，可空）
    - `click_count`：访问次数（int）
    - `last_accessed_at`：最后访问时间（RFC3339Nano 字符串，可空）
  - **过期策略**：如设置 `expire_at`，则对 `shortener:link:{code}` 设置 TTL（到期自动删除）。

- **短码生成策略**

  - 使用 Redis 的 `INCR shortener:next_id` 获取自增 ID，再通过 Base62 编码得到短码。
  - Base62 字符集：`0-9A-Za-z`。
  - 避免碰撞：自增 ID 天然唯一；如允许自定义短码，则写入前检查 `shortener:link:{code}` 是否存在。

---

## 八、接口设计

### 1. 创建短链接

- **URL**：`POST /api/v1/shorten`
- **请求体（JSON）**：

```json
{
  "url": "https://example.com/very/long/url",
  "custom_code": "myalias",
  "expire_at": "2026-12-31T23:59:59Z"
}
```

- **成功响应 200**：

```json
{
  "code": "abc123",
  "short_url": "https://short.domain/abc123",
  "long_url": "https://example.com/very/long/url",
  "expire_at": "2026-12-31T23:59:59Z"
}
```

- **可能错误**
  - `400`：URL 非法 / 自定义短码不合法或已被占用。
  - `422`：参数格式错误。
  - `500`：内部错误（数据库写入失败等）。

### 2. 短链重定向

- **URL**：`GET /{code}`
- **行为**：
  - 若存在且未过期：返回 `302 Found`，`Location` 头为原始 URL。
  - 若不存在或已过期：返回 `404 Not Found`（可返回简单 JSON 或 HTML）。

- **错误响应（示例 JSON）**：

```json
{
  "error": "not_found",
  "message": "short link not found or expired"
}
```

### 3. 查询短链信息（可选管理接口）

- **URL**：`GET /api/v1/links/{code}`
- **成功响应 200**：

```json
{
  "code": "abc123",
  "long_url": "https://example.com/very/long/url",
  "created_at": "2026-01-19T10:00:00Z",
  "expire_at": "2026-12-31T23:59:59Z",
  "click_count": 123,
  "last_accessed_at": "2026-01-19T11:00:00Z"
}
```

---

## 九、错误码与响应规范

- **统一 JSON 错误结构**：

```json
{
  "error": "invalid_request",
  "message": "url is required"
}
```

- 常见错误类型：
  - **`invalid_request`**：参数缺失或格式错误。
  - **`conflict`**：自定义短码已被占用。
  - **`not_found`**：短码不存在或已过期。
  - **`internal_error`**：系统内部错误。

---

## 十、日志与监控设计

- **日志内容**
  - 请求基本信息：方法、路径、状态码、耗时、客户端 IP。
  - 错误栈：使用结构化日志（如 `logrus` 或 `zap`）。
  - 数据库操作错误详细记录。

- **监控接口**
  - `GET /healthz`：返回服务健康状态（示例：`{ "status": "ok" }`）。
  - 预留 Prometheus metrics 端点（后期可扩展）。

---

## 十一、Docker 与部署设计

- **目录结构（整体）**

```text
backend/
  ├── cmd/server/main.go
  ├── internal/...
  ├── go.mod / go.sum
  ├── Dockerfile          <-- 使用 golang:1.23 多阶段构建
frontend/
  ├── src/...             <-- React + Vite 前端代码
  ├── Dockerfile          <-- 使用 nginx:1.28-alpine 作为运行时
docker-compose.yml        <-- 编排 backend、frontend、redis 三个服务
```

-- **容器设计**
  - **后端服务容器：`backend`**
    - 多阶段构建，使用 `golang:1.23` 作为构建镜像，运行阶段使用精简 `alpine`。
    - 暴露端口 `8080`（或通过环境变量 `PORT` 配置）。
    - 通过环境变量 `REDIS_ADDR/REDIS_PASSWORD/REDIS_DB` 连接 Redis，`BASE_URL` 用于拼接短链接域名。
  - **数据库容器：`redis`**
    - 使用 `redis:7.4` 镜像。
    - 通过 `redis-server --appendonly yes` 开启 AOF 持久化。
    - 使用 `redis-data` volume 持久化 `/data`。
  - **前端容器：`frontend`**
    - 使用 `node:20-alpine` 构建前端资源，运行阶段使用 `nginx:1.28-alpine`。
    - 将构建产物挂载到 Nginx 默认静态目录，对外暴露 `3000:80` 端口。

- **`docker-compose.yml` 关键点**
  - 服务 `backend`：
    - `build: ./backend`
    - `ports: ["8080:8080"]`
    - `environment`：`PORT`, `REDIS_ADDR`, `REDIS_PASSWORD`, `REDIS_DB`, `BASE_URL` 等。
    - `depends_on`: `redis`
  - 服务 `redis`：
    - `image: redis:7.4`
    - `command`: `redis-server --appendonly yes`
    - `volumes`：`redis-data:/data`。
  - 服务 `frontend`：
    - `build: ./frontend`
    - `ports: ["3000:80"]`
    - `depends_on`: `backend`

---

## 十二、后续扩展方向

- **增加认证与用户系统**：实现“每个用户自己的短链空间”，支持登录与权限隔离。
- **批量生成短链**：支持一次请求中批量提交多个 URL。
- **自定义域名支持**：不同业务线可使用不同短域。
- **高可用与水平扩展**
  - 增加 Redis 缓存层，部署多实例并通过 Nginx / API Gateway 负载均衡。
- **统计报表**：提供按时间维度的访问统计接口和后台管理页面。

