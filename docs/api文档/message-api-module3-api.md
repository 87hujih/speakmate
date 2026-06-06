# Module 3: Message API + Conversation Agent 接口文档

本文档说明当前已实现的文本消息发送接口。前端可以用它完成训练页的核心对话闭环：

```text
创建 Session -> 发送用户文本 -> 收到 Conversation Agent 回复 -> 查询消息历史 -> 结束训练
```

当前版本只支持文本消息和普通 JSON 响应，已抽象 `ConversationAgent`，启动时固定注入本地 `MockConversationAgent`。当前不接真实 LLM、不做 SSE 流式输出、不处理语音。

## 基本信息

| 项目 | 说明 |
|---|---|
| API 前缀 | `/api/v1` |
| 响应格式 | JSON |
| 成功响应结构 | `{ "code": 0, "message": "success", "data": ... }` |
| 错误响应结构 | `{ "code": 业务错误码, "message": "错误说明" }` |
| 当前数据来源 | 后端内存数据 |
| 是否需要登录 | 当前版本不需要 |
| 前置条件 | 必须先通过 `POST /api/v1/sessions` 创建 `running` Session |

## 状态与轮次规则

| 规则 | 说明 |
|---|---|
| 可发送状态 | 只有 `running` Session 可以发送消息 |
| 禁止发送状态 | `finished` Session 会返回 `409 / 2004` |
| 轮次递增 | 每次成功发送用户消息并生成 AI 回复后，`turn_count + 1` |
| 消息保存 | 每次成功发送会保存 2 条消息：用户消息和 AI 消息 |
| 消息历史 | 调用 `GET /api/v1/sessions/:id` 可以看到已保存的消息列表 |
| 数据持久性 | 当前使用内存存储，服务重启后 Session 和消息数据会丢失 |

## Conversation Agent 规则

当前 AI 回复由 `internal/agent` 下的 `MockConversationAgent` 生成，不请求外部模型。消息发送业务只依赖 `ConversationAgent` 接口，后续可在启动装配时替换为真实 LLM Agent。

| 场景编码 | 回复方向 |
|---|---|
| `interview` | 围绕项目经历、技术设计、团队协作继续追问 |
| `restaurant` | 围绕菜单偏好、过敏信息、饮品或订单确认继续追问 |
| `meeting` | 围绕阻塞点、方案取舍、下一步行动继续追问 |

阶段字段 `stage` 根据当前场景的 `stages` 和 `turn_count` 推进。轮次数超过阶段数量时使用最后一个阶段。

## 错误码

| HTTP 状态码 | 业务错误码 | message | 说明 |
|---|---:|---|---|
| `400` | `2002` | `invalid session id` | 路径参数 `:id` 不是合法正整数 |
| `400` | `3001` | `invalid message request` | 请求体不是合法 JSON，或缺少 `content` |
| `400` | `3002` | `message content is required` | `content` 去除首尾空白后为空 |
| `404` | `2003` | `session not found` | Session 不存在 |
| `409` | `2004` | `session already finished` | Session 已结束，不允许继续发送消息 |
| `500` | `500` | `internal server error` | 非预期服务端错误 |

## 发送文本消息

向一个正在进行中的训练 Session 发送用户文本，并返回本轮保存的用户消息和 AI 回复。

### 接口路径

```http
POST /api/v1/sessions/:id/messages
```

### 路径参数

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `id` | number | 是 | Session ID，必须是正整数 |

### 请求参数

请求体类型：`application/json`

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `content` | string | 是 | 用户输入文本。后端会去除首尾空白，去除后不能为空 |

示例请求：

```json
{
  "content": "I worked on a robot control project in college."
}
```

### 成功响应

HTTP 状态码：`200`

| 字段 | 类型 | 说明 |
|---|---|---|
| `user_message` | object | 本次保存的用户消息 |
| `user_message.id` | number | 消息 ID |
| `user_message.session_id` | number | 所属 Session ID |
| `user_message.role` | string | 固定为 `user` |
| `user_message.content` | string | 去除首尾空白后的用户输入 |
| `user_message.stage` | string | 用户消息所属训练阶段 |
| `user_message.created_at` | string | 消息创建时间，RFC3339 格式 |
| `ai_message` | object | 本次生成并保存的 AI 回复 |
| `ai_message.id` | number | 消息 ID |
| `ai_message.session_id` | number | 所属 Session ID |
| `ai_message.role` | string | 固定为 `ai` |
| `ai_message.content` | string | AI 回复内容 |
| `ai_message.stage` | string | AI 回复推进到的训练阶段 |
| `ai_message.created_at` | string | 消息创建时间，RFC3339 格式 |
| `stage` | string | 当前响应对应的训练阶段，与 `ai_message.stage` 一致 |
| `next_goal` | string | Agent 给出的下一步追问目标，当前由 Mock Agent 稳定返回 |
| `turn_count` | number | 发送成功后的 Session 对话轮次 |

示例响应：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "user_message": {
      "id": 1,
      "session_id": 1,
      "role": "user",
      "content": "I worked on a robot control project in college.",
      "stage": "自我介绍",
      "created_at": "2026-06-06T07:30:00Z"
    },
    "ai_message": {
      "id": 2,
      "session_id": 1,
      "role": "ai",
      "content": "That project sounds relevant. Could you explain your role in the project and one technical challenge you solved?",
      "stage": "项目经历",
      "created_at": "2026-06-06T07:30:00Z"
    },
    "stage": "项目经历",
    "next_goal": "ask user to describe personal project contribution",
    "turn_count": 1
  }
}
```

## 错误响应

### Session ID 非法

请求：

```http
POST /api/v1/sessions/abc/messages
```

响应：

```json
{
  "code": 2002,
  "message": "invalid session id"
}
```

### 请求体非法

请求体不是合法 JSON，或缺少 `content`：

```json
{}
```

响应：

```json
{
  "code": 3001,
  "message": "invalid message request"
}
```

### 消息内容为空

请求体：

```json
{
  "content": "   "
}
```

响应：

```json
{
  "code": 3002,
  "message": "message content is required"
}
```

### Session 不存在

请求：

```http
POST /api/v1/sessions/999/messages
```

响应：

```json
{
  "code": 2003,
  "message": "session not found"
}
```

### Session 已结束

对 `finished` Session 继续发送消息：

```http
POST /api/v1/sessions/1/messages
```

响应：

```json
{
  "code": 2004,
  "message": "session already finished"
}
```

## curl 示例

### 完整训练消息闭环

```bash
curl -X POST http://localhost:8080/api/v1/sessions \
  -H "Content-Type: application/json" \
  -d '{"scenario_id":1}'

curl -X POST http://localhost:8080/api/v1/sessions/1/messages \
  -H "Content-Type: application/json" \
  -d '{"content":"I worked on a robot control project in college."}'

curl http://localhost:8080/api/v1/sessions/1

curl -X POST http://localhost:8080/api/v1/sessions/1/messages \
  -H "Content-Type: application/json" \
  -d '{"content":"I was responsible for the backend API and debugging control issues."}'

curl http://localhost:8080/api/v1/sessions/1

curl -X POST http://localhost:8080/api/v1/sessions/1/finish

curl -X POST http://localhost:8080/api/v1/sessions/1/messages \
  -H "Content-Type: application/json" \
  -d '{"content":"Can we continue?"}'
```

### 错误场景

```bash
curl -X POST http://localhost:8080/api/v1/sessions/abc/messages \
  -H "Content-Type: application/json" \
  -d '{"content":"Hello"}'

curl -X POST http://localhost:8080/api/v1/sessions/999/messages \
  -H "Content-Type: application/json" \
  -d '{"content":"Hello"}'

curl -X POST http://localhost:8080/api/v1/sessions/1/messages \
  -H "Content-Type: application/json" \
  -d '{}'

curl -X POST http://localhost:8080/api/v1/sessions/1/messages \
  -H "Content-Type: application/json" \
  -d '{"content":"   "}'
```

## 前端使用建议

- 训练页发送按钮调用 `POST /api/v1/sessions/:id/messages`。
- 发送前前端也应做一次 `content.trim()` 校验，避免空内容请求。
- 发送中禁用输入框和发送按钮，避免用户重复点击造成连续轮次递增。
- 成功后把 `user_message` 和 `ai_message` 直接追加到本地消息列表，不需要立即重新拉取 Session。
- 如果需要恢复页面状态，调用 `GET /api/v1/sessions/:id`，使用返回的 `messages` 和 `turn_count` 重新渲染。
- `ai_message.role` 当前固定为 `ai`，前端不要按 `assistant` 判断。
- `turn_count` 表示成功完成的用户输入轮次，不等于消息条数。一次成功发送会新增 2 条消息。
- `stage` 可以用于训练进度展示，但不要把它作为强业务状态机，后续真实 Agent 接入后可能会调整阶段策略。
- 收到 `409 / 2004` 后应禁用输入框和发送按钮，并提示“本次训练已结束”。
- 收到 `404 / 2003` 可以提示“训练不存在或服务已重启”，因为当前版本使用内存存储。
- 收到 `400 / 2002` 应检查前端路由参数或本地保存的 `session_id`。
- 收到 `400 / 3001` 或 `400 / 3002` 应提示用户重新输入有效文本。

## 推荐调用流程

```text
1. GET /api/v1/scenarios
2. POST /api/v1/sessions
3. 渲染 create session 返回的 opening_message
4. POST /api/v1/sessions/:id/messages
5. 将 user_message 和 ai_message 追加到前端消息列表
6. 需要恢复状态时 GET /api/v1/sessions/:id
7. POST /api/v1/sessions/:id/finish
8. finished 后禁用继续发送
```

## 验证命令

启动服务：

```bash
go run ./cmd/server
```

自动化测试：

```bash
go test ./internal/router ./internal/service
```
