# Personal Backend Hub

基于 `go-kratos` 的个人 API 中台，统一承载个人后端小工具接口。集成了短链接服务、开发者工具箱、AI 智能助手等功能。

## 功能列表

### ✅ 已实现能力

| 模块 | 功能 | 状态 |
|------|------|------|
| 🔗 短链接 | 创建、跳转、详情查询 | ✅ |
| 🛠️ 工具箱 | Ping、时间戳、Token、摘要(Digest) | ✅ |
| 🤖 AI 助手 | 智能聊天、翻译、总结、润色、语法修正 | ✅ |
| 🌐 IP 查询 | 获取 IP 归属地信息 | ✅ |
| 🖼️ 图床上传 | 图片上传存储 | ✅ |
| 🔐 认证系统 | JWT 登录、注册、用户信息 | ✅ |
| 📝 笔记/收藏夹 | 笔记和书签管理 | ✅ |
| ⏰ 定时提醒 | 提醒事项管理 | ✅ |
| 🚀 GitHub Webhook | 接收 GitHub 事件 | ✅ |

## 技术栈

- **框架**: go-kratos v2
- **语言**: Go 1.25
- **数据库**: PostgreSQL 16
- **容器**: Docker Compose
- **AI 接口**: OpenAI / DeepSeek API (可选)

## 快速开始

### 1. 启动 PostgreSQL

```bash
docker compose up -d postgres

默认数据库配置：

配置项	值
Host	127.0.0.1
Port	5432
DB	personal_backend_hub
User	postgres
Password	postgres
2. 配置环境变量（可选）
如需启用 AI 功能，设置 API Key：

bash
# Windows PowerShell
$env:OPENAI_API_KEY="your-openai-key"
$env:DEEPSEEK_API_KEY="your-deepseek-key"

# Linux/Mac
export OPENAI_API_KEY="your-openai-key"
export DEEPSEEK_API_KEY="your-deepseek-key"
3. 启动服务
bash
go run ./cmd/ownhub -conf ./configs/config.yaml
服务启动后：

HTTP: http://127.0.0.1:8000

gRPC: 127.0.0.1:9000

前端页面: http://127.0.0.1:8000

接口文档
🔐 认证
方法	路径	描述
POST	/api/v1/auth/register	用户注册
POST	/api/v1/auth/login	用户登录
GET	/api/v1/auth/me	获取当前用户信息
🛠️ 工具箱
方法	路径	描述
GET	/api/v1/tools/ping	健康检查
GET	/api/v1/tools/timestamp	获取时间戳（支持时区）
POST	/api/v1/tools/token	生成随机 Token
POST	/api/v1/tools/digest	文本摘要（MD5/SHA1/SHA256/SHA512）
🌐 IP 查询
方法	路径	描述
GET	/api/v1/toolbox/ip?ip=8.8.8.8	查询 IP 归属地
🔗 短链接
方法	路径	描述
POST	/api/v1/short-links	创建短链接
GET	/api/v1/short-links/{code}	跳转原始 URL
GET	/api/v1/short-links/{code}/detail	查询短链接详情
GET	/s/{code}	快捷跳转
🖼️ 图床上传
方法	路径	描述
POST	/api/v1/uploads/image	上传图片（需登录）
请求头需要 Authorization: Bearer <token>，表单字段：file

🤖 AI 助手
方法	路径	描述
POST	/api/v1/tools/ai/chat	AI 智能对话（支持上下文）
POST	/api/v1/tools/ai/process	文本处理（翻译/总结/润色/语法修正）
AI 处理示例：

bash
# 翻译
curl -X POST /api/v1/tools/ai/process \
  -H "Content-Type: application/json" \
  -d '{"text":"Hello world","action":"translate","target_lang":"中文"}'

# 总结
curl -X POST /api/v1/tools/ai/process \
  -H "Content-Type: application/json" \
  -d '{"text":"长文本内容...","action":"summarize"}'
📝 笔记/收藏夹
方法	路径	描述
POST	/api/v1/notes	创建笔记/收藏
GET	/api/v1/notes	获取列表
GET	/api/v1/notes/{id}	获取详情
kind 字段支持 note（笔记）和 bookmark（书签）

⏰ 提醒
方法	路径	描述
POST	/api/v1/reminders	创建提醒
GET	/api/v1/reminders	获取提醒列表
POST	/api/v1/reminders/done?id=1	标记完成
🚀 GitHub Webhook
方法	路径	描述
POST	/api/v1/webhooks/github	接收 GitHub 事件
配置说明
配置文件：configs/config.yaml

配置项	说明
app.public_base_url	短链和图床 URL 基础地址
app.jwt_secret	JWT 签名密钥
app.upload_dir	上传文件存储目录
app.github_webhook_secret	GitHub Webhook 签名密钥
数据表
服务启动时会自动创建以下数据表：

users - 用户表

short_links - 短链接表

notes - 笔记/收藏夹表

reminders - 提醒表

uploads - 上传文件记录表

项目结构
text
personal-backend-hub/
├── api/              # Proto 定义
├── cmd/ownhub/       # 程序入口
├── internal/
│   ├── biz/          # 业务逻辑层
│   ├── data/         # 数据访问层
│   ├── server/       # HTTP/gRPC 服务器
│   └── service/      # 服务层
├── web/              # 静态前端
├── storage/          # 上传文件存储
├── configs/          # 配置文件
└── docker-compose.yml
开发说明
数据表会在服务启动时自动创建（GORM AutoMigrate）

图床文件默认保存在 ./storage/uploads

IP 查询通过 ip-api.com 获取信息

AI 功能支持模拟模式（未配置 API Key 时自动降级）
