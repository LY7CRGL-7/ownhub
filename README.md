# Personal Backend Hub

## 简介
基于 go-kratos 的个人 API 中台，用于统一承载各类后端工具接口。

---

## 功能

[短链接]
- 创建 / 跳转 / 详情

[工具箱]
- Ping / 时间戳 / Token / 摘要

[AI 助手]
- 聊天 / 翻译 / 总结 / 润色

[其他]
- IP 查询
- 图床上传
- JWT 认证
- 笔记 / 收藏夹
- 定时提醒
- GitHub Webhook

---

## 技术栈

- Go 1.25
- go-kratos v2
- PostgreSQL 16
- Docker Compose
- OpenAI / DeepSeek（可选）

---

## 快速开始

[1] 启动数据库

docker compose up -d postgres

默认配置：
Host:     127.0.0.1
Port:     5432
DB:       personal_backend_hub
User:     postgres
Password: postgres

---

[2] 配置 AI Key（可选）

Windows:
$env:OPENAI_API_KEY="your-key"
$env:DEEPSEEK_API_KEY="your-key"

Linux / Mac:
export OPENAI_API_KEY="your-key"
export DEEPSEEK_API_KEY="your-key"

---

[3] 启动服务

go run ./cmd/ownhub -conf ./configs/config.yaml

服务地址：
HTTP: http://127.0.0.1:8000
gRPC: 127.0.0.1:9000

---

## 接口

[认证]
POST /api/v1/auth/register
POST /api/v1/auth/login
GET  /api/v1/auth/me

[工具箱]
GET  /api/v1/tools/ping
GET  /api/v1/tools/timestamp
POST /api/v1/tools/token
POST /api/v1/tools/digest

[IP 查询]
GET /api/v1/toolbox/ip?ip=8.8.8.8

[短链接]
POST /api/v1/short-links
GET  /api/v1/short-links/{code}
GET  /api/v1/short-links/{code}/detail
GET  /s/{code}

[图床上传]
POST /api/v1/uploads/image
Header: Authorization: Bearer <token>
Form: file

[AI 助手]
POST /api/v1/tools/ai/process

示例：
{
"text": "Hello world",
"action": "translate",
"target_lang": "中文"
}

[笔记 / 收藏夹]
POST /api/v1/notes
GET  /api/v1/notes
GET  /api/v1/notes/{id}
kind: note / bookmark

[提醒]
POST /api/v1/reminders
GET  /api/v1/reminders
POST /api/v1/reminders/done?id=1

[Webhook]
POST /api/v1/webhooks/github

---

## 配置

文件：configs/config.yaml

app.public_base_url         短链 / 图床基础 URL
app.jwt_secret              JWT 密钥
app.upload_dir              上传目录
app.github_webhook_secret   Webhook 密钥

---

## 数据表

users
short_links
notes
reminders
uploads

---

## 项目结构

personal-backend-hub/
├── api/
├── cmd/ownhub/
├── internal/
│   ├── biz/
│   ├── data/
│   ├── server/
│   └── service/
├── web/
├── storage/
├── configs/
└── docker-compose.yml

---

## 说明

- 启动自动建表（GORM AutoMigrate）
- 上传文件：./storage/uploads
- IP 查询：ip-api.com
- AI 未配置 Key 自动降级