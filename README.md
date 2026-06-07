# SpeakMate AI

> 场景化 AI 英语口语陪练系统，围绕面试、点餐、会议等真实场景，帮助用户完成英语对话练习、表达纠错、能力评分和课后复盘。

SpeakMate AI 不是一个开放闲聊机器人。它更像一个有训练目标的英语陪练教练：先给用户一个具体场景，再通过 AI 追问推动对话，最后把用户的表达问题整理成可执行的练习建议。

本项目面向七牛云 XEngineer 暑期实训营「AI 英语口语陪练」议题设计。当前仓库已完成 Go + Gin 后端基础骨架、正式 Vite + React + TypeScript 前端应用、文本训练闭环、单段录音上传、实时音频 WebSocket 分片入口和可配置腾讯云 ASR Provider。完整技术方案见 [docs/project-blueprint.md](docs/project-blueprint.md)。

## 项目背景

很多英语学习者并不是不知道单词或语法，而是缺少真实表达环境：

- 不知道在面试、点餐、会议等具体场景里该怎么开口；
- 和真人练习成本高，时间不灵活，也容易有压力；
- 普通 AI 聊天可以对话，但通常缺少训练目标和系统反馈；
- 练完之后很难知道自己哪里说错了、哪里可以说得更自然。

SpeakMate AI 试图把一次口语练习拆成清晰的训练闭环：

```text
选择场景 -> 开始对话 -> AI 追问 -> 表达纠错 -> 多维评分 -> 课后报告 -> 下次练习建议
```

## 核心体验

### 1. 场景化训练

用户先选择一个真实口语场景，每个场景都有明确的 AI 角色、训练目标和评分重点。

| 场景 | AI 角色 | 训练目标 |
|---|---|---|
| 英语面试 | 技术面试官 | 练习自我介绍、项目经历、技术追问和结尾提问 |
| 餐厅点餐 | 餐厅服务员 | 练习询问菜单、表达偏好、处理特殊需求 |
| 工作会议 | 项目同事 | 练习表达观点、补充说明、澄清问题和总结结论 |

### 2. AI 主动追问

AI 不只是被动回答用户，而是根据当前场景推动对话。例如英语面试场景会逐步引导用户完成：

```text
自我介绍 -> 项目经历 -> 技术细节 -> 团队协作 -> 反问面试官
```

这样用户练到的不是孤立句子，而是一段完整的真实任务表达。

### 3. 低打断式纠错

口语训练最怕刚开口就被打断。SpeakMate 的纠错策略是：

- 对话过程中只展示轻量提示，不频繁打断表达；
- 单轮结束后异步分析语法、用词和表达自然度；
- 训练结束后集中生成完整课后报告。

示例：

```text
原句：I am study computer science and I have did a project about robot.

建议：I am studying computer science, and I have done a project on robot control.

问题：
- "am study" 应改为 "am studying"
- "have did" 应改为 "have done"
- "about robot" 可优化为 "on robot control"
```

### 4. 多维度评分

系统计划从五个维度评价一次训练结果：

| 维度 | 说明 |
|---|---|
| 流利度 | 回答是否连贯，是否有明显停顿、重复和填充词 |
| 语法准确度 | 时态、主谓一致、句法结构、介词搭配是否准确 |
| 表达自然度 | 是否符合英语常用表达习惯 |
| 词汇丰富度 | 是否使用了和场景匹配的词汇与短语 |
| 场景完成度 | 是否完成当前场景的核心任务 |

### 5. 课后报告

每次训练结束后，系统会生成一份结构化报告，帮助用户知道下一步该练什么：

- 本次训练概览；
- 综合评分和分项评分；
- 高频错误；
- 更自然的推荐表达；
- 场景完成情况；
- 下次练习建议。

## 系统设计

后端以 Go + Gin 为基础，后续围绕训练 Session、AI Agent、ASR、纠错、评分和报告生成逐步扩展。

```text
Browser
  -> Gin API
  -> Session Service
  -> Eino Agent Workflow
       -> Conversation Agent
       -> Correction Agent
       -> Scoring Agent
       -> Summary Agent
  -> Redis / MySQL
```

规划中的核心模块：

| 模块 | 职责 |
|---|---|
| Scenario Service | 管理训练场景、AI 角色和场景目标 |
| Session Service | 管理一次口语训练的生命周期 |
| Conversation Agent | 根据场景生成 AI 回复和追问 |
| Correction Agent | 分析语法、用词和表达问题 |
| Scoring Agent | 生成分项评分和评分理由 |
| Summary Agent | 汇总训练过程并生成课后报告 |
| ASR Service | 将用户语音转换为英文文本 |
| Report Service | 保存并查询训练报告 |

## 当前进度

| 模块 | 状态 | 说明 |
|---|---|---|
| Gin 服务骨架 | 已完成 | `cmd/server` 启动 HTTP 服务 |
| 健康检查接口 | 已完成 | `GET /health` 返回统一 JSON 响应 |
| 配置加载 | 已完成 | 支持 `APP_PORT`、LLM、ASR、Redis、反馈 Mock 和存储模式环境变量，默认端口 `8080` |
| 统一响应结构 | 已完成 | 成功响应格式为 `{ code, message, data }` |
| 前端原型 | 已完成 | `web/preview.html` 作为本地静态原型参考 |
| 正式前端应用 | 已完成（第一版） | `web/` 已接入 Vite + React + TypeScript，覆盖场景选择、文本训练、反馈评分、报告和历史记录 |
| 场景 API（Module 1） | 已完成 | 场景列表和详情接口，见 [docs/api文档/scenario-api-module1-api.md](docs/api文档/scenario-api-module1-api.md) |
| 训练 Session API（Module 2） | 已完成 | 创建、查询、结束训练 Session，见 [docs/api文档/session-api-module2-api.md](docs/api文档/session-api-module2-api.md) |
| 消息 API（Module 3） | 已完成 | 发送文本、AI 回复、轮次更新和消息历史查询，见 [docs/api文档/message-api-module3-api.md](docs/api文档/message-api-module3-api.md) |
| 反馈 API（Module 4） | 已完成（第一版） | 消息发送后同步生成纠错/评分摘要，支持单条消息纠错、Session 纠错列表和当前评分查询，见 [docs/api文档/feedback-api.md](docs/api文档/feedback-api.md) |
| 课后报告 API（Module 5） | 已完成（第一版） | 训练结束后可基于消息、纠错和评分生成结构化报告，支持重复查询，见 [docs/api文档/report-api.md](docs/api文档/report-api.md) |
| MySQL 持久化与历史记录 | 已完成（第一版） | 支持 `memory` / `mysql` 存储模式切换，Session、Message、Correction、Score、Report 可落库，历史列表见 [docs/api文档/history-api.md](docs/api文档/history-api.md) |
| SSE 流式事件 | 已完成 | 支持 `GET /api/v1/sessions/:id/stream`，真实 LLM 模式推送模型 delta，Mock/fallback 模式推送本地 fake delta，并继续推送纠错、评分、报告和错误事件；`REDIS_ENABLED=true` 时使用 Redis Pub/Sub + 短期事件 List，见 [docs/api文档/sse-api.md](docs/api文档/sse-api.md) |
| Conversation Agent | 已接入 | 默认使用 Mock fake streaming；配置 API Key 且关闭 Mock 后使用 OpenAI-compatible LLM streaming，按 `LLM_FALLBACK_TO_MOCK` 决定失败时是否降级 Mock |
| AI 纠错、评分与总结 | 已完成（第一版） | Correction / Scoring / Summary 模型、Mock/LLM Agent、内存 Feedback/Report Repository、fail-open 降级和查询 API 已接入 |
| 语音能力 | 已完成（Mock / 腾讯云 ASR） | 训练页支持浏览器录音、WebSocket 分片和单段上传 fallback；后端可使用 Mock ASR 或腾讯云 `FlashRecognizer` 生成 final transcript 后复用消息训练链路；WebSocket 连接状态可写入 Redis 并自动过期，见 [docs/api文档/audio-api.md](docs/api文档/audio-api.md) 和 [docs/api文档/audio-websocket-api.md](docs/api文档/audio-websocket-api.md) |
| Redis 会话与事件状态 | 已完成 | `REDIS_ENABLED=false` 使用 memory state store；`REDIS_ENABLED=true` 时训练上下文快照、当前阶段/轮次、临时评分、纠错摘要、SSE 事件和 WebSocket 连接状态写入 Redis，所有 key 均设置 TTL |

## 技术栈

| 分类 | 技术 |
|---|---|
| 后端 | Go、Gin |
| AI 编排 | Eino |
| 实时通信 | SSE、WebSocket |
| 数据存储 | MySQL、Redis |
| 前端 | Vite、React、TypeScript |
| 当前原型 | HTML、CSS、JavaScript |

## 快速开始

默认配置使用 Mock Agent 和内存存储，不需要 API Key、MySQL 或 Redis：

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

示例环境变量见 [.env.example](.env.example)。Go 服务不会自动读取 `.env` 文件，可以用 shell、direnv、dotenv 工具或 Docker Compose 注入环境变量。

### 配置说明

关键配置按用途拆分：

| 用途 | 变量 |
|---|---|
| 服务 | `APP_PORT`、`REQUEST_TIMEOUT_SECONDS` |
| 跨域 | `CORS_ALLOWED_ORIGINS`、`CORS_ALLOWED_METHODS`、`CORS_ALLOWED_HEADERS`、`CORS_ALLOW_CREDENTIALS` |
| MySQL | `STORAGE_MODE=mysql`、`MYSQL_DSN` |
| Redis 短期状态 | `REDIS_ENABLED`、`REDIS_ADDR`、`REDIS_PASSWORD`、`REDIS_DB`、`REDIS_CONNECT_TIMEOUT_SECONDS` |
| 外部服务超时 | `EXTERNAL_SERVICE_TIMEOUT_SECONDS`，可被 `LLM_TIMEOUT_SECONDS` / `ASR_TIMEOUT_SECONDS` 覆盖 |
| Mock / fallback | `LLM_USE_MOCK`、`LLM_FALLBACK_TO_MOCK`、`CORRECTION_USE_MOCK`、`SCORING_USE_MOCK`、`SUMMARY_USE_MOCK`、`ASR_USE_MOCK` |
| 腾讯云 ASR | `ASR_PROVIDER=tencent`、`TENCENT_ASR_APP_ID`、`TENCENT_ASR_SECRET_ID`、`TENCENT_ASR_SECRET_KEY`、`TENCENT_ASR_ENGINE_TYPE` |

本分支支持真实 LLM streaming、腾讯云 ASR 文件识别极速版，以及可配置 Redis 短期状态与事件总线。Redis 只管理训练过程中的短期状态，不替代 MySQL 长期仓库。

### Redis 会话状态与事件

默认本地开发和自动测试使用 memory 模式：

```bash
REDIS_ENABLED=false
```

开启 Redis 模式：

```bash
REDIS_ENABLED=true
REDIS_ADDR=127.0.0.1:6379
REDIS_PASSWORD=
REDIS_DB=0
REDIS_CONNECT_TIMEOUT_SECONDS=30
```

Redis 模式下，服务启动时会 ping Redis；连接失败或 `REDIS_ADDR` 缺失会直接启动失败，不会静默降级到 memory。运行过程中 Redis 状态写入或事件发布失败会返回明确错误，例如 `session state store failed` 或 `stream event publish failed`。Compose 默认让后端依赖健康的 Redis 服务，本地单独验证可先执行：

```bash
docker compose up -d redis
```

短期 key 设计：

| Key | 类型 | TTL | 说明 |
|---|---|---:|---|
| `session:{id}:messages` | List | 2h | 当前训练上下文快照 |
| `session:{id}:state` | Hash | 2h | 当前阶段、轮次、状态 |
| `session:{id}:partial_score` | Hash | 2h | 当前临时分项评分 |
| `session:{id}:corrections` | List | 2h | 临时纠错摘要 |
| `session:{id}:events` | Pub/Sub channel + List | 30m | SSE / WebSocket 事件分发与短期留存 |
| `ws:{session_id}:connection` | Hash | 30m | WebSocket 连接状态 |

MySQL 仍负责 Session、Message、Correction、Score、Report 等长期数据。Redis 中的状态可以过期或清空，不应作为历史记录来源。

### 前端启动

首次启动前安装依赖：

```bash
cd web
npm install
```

本地前后端分开启动时，建议显式配置后端 API 地址：

```powershell
$env:VITE_API_BASE_URL="http://localhost:8080/api/v1"
npm run dev
```

macOS / Linux：

```bash
VITE_API_BASE_URL=http://localhost:8080/api/v1 npm run dev
```

前端默认端口是 `5173`。同源部署时可以使用 `VITE_API_BASE_URL=/api/v1`，前端 Nginx 会把 `/api/`、SSE 和 WebSocket 请求代理到后端。

### MySQL 与 Migration

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

迁移文件按文件名顺序执行。当前 SQL 使用 `CREATE TABLE IF NOT EXISTS` 和 `ON DUPLICATE KEY UPDATE`，可重复执行；失败时会打印具体 migration 文件和语句序号。`STORAGE_MODE=mysql` 但 `MYSQL_DSN` 为空时，服务会在启动阶段返回明确配置错误。

### Docker Compose

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

Compose 默认使用 MySQL 持久化和 Mock Agent。`migrate` 服务会在后端启动前执行 `/app/migrate -dir /app/migrations`。如果要清空本地容器数据：

```bash
docker compose down -v
```

部署时建议保持前端 `VITE_API_BASE_URL=/api/v1`，由 Nginx 反向代理到后端；如果前后端分域部署，需要把前端域名加入 `CORS_ALLOWED_ORIGINS`。
后端 Compose 默认设置 `REDIS_ENABLED=true` 并连接 `redis:6379`，因此 Redis 健康检查失败时后端不会启动。

### LLM Streaming / ASR Provider

自动测试和默认本地运行不请求真实模型：

```bash
LLM_USE_MOCK=true
LLM_FALLBACK_TO_MOCK=true
CORRECTION_USE_MOCK=true
SCORING_USE_MOCK=true
SUMMARY_USE_MOCK=true
ASR_USE_MOCK=true
```

开启真实 LLM streaming 时需要配置：

```bash
LLM_USE_MOCK=false
LLM_PROVIDER=openai-compatible
LLM_BASE_URL=https://your-provider.example.com/v1
LLM_API_KEY=your-api-key
LLM_MODEL=your-model
LLM_FALLBACK_TO_MOCK=true
```

`GET /api/v1/sessions/:id/stream` 会推送真实模型 delta；如果 `LLM_FALLBACK_TO_MOCK=true` 且上游失败，会降级为本地 fake streaming，保证演示链路继续可用。设置 `LLM_FALLBACK_TO_MOCK=false` 时，上游 streaming 失败会通过 SSE `error` 事件和普通 JSON 错误返回。Correction / Scoring / Summary 仍通过各自 Mock 开关控制。

开启腾讯云 ASR 时需要配置完整密钥，不会在配置缺失时静默降级到 Mock：

```bash
ASR_USE_MOCK=false
ASR_PROVIDER=tencent
TENCENT_ASR_APP_ID=your-app-id
TENCENT_ASR_SECRET_ID=your-secret-id
TENCENT_ASR_SECRET_KEY=your-secret-key
TENCENT_ASR_ENGINE_TYPE=16k_en
TENCENT_ASR_VOICE_FORMAT=ogg-opus
```

腾讯云模式使用 `FlashRecognizer` 识别完整音频，支持单段上传和 WebSocket `end` 后的 final transcript。当前不实现腾讯云实时 partial transcript；WebSocket 分片阶段只缓存音频并返回 partial 占位。`webm` 容器当前不支持真实腾讯云识别，建议使用 `ogg-opus`、`m4a/mp4` 或 `wav`，后续可以通过后端转码兼容只支持 `webm` 的浏览器。真实密钥只应放在本地环境变量或部署密钥系统中，不要提交到 git。

### 前后端联调路径

1. 启动后端：`go run ./cmd/server`
2. 启动前端：`cd web && npm run dev`
3. 打开 `http://localhost:5173`，验证完整训练闭环：

```text
选择场景 -> 创建训练 -> 发送文本或实时录音/录音上传 -> AI 回复 -> 查看纠错评分 -> 结束训练 -> 生成报告 -> 查询历史记录 -> 回到报告/训练详情
```

### 验证命令

```bash
go test ./...
cd web
npm test
npm run build
```

查看静态前端原型参考：

```text
web/preview.html
```

该页面是早期静态交互原型，用于展示目标产品体验，不依赖后端接口。正式前端应用位于 `web/src/`，通过 `web/src/api/client.ts` 统一访问后端接口。

## 项目结构

```text
speakmate/
├── Dockerfile               # 后端 server + migration 多阶段镜像
├── docker-compose.yml       # MySQL、Redis、迁移任务、后端和前端静态服务
├── cmd/server/              # 服务入口
├── cmd/migrate/             # MySQL migration 执行入口
├── internal/
│   ├── config/              # 环境配置
│   ├── agent/               # Conversation/Feedback/Summary/ASR Agent、Prompt 和 Mock/LLM 实现
│   ├── handler/             # HTTP Handler
│   ├── infra/database/      # MySQL 连接初始化
│   ├── infra/llm/           # OpenAI-compatible LLM HTTP Client
│   ├── infra/asr/           # Tencent Cloud ASR FlashRecognizer adapter
│   ├── middleware/          # CORS、请求日志、recover、请求超时
│   ├── repository/          # memory/mysql 仓库实现
│   ├── response/            # 统一响应结构
│   ├── security/            # 日志敏感信息脱敏
│   ├── state/               # Session 短期状态 store，支持 memory / Redis
│   ├── stream/              # Session 级 SSE 事件模型和 memory / Redis 事件总线
│   └── router/              # Gin 路由
├── migrations/              # MySQL 表结构和默认场景 seed
├── web/                     # Vite + React + TypeScript 前端应用
│   ├── src/                 # 页面、组件、API client 和类型定义
│   ├── Dockerfile           # 前端构建 + Nginx 静态部署
│   ├── nginx.conf           # 静态资源、API、SSE、WebSocket 反向代理
│   ├── package.json
│   └── preview.html         # 静态交互原型参考
├── docs/project-blueprint.md # 完整产品与技术方案
├── go.mod
├── go.sum
└── README.md
```

## 后续规划

- 扩展更多 LLM Provider，并细化真实模型联调配置；
- 评估腾讯云 `SpeechRecognizer` 实现真实 partial transcript；
- 增加后端转码以兼容只支持 `webm/opus` 的浏览器；
- 补充迁移执行工具和部署环境数据库初始化流程；
- 为 Redis 模式补充更多 Demo QA 脚本和多实例联调记录。
