# Personal Backend Hub

基于 `go-kratos` 的个人 API 中台，用来统一承载“我自己的后端小工具接口”。

## 已实现能力

- IP 信息查询
- 短链生成与跳转
- 文件上传图床
- JWT 登录与自有账号体系
- 简单笔记 / 收藏夹 API
- 定时提醒 API
- GitHub Webhook 接收

## 技术栈

- Go 1.25
- go-kratos v2
- PostgreSQL 16
- Docker Compose

## 启动 PostgreSQL

```bash
docker compose up -d postgres
```

默认数据库配置：

- Host: `127.0.0.1`
- Port: `5432`
- DB: `personal_backend_hub`
- User: `postgres`
- Password: `postgres`

## 启动服务

```bash
go run ./cmd/ownhub -conf ./configs/config.yaml
```

服务启动后：

- HTTP: `http://127.0.0.1:8000`
- gRPC: `127.0.0.1:9000`

## 主要接口

### 认证

- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `GET /api/v1/auth/me`

### 工具箱

- `GET /api/v1/toolbox/ip?ip=8.8.8.8`
- `GET /api/v1/tools/ping`
- `GET /api/v1/tools/timestamp`
- `POST /api/v1/tools/token`
- `POST /api/v1/tools/digest`

### 短链

- `POST /api/v1/short-links`
- `GET /api/v1/short-links/{code}`
- `GET /s/{code}`

### 图床上传

- `POST /api/v1/uploads/image`
- 需要 `Authorization: Bearer <token>`
- 表单字段：`file`

### 笔记 / 收藏夹

- `POST /api/v1/notes`
- `GET /api/v1/notes`
- `kind` 支持 `note` / `bookmark`

### 提醒

- `POST /api/v1/reminders`
- `GET /api/v1/reminders`
- `POST /api/v1/reminders/done?id=1`

### GitHub Webhook

- `POST /api/v1/webhooks/github`

## 配置项

见 [configs/config.yaml](C:/Users/DELL/Documents/Codex/2026-05-13/api-personal-backend-hub/configs/config.yaml)

重点字段：

- `app.public_base_url`: 生成短链和图床 URL 的基础地址
- `app.jwt_secret`: JWT 签名密钥
- `app.upload_dir`: 上传文件存储目录
- `app.github_webhook_secret`: GitHub Webhook 签名密钥，可为空

## 说明

- 数据表会在服务启动时自动创建。
- 图床文件默认落到 `./storage/uploads`。
- IP 查询当前通过 `ip-api.com` 拉取信息。
