# Module 3: Message API 接口文档

本文档说明当前已实现的消息发送接口。前端可以用它完成训练页里的文本发送、接收场景化 Mock AI 回复、更新对话轮次和恢复消息历史。

本模块仍然使用 Mock Conversation，不调用真实 LLM，不提供 SSE 流式输出，不处理语音 ASR/TTS。

## 基本信息

| 项目 | 说明 |
|---|---|
| API 前缀 | `/api/v1` |
| 响应格式 | JSON |
| 成功响应结构 | `{ "code": 0, "message": "success", "data": ... }` |
| 错误响应结构 | `{ "code": 业务错误码, "message": "错误说明" }` |
| 当前数据来源 | 后端内存数据 |
| 是否需要登录 | 当前版本不需要 |
| 依赖接口 | 先通过 `POST /api/v1/sessions` 创建 `running` Session |

## 状态规则

消息发送只允许发生在 `running` 状态的 Session 上。

```text
POST /api/v1/sessions
  -> running Session
  -> POST /api/v1/sessions/:id/messages
  -> turn_count + 1
  -> GET /api/v1/sessions/:id 可看到消息历史
  -> POST /api/v1/sessions/:id/finish
  -> finished Session，不再允许发送消息
```

约束：

- `content` 会在后端去除首尾空白后保存。
- 去除首尾空白后为空字符串时返回 `400 / 3002`。
- 一次发送成功会保存两条消息：用户消息和 AI Mock 回复。
- `turn_count` 表示成功发送的用户轮次数。第一次发送成功后为 `1`。
- 当前使用内存存储，服务重启后 Session 和消息都会丢失。

## 错误码

| HTTP 状态码 | 业务错误码 | message | 说明 |
|---|---:|---|---|
| `400` | `2002` | `invalid session id` | 路径参数 `:id` 不是合法正整数 |
| `400` | `3001` | `invalid message request` | 请求体不是合法 JSON，或缺少 `content` 字段 |
| `400` | `3002` | `message content is required` | `content` 去除首尾空白后为空 |
| `404` | `2003` | `session not found` | Session 不存在 |
| `409` | `2004` | `session already finished` | Session 已结束，不能继续发送消息 |
| `500` | `500` | `internal server error` | 非预期服务端错误 |

## 发送文本消息

向一个正在进行中的训练 Session 发送用户文本，并获得一条场景化 Mock AI 回复。

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
| `content` | string | 是 | 用户输入文本，后端会去除首尾空白，去除后不能为空 |

示例请求：

```json
{
  "content": "I am a computer science student, and I built a robot control project last semester."
}
```

### 成功响应

HTTP 状态码：`200`

| 字段 | 类型 | 说明 |
|---|---|---|
| `user_message` | object | 本次保存的用户消息 |
| `user_message.id` | number | 消息 ID，当前为 Session 内按消息顺序递增 |
| `user_message.session_id` | number | 所属 Session ID |
| `user_message.role` | string | 固定为 `user` |
| `user_message.content` | string | 去除首尾空白后的用户文本 |
| `user_message.stage` | string | 用户消息所属训练阶段 |
| `user_message.created_at` | string | 消息创建时间，RFC3339 格式 |
| `ai_message` | object | 本次生成并保存的 AI Mock 回复 |
| `ai_message.id` | number | 消息 ID，紧跟用户消息之后 |
| `ai_message.session_id` | number | 所属 Session ID |
| `ai_message.role` | string | 固定为 `assistant` |
| `ai_message.content` | string | 场景化 Mock AI 回复 |
| `ai_message.stage` | string | AI 回复推进到的训练阶段 |
| `ai_message.created_at` | string | 消息创建时间，RFC3339 格式 |
| `stage` | string | 当前 AI 回复推进到的训练阶段，等同于 `ai_message.stage` |
| `turn_count` | number | 发送成功后的对话轮次 |

示例：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "user_message": {
      "id": 1,
      "session_id": 1,
      "role": "user",
      "content": "I am a computer science student, and I built a robot control project last semester.",
      "stage": "自我介绍",
      "created_at": "2026-06-06T12:00:00Z"
    },
    "ai_message": {
      "id": 2,
      "session_id": 1,
      "role": "assistant",
      "content": "That project sounds useful. Could you explain your role in the project and one technical challenge you solved?",
      "stage": "项目经历",
      "created_at": "2026-06-06T12:00:00Z"
    },
    "stage": "项目经历",
    "turn_count": 1
  }
}
```

### 错误响应

#### Session ID 非法

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

#### 请求体非法

请求体不是合法 JSON：

```json
{
```

响应：

```json
{
  "code": 3001,
  "message": "invalid message request"
}
```

#### 缺少 content

请求体：

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

#### content 为空

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

#### Session 不存在

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

#### Session 已结束

对已经 `finished` 的 Session 继续发送消息：

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

### curl 示例

启动服务：

```bash
go run ./cmd/server
```

创建训练 Session：

```bash
curl -X POST http://localhost:8080/api/v1/sessions \
  -H "Content-Type: application/json" \
  -d '{"scenario_id":1}'
```

发送第一条消息：

```bash
curl -X POST http://localhost:8080/api/v1/sessions/1/messages \
  -H "Content-Type: application/json" \
  -d '{"content":"I am a computer science student, and I built a robot control project last semester."}'
```

查询消息历史：

```bash
curl http://localhost:8080/api/v1/sessions/1
```

发送第二条消息：

```bash
curl -X POST http://localhost:8080/api/v1/sessions/1/messages \
  -H "Content-Type: application/json" \
  -d '{"content":"I mainly worked on the backend API and deployment."}'
```

结束训练：

```bash
curl -X POST http://localhost:8080/api/v1/sessions/1/finish
```

结束后继续发送，预期返回 `409 / 2004`：

```bash
curl -X POST http://localhost:8080/api/v1/sessions/1/messages \
  -H "Content-Type: application/json" \
  -d '{"content":"Can we continue?"}'
```

错误场景：

```bash
curl -X POST http://localhost:8080/api/v1/sessions/999/messages \
  -H "Content-Type: application/json" \
  -d '{"content":"Hello"}'

curl -X POST http://localhost:8080/api/v1/sessions/abc/messages \
  -H "Content-Type: application/json" \
  -d '{"content":"Hello"}'

curl -X POST http://localhost:8080/api/v1/sessions/1/messages \
  -H "Content-Type: application/json" \
  -d '{"content":"   "}'
```

## Mock 回复规则

当前后端按 Session 关联场景的 `scenario.code` 生成 Mock AI 回复：

| 场景 code | 回复方向 |
|---|---|
| `interview` | 围绕项目经历、技术细节、团队协作继续追问 |
| `restaurant` | 围绕菜品偏好、过敏或饮食限制、订单确认继续追问 |
| `meeting` | 围绕优先级、阻塞点、下一步行动继续追问 |

阶段规则：

- 用户消息的 `stage` 使用发送前的当前阶段。
- AI 消息的 `stage` 推进到下一阶段。
- 轮次数超过阶段数量后，继续使用最后一个阶段。

## 前端使用建议

- 进入训练页时先调用 `POST /api/v1/sessions`，保存返回的 `session_id`。
- 创建响应里的 `opening_message` 可作为第一条 AI 开场消息渲染，但它当前不会写入 `messages` 历史。
- 用户点击发送时，先在前端禁用发送按钮，等待 `POST /api/v1/sessions/:id/messages` 返回。
- 成功后把 `user_message` 和 `ai_message` 同时追加到本地消息列表，不需要再立刻请求详情接口。
- 用返回的 `turn_count` 更新训练轮次，用返回的 `stage` 更新当前阶段展示。
- 页面刷新、重新进入训练页或本地状态丢失时，调用 `GET /api/v1/sessions/:id` 恢复 `messages`、`turn_count` 和 `status`。
- `status === "finished"` 或接口返回 `409 / 2004` 时，禁用输入框和发送按钮。
- `400 / 3002` 可以直接提示“请输入内容后再发送”。
- `404 / 2003` 可以提示“训练不存在或服务已重启”，因为当前版本使用内存存储。
- 前端不要根据 Mock 回复文本写死业务判断，后续接入真实 LLM 后 `ai_message.content` 会变化。
- `created_at` 当前是 UTC RFC3339 字符串，展示时由前端按本地时区格式化。

## 推荐调用流程

```text
1. GET /api/v1/scenarios
2. 用户选择场景
3. POST /api/v1/sessions
4. 渲染 opening_message
5. POST /api/v1/sessions/:id/messages
6. 追加 user_message 和 ai_message
7. GET /api/v1/sessions/:id 重新进入页面时恢复状态
8. POST /api/v1/sessions/:id/finish
```

## 验证命令

自动化测试：

```bash
go test ./internal/router ./internal/service
```

如果本地 Go 模块缓存没有写权限，可以使用仓库内已有的本地验证 modfile：

```bash
go test -modfile=tmp/routeverify.mod ./...
```
