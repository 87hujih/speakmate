# SpeakMate AI

> 场景化 AI 英语口语陪练系统，围绕面试、点餐、会议等真实场景，帮助用户完成英语对话练习、表达纠错、能力评分和课后复盘。

SpeakMate AI 不是一个开放闲聊机器人。它更像一个有训练目标的英语陪练教练：先给用户一个具体场景，再通过 AI 追问推动对话，最后把用户的表达问题整理成可执行的练习建议。

本项目面向七牛云 XEngineer 暑期实训营「AI 英语口语陪练」议题设计。当前仓库已完成 Go + Gin 后端基础骨架、正式 Vite + React + TypeScript 前端应用和静态前端原型参考。完整技术方案见 [docs/project-blueprint.md](docs/project-blueprint.md)。

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
| 配置加载 | 已完成 | 支持 `APP_PORT`、LLM、反馈 Mock 和存储模式环境变量，默认端口 `8080` |
| 统一响应结构 | 已完成 | 成功响应格式为 `{ code, message, data }` |
| 前端原型 | 已完成 | `web/preview.html` 作为本地静态原型参考 |
| 正式前端应用 | 已完成（第一版） | `web/` 已接入 Vite + React + TypeScript，覆盖场景选择、文本训练、反馈评分、报告和历史记录 |
| 场景 API（Module 1） | 已完成 | 场景列表和详情接口，见 [docs/api文档/scenario-api-module1-api.md](docs/api文档/scenario-api-module1-api.md) |
| 训练 Session API（Module 2） | 已完成 | 创建、查询、结束训练 Session，见 [docs/api文档/session-api-module2-api.md](docs/api文档/session-api-module2-api.md) |
| 消息 API（Module 3） | 已完成 | 发送文本、AI 回复、轮次更新和消息历史查询，见 [docs/api文档/message-api-module3-api.md](docs/api文档/message-api-module3-api.md) |
| 反馈 API（Module 4） | 已完成（第一版） | 消息发送后同步生成纠错/评分摘要，支持单条消息纠错、Session 纠错列表和当前评分查询，见 [docs/api文档/feedback-api.md](docs/api文档/feedback-api.md) |
| 课后报告 API（Module 5） | 已完成（第一版） | 训练结束后可基于消息、纠错和评分生成结构化报告，支持重复查询，见 [docs/api文档/report-api.md](docs/api文档/report-api.md) |
| MySQL 持久化与历史记录 | 已完成（第一版） | 支持 `memory` / `mysql` 存储模式切换，Session、Message、Correction、Score、Report 可落库，历史列表见 [docs/api文档/history-api.md](docs/api文档/history-api.md) |
| SSE 流式事件 | 已完成（第一版） | 支持 `GET /api/v1/sessions/:id/stream`，推送 AI 回复分片、纠错、评分、报告和错误事件，见 [docs/api文档/sse-api.md](docs/api文档/sse-api.md) |
| Conversation Agent | 已接入 | 默认使用 Mock；配置 API Key 且关闭 Mock 后使用 OpenAI-compatible LLM，失败时降级 Mock |
| AI 纠错、评分与总结 | 已完成（第一版） | Correction / Scoring / Summary 模型、Mock/LLM Agent、内存 Feedback/Report Repository、fail-open 降级和查询 API 已接入 |
| 语音能力 | 规划中 | 浏览器录音、ASR、WebSocket 音频分片 |

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

默认配置会使用本地 Mock Agent 和内存存储，不需要 API Key 或数据库：

```bash
STORAGE_MODE=memory
LLM_USE_MOCK=true
go run ./cmd/server
```

启动后端服务：

```bash
go run ./cmd/server
```

检查服务状态：

```bash
curl http://localhost:8080/health
```

预期响应：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "status": "ok"
  }
}
```

运行测试：

```bash
go test ./...
```

### 前端启动

首次启动前安装依赖：

```bash
cd web
npm install
```

如果前后端分开启动，建议显式配置后端 API 地址：

```powershell
$env:VITE_API_BASE_URL="http://localhost:8080/api/v1"
npm run dev
```

macOS / Linux：

```bash
VITE_API_BASE_URL=http://localhost:8080/api/v1 npm run dev
```

前端默认端口是 `5173`。如果不设置 `VITE_API_BASE_URL`，前端会请求同源 `/api/v1`，适合反向代理或同源部署场景。

前端验证命令：

```bash
cd web
npm test
npm run build
```

### 前后端联调路径

1. 启动后端：

```bash
go run ./cmd/server
```

2. 启动前端并配置 API 地址：

```powershell
cd web
$env:VITE_API_BASE_URL="http://localhost:8080/api/v1"
npm run dev
```

3. 在浏览器打开 `http://localhost:5173`，验证完整文本训练闭环：

```text
选择场景 -> 创建训练 -> 发送文本消息 -> AI 回复 -> 查看纠错评分 -> 结束训练 -> 生成报告 -> 查询历史记录 -> 回到报告/训练详情
```

### 存储配置

默认 `STORAGE_MODE=memory`，适合本地开发和自动测试；服务重启后训练数据会丢失。

```bash
STORAGE_MODE=memory
go run ./cmd/server
```

切换 MySQL 持久化前，先创建数据库并按顺序执行 migrations：

```bash
mysql -u root -p -e 'CREATE DATABASE IF NOT EXISTS speakmate DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;'
mysql -u root -p speakmate < migrations/001_create_core_tables.sql
mysql -u root -p speakmate < migrations/002_seed_default_scenarios.sql
```

然后设置 DSN 启动服务：

```bash
STORAGE_MODE=mysql
MYSQL_DSN='speakmate:password@tcp(127.0.0.1:3306)/speakmate?parseTime=true&loc=UTC'
go run ./cmd/server
```

`STORAGE_MODE=mysql` 但 `MYSQL_DSN` 为空时，服务会在启动阶段返回明确配置错误。示例环境变量见 [.env.example](.env.example)。

### LLM 配置

自动测试只使用 Fake / Mock，不会请求真实模型。本地默认 `LLM_USE_MOCK=true`，只有配置完整且关闭 Mock 时才会请求 OpenAI-compatible API：

```bash
APP_PORT=8080
LLM_PROVIDER=openai-compatible
LLM_BASE_URL=https://api.example.com/v1
LLM_API_KEY=replace-with-your-api-key
LLM_MODEL=replace-with-your-model
LLM_TIMEOUT_SECONDS=30
LLM_USE_MOCK=false
go run ./cmd/server
```

如果 `LLM_USE_MOCK=true`，或 `LLM_BASE_URL` / `LLM_API_KEY` / `LLM_MODEL` 不完整，服务会继续使用 Mock Agent。真实 LLM 调用失败时当前启动策略会降级为 Mock 回复，保证本地演示链路可用。

### 反馈 Mock / LLM 配置

纠错、评分和课后总结默认同样使用 Mock，不依赖真实模型：

```bash
LLM_USE_MOCK=true
CORRECTION_USE_MOCK=true
SCORING_USE_MOCK=true
SUMMARY_USE_MOCK=true
FEEDBACK_FAIL_OPEN=true
go run ./cmd/server
```

如果要让纠错、评分和总结也使用真实 OpenAI-compatible LLM，需要关闭全局 Mock 和对应 Mock：

```bash
LLM_PROVIDER=openai-compatible
LLM_BASE_URL=https://api.example.com/v1
LLM_API_KEY=replace-with-your-api-key
LLM_MODEL=replace-with-your-model
LLM_USE_MOCK=false
CORRECTION_USE_MOCK=false
SCORING_USE_MOCK=false
SUMMARY_USE_MOCK=false
FEEDBACK_FAIL_OPEN=true
go run ./cmd/server
```

`FEEDBACK_FAIL_OPEN=true` 表示反馈生成失败时不阻断主对话链路；设置为 `false` 时，纠错或评分失败会让消息接口返回 `502 / 3004 feedback agent failed`。
课后报告的 Summary Agent 自带 Mock fallback；模型调用或 JSON 解析失败时会尽量返回 Mock 报告内容。关闭 `SUMMARY_USE_MOCK` 且 LLM 配置完整时才会请求真实模型。

查看静态前端原型参考：

```text
web/preview.html
```

该页面是早期静态交互原型，用于展示目标产品体验，不依赖后端接口。正式前端应用位于 `web/src/`，通过 `web/src/api/client.ts` 统一访问后端接口。

## 项目结构

```text
speakmate/
├── cmd/server/              # 服务入口
├── internal/
│   ├── config/              # 环境配置
│   ├── agent/               # Conversation Agent、Prompt 和 Mock/LLM 实现
│   ├── handler/             # HTTP Handler
│   ├── infra/database/      # MySQL 连接初始化
│   ├── infra/llm/           # OpenAI-compatible LLM HTTP Client
│   ├── repository/          # memory/mysql 仓库实现
│   ├── response/            # 统一响应结构
│   ├── stream/              # Session 级 SSE 事件模型和内存事件总线
│   └── router/              # Gin 路由
├── migrations/              # MySQL 表结构和默认场景 seed
├── web/                     # Vite + React + TypeScript 前端应用
│   ├── src/                 # 页面、组件、API client 和类型定义
│   ├── package.json
│   └── preview.html         # 静态交互原型参考
├── docs/project-blueprint.md # 完整产品与技术方案
├── go.mod
├── go.sum
└── README.md
```

## 后续规划

- 接入 LLM，完成基于场景的真实 AI 追问；
- 将当前模拟 AI 回复分片升级为真实 LLM streaming；
- 接入 ASR，支持浏览器录音和语音识别；
- 补充迁移执行工具和部署环境数据库初始化流程；
- 使用 Redis 管理训练过程中的上下文和临时状态。
