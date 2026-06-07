# SpeakMate AI

SpeakMate AI 是一个场景化 AI 英语口语陪练系统。它围绕面试、点餐、会议等真实场景，引导用户完成英语对话练习，并在练习后提供表达纠错、能力评分和复盘建议。

它不是开放式闲聊机器人，而是一个带训练目标的英语陪练教练：用户先选择具体场景，AI 再按场景角色追问，最后把表达问题整理成可执行的练习建议。

demo视频：https://www.bilibili.com/video/BV1orE86hEJd/?vd_source=e0d0a3eee43f72521fe869af240d5b0e

## 功能概览

- 场景化训练：内置英语面试、餐厅点餐、工作会议等训练场景。
- AI 主动追问：根据场景目标推动完整对话，而不是只回答单句问题。
- 表达纠错：分析语法、用词和表达自然度，并给出更自然的英文表达。
- 多维评分：从流利度、语法准确度、表达自然度、词汇丰富度和场景完成度评估练习结果。
- 课后报告：汇总训练过程、常见错误、推荐表达和下次练习建议。
- 语音入口：支持浏览器录音、WebSocket 音频分片和单段音频上传 fallback。
- 历史记录：支持查询训练历史、报告和训练详情。

## 当前状态

| 能力 | 状态 |
|---|---|
| Go + Gin 后端 API | 已支持 |
| Vite + React + TypeScript 前端 | 已支持 |
| 文本训练闭环 | 已支持 |
| Mock Agent / Mock ASR | 已支持 |
| OpenAI-compatible LLM streaming | 已支持 |
| 腾讯云 ASR FlashRecognizer | 已支持 |
| MySQL 持久化 | 已支持 |
| Redis 短期状态与 SSE 事件总线 | 已支持 |
| 真实实时 partial transcript | 计划中 |
| 后端音频转码兼容 `webm/opus` | 计划中 |

## 技术栈

| 分类 | 技术 |
|---|---|
| 后端 | Go、Gin |
| AI 能力 | Prompt、Agent 接口、Mock/LLM 实现、OpenAI-compatible streaming |
| 语音识别 | Mock ASR、腾讯云 ASR FlashRecognizer |
| 实时通信 | SSE、WebSocket |
| 数据存储 | MySQL、Redis、内存存储 |
| 前端 | Vite、React、TypeScript |
| 部署 | Docker、Docker Compose、Nginx |

## 快速开始

默认配置使用 Mock Agent、Mock ASR 和内存存储，不需要 API Key、MySQL 或 Redis。

### 前置要求

- Go 1.26 或更高版本
- Node.js 与 npm
- Docker 与 Docker Compose，可选

### 启动后端

macOS / Linux：

```bash
STORAGE_MODE=memory LLM_USE_MOCK=true ASR_USE_MOCK=true go run ./cmd/server
```

PowerShell：

```powershell
$env:STORAGE_MODE="memory"
$env:LLM_USE_MOCK="true"
$env:ASR_USE_MOCK="true"
go run ./cmd/server
```

检查服务状态：

```bash
curl http://localhost:8080/health
```

预期响应：

```json
{"code":0,"message":"success","data":{"status":"ok"}}
```

### 启动前端

```bash
cd web
npm install
npm run dev
```

本地前后端分开启动时，建议显式配置后端 API 地址。

PowerShell：

```powershell
$env:VITE_API_BASE_URL="http://localhost:8080/api/v1"
npm run dev
```

macOS / Linux：

```bash
VITE_API_BASE_URL=http://localhost:8080/api/v1 npm run dev
```

前端默认访问地址是 `http://localhost:5173`。

## Docker Compose

一条命令拉起 MySQL、Redis、迁移任务、后端和前端静态服务：

```bash
docker compose up --build
```

启动后访问：

```text
前端：http://localhost:5173
后端健康检查：http://localhost:8080/health
MySQL：127.0.0.1:3306
Redis：127.0.0.1:6379
```

Compose 默认使用 MySQL 持久化、Redis 短期状态和 Mock Agent。迁移任务会在后端启动前执行。

如果本机端口已被占用，可以通过环境变量避让宿主端口：

```powershell
$env:COMPOSE_MYSQL_PORT="33306"
$env:COMPOSE_REDIS_PORT="36379"
$env:BACKEND_HOST_PORT="18080"
$env:FRONTEND_HOST_PORT="15173"
docker compose up --build
```

## 配置

示例环境变量见 [.env.example](.env.example)。Go 服务不会自动读取 `.env` 文件，可以用 shell、direnv、dotenv 工具或 Docker Compose 注入环境变量。

常用配置：

| 用途 | 变量 |
|---|---|
| 服务 | `APP_PORT`、`REQUEST_TIMEOUT_SECONDS`、`REQUEST_BODY_LIMIT_BYTES` |
| 跨域 | `CORS_ALLOWED_ORIGINS`、`CORS_ALLOWED_METHODS`、`CORS_ALLOWED_HEADERS` |
| 存储 | `STORAGE_MODE`、`MYSQL_DSN` |
| Redis | `REDIS_ENABLED`、`REDIS_ADDR`、`REDIS_PASSWORD`、`REDIS_DB` |
| LLM | `LLM_USE_MOCK`、`LLM_PROVIDER`、`LLM_BASE_URL`、`LLM_API_KEY`、`LLM_MODEL` |
| ASR | `ASR_USE_MOCK`、`ASR_PROVIDER`、`TENCENT_ASR_APP_ID`、`TENCENT_ASR_SECRET_ID`、`TENCENT_ASR_SECRET_KEY` |
| 前端 | `VITE_API_BASE_URL` |

真实密钥只应放在本地环境变量或部署密钥系统中，不要提交到 git。

## 数据库迁移

内存模式适合本地开发和测试，服务重启后数据会丢失。切换 MySQL 前先创建数据库并执行迁移：

```bash
docker compose up -d mysql redis
```

PowerShell：

```powershell
$env:MYSQL_DSN="speakmate:speakmate@tcp(127.0.0.1:3306)/speakmate?parseTime=true&loc=UTC"
go run ./cmd/migrate -dir migrations
$env:STORAGE_MODE="mysql"
go run ./cmd/server
```

macOS / Linux：

```bash
export MYSQL_DSN='speakmate:speakmate@tcp(127.0.0.1:3306)/speakmate?parseTime=true&loc=UTC'
go run ./cmd/migrate -dir migrations
STORAGE_MODE=mysql go run ./cmd/server
```

如果安装了 `make`，也可以执行：

```bash
make migrate
```

## 验证

后端：

```bash
go test ./...
```

前端：

```bash
cd web
npm test
npm run build
```

Mock 模式 API 闭环脚本：

```powershell
powershell -ExecutionPolicy Bypass -File scripts/demo-mock.ps1
```

真实 LLM / ASR / Redis 联调脚本：

```powershell
powershell -ExecutionPolicy Bypass -File scripts/demo-real-services.ps1
```

## 项目结构

```text
speakmate/
├── cmd/
│   ├── server/              # 后端服务入口
│   └── migrate/             # MySQL migration 执行入口
├── internal/
│   ├── agent/               # Conversation/Feedback/Summary/ASR Agent 与 Mock/LLM 实现
│   ├── config/              # 环境配置
│   ├── handler/             # HTTP Handler
│   ├── infra/               # MySQL、Redis、LLM、ASR 基础设施适配
│   ├── middleware/          # CORS、日志、recover、timeout、限流
│   ├── repository/          # memory/mysql 仓库实现
│   ├── response/            # 统一响应结构
│   ├── router/              # Gin 路由
│   ├── service/             # 业务服务
│   ├── state/               # Session 短期状态 store
│   └── stream/              # SSE 事件模型和事件总线
├── migrations/              # MySQL 表结构和默认场景 seed
├── web/                     # Vite + React + TypeScript 前端应用
│   ├── src/                 # 页面、组件、API client 和类型定义
│   ├── Dockerfile           # 前端构建 + Nginx 静态部署
│   └── nginx.conf           # 静态资源、API、SSE、WebSocket 反向代理
├── scripts/                 # Demo 和联调脚本
├── docker-compose.yml
├── Dockerfile
├── go.mod
└── README.md
```

## 开发与贡献

提交变更前建议至少执行：

```bash
go test ./...
cd web
npm test
npm run build
```

后端质量门禁也会检查 `go mod tidy`、`gofmt`、`go vet`、race test 和构建。提交规范可参考 [PR提交规范.md](PR提交规范.md)。

## 许可证

当前仓库尚未声明开源许可证。公开发布前请补充根目录 `LICENSE` 文件，并在本节写明许可证类型。
