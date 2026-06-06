# Module 2: Session API 接口文档

本文档说明当前已实现的训练 Session 生命周期接口。前端可以用它完成创建训练、查看训练状态、结束训练这三个基础流程。

消息发送、AI Mock 回复和轮次递增由 Module 3 提供。课后报告不属于本模块。

## 基本信息

| 项目 | 说明 |
|---|---|
| API 前缀 | `/api/v1` |
| 响应格式 | JSON |
| 成功响应结构 | `{ "code": 0, "message": "success", "data": ... }` |
| 错误响应结构 | `{ "code": 业务错误码, "message": "错误说明" }` |
| 当前数据来源 | 后端内存数据 |
| 是否需要登录 | 当前版本不需要 |
| 默认用户 | 未传 `user_id` 时后端使用 `1` |

## 状态规则

当前版本只支持两个状态：

| 状态 | 说明 |
|---|---|
| `running` | Session 已创建，训练进行中 |
| `finished` | Session 已结束 |

状态流转：

```text
running -> finished
```

约束：

- 创建成功后状态直接是 `running`。
- 只有 `running` 状态可以结束。
- 已经 `finished` 的 Session 再次结束会返回 `409`。
- 当前使用内存存储，服务重启后 Session 数据会丢失。

## 错误码

| HTTP 状态码 | 业务错误码 | message | 说明 |
|---|---:|---|---|
| `400` | `2001` | `invalid session request` | 创建请求体非法，或 `scenario_id` 缺失/不是正整数 |
| `400` | `2002` | `invalid session id` | 路径参数 `:id` 不是合法正整数 |
| `404` | `1001` | `scenario not found` | 创建 Session 时场景不存在 |
| `404` | `2003` | `session not found` | Session 不存在 |
| `409` | `2004` | `session already finished` | 重复结束，或对已结束 Session 执行运行中操作 |

## 创建训练 Session

基于某个训练场景创建一次训练。用于用户点击“开始训练”时调用。

```http
POST /api/v1/sessions
```

### 请求参数

请求体类型：`application/json`

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `scenario_id` | number | 是 | 训练场景 ID，必须存在且为正整数 |
| `user_id` | number | 否 | 用户 ID，当前无登录系统时可不传，后端默认使用 `1` |

示例请求：

```json
{
  "scenario_id": 1
}
```

### 成功响应

HTTP 状态码：`200`

| 字段 | 类型 | 说明 |
|---|---|---|
| `session_id` | number | Session 内部 ID |
| `session_no` | string | 展示或排查用的 Session 编号 |
| `scenario_id` | number | 关联场景 ID |
| `status` | string | 当前状态，创建成功后为 `running` |
| `opening_message` | string | 关联场景的 AI 开场白 |

示例：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "session_id": 1,
    "session_no": "S202606060001",
    "scenario_id": 1,
    "status": "running",
    "opening_message": "Hello, welcome to the interview. Could you start by briefly introducing yourself and telling me about one project you are proud of?"
  }
}
```

### 错误响应

#### 请求体非法

请求：

```http
POST /api/v1/sessions
```

请求体：

```json
{}
```

响应：

```json
{
  "code": 2001,
  "message": "invalid session request"
}
```

#### 场景不存在

请求体：

```json
{
  "scenario_id": 999
}
```

响应：

```json
{
  "code": 1001,
  "message": "scenario not found"
}
```

### curl 示例

```bash
curl -X POST http://localhost:8080/api/v1/sessions \
  -H "Content-Type: application/json" \
  -d '{"scenario_id":1}'

curl -X POST http://localhost:8080/api/v1/sessions \
  -H "Content-Type: application/json" \
  -d '{}'

curl -X POST http://localhost:8080/api/v1/sessions \
  -H "Content-Type: application/json" \
  -d '{"scenario_id":999}'
```

## 获取训练 Session

获取一次训练的当前状态、关联场景摘要、轮次和消息列表。用于训练页初始化或刷新状态。

```http
GET /api/v1/sessions/:id
```

### 路径参数

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `id` | number | 是 | Session ID，必须是正整数 |

### 成功响应

HTTP 状态码：`200`

| 字段 | 类型 | 说明 |
|---|---|---|
| `session_id` | number | Session 内部 ID |
| `session_no` | string | Session 编号 |
| `scenario` | object | 关联场景摘要 |
| `scenario.id` | number | 场景 ID |
| `scenario.code` | string | 场景编码 |
| `scenario.name` | string | 场景名称 |
| `scenario.description` | string | 场景简介 |
| `scenario.difficulty` | string | 场景难度 |
| `status` | string | 当前状态，可能是 `running` 或 `finished` |
| `turn_count` | number | 当前对话轮次，创建后为 `0`，每次成功发送用户消息后加 `1` |
| `messages` | array | 已保存的消息列表，创建后为空数组 |
| `messages[].id` | number | 消息 ID，当前为 Session 内按消息顺序递增 |
| `messages[].session_id` | number | 消息所属 Session ID |
| `messages[].role` | string | 消息角色，可能为 `user` 或 `assistant` |
| `messages[].content` | string | 消息内容 |
| `messages[].stage` | string | 消息所属训练阶段 |
| `messages[].created_at` | string | 消息创建时间，RFC3339 格式 |
| `created_at` | string | Session 创建时间，RFC3339 格式 |
| `ended_at` | string/null | Session 结束时间，未结束时为 `null` |

示例：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "session_id": 1,
    "session_no": "S202606060001",
    "scenario": {
      "id": 1,
      "code": "interview",
      "name": "英语面试",
      "description": "练习自我介绍、项目经历和技术追问",
      "difficulty": "medium"
    },
    "status": "running",
    "turn_count": 0,
    "messages": [],
    "created_at": "2026-06-06T12:00:00Z",
    "ended_at": null
  }
}
```

### 错误响应

#### Session ID 非法

请求：

```http
GET /api/v1/sessions/abc
```

响应：

```json
{
  "code": 2002,
  "message": "invalid session id"
}
```

#### Session 不存在

请求：

```http
GET /api/v1/sessions/999
```

响应：

```json
{
  "code": 2003,
  "message": "session not found"
}
```

### curl 示例

```bash
curl http://localhost:8080/api/v1/sessions/1
curl http://localhost:8080/api/v1/sessions/999
curl http://localhost:8080/api/v1/sessions/abc
```

## 结束训练 Session

结束一次正在进行中的训练。用于用户点击“结束训练”时调用。

```http
POST /api/v1/sessions/:id/finish
```

### 路径参数

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `id` | number | 是 | Session ID，必须是正整数 |

### 请求参数

无 query 参数，无 request body。

### 成功响应

HTTP 状态码：`200`

| 字段 | 类型 | 说明 |
|---|---|---|
| `session_id` | number | Session 内部 ID |
| `status` | string | 结束后为 `finished` |
| `turn_count` | number | 当前对话轮次 |
| `ended_at` | string | Session 结束时间，RFC3339 格式 |

示例：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "session_id": 1,
    "status": "finished",
    "turn_count": 0,
    "ended_at": "2026-06-06T12:05:00Z"
  }
}
```

### 错误响应

#### Session ID 非法

请求：

```http
POST /api/v1/sessions/abc/finish
```

响应：

```json
{
  "code": 2002,
  "message": "invalid session id"
}
```

#### Session 不存在

请求：

```http
POST /api/v1/sessions/999/finish
```

响应：

```json
{
  "code": 2003,
  "message": "session not found"
}
```

#### Session 已结束

重复调用结束接口：

```http
POST /api/v1/sessions/1/finish
```

响应：

```json
{
  "code": 2004,
  "message": "session already finished"
}
```

### curl 示例

```bash
curl -X POST http://localhost:8080/api/v1/sessions/1/finish
curl -X POST http://localhost:8080/api/v1/sessions/1/finish
curl -X POST http://localhost:8080/api/v1/sessions/999/finish
curl -X POST http://localhost:8080/api/v1/sessions/abc/finish
```

## 前端使用建议

- 进入训练页前，先通过 `POST /api/v1/sessions` 创建 Session。
- 创建成功后保存 `session_id`，后续消息发送接口应使用这个 ID。
- 创建响应中的 `opening_message` 可以直接渲染为 AI 的第一条开场消息。
- 训练页刷新时调用 `GET /api/v1/sessions/:id` 恢复状态。
- `messages` 不包含创建 Session 时返回的 `opening_message`，只包含消息发送接口保存的用户消息和 AI 回复。
- `status === "finished"` 时应禁用输入框、发送按钮和结束按钮。
- 调用结束接口后，以返回的 `ended_at` 作为最终结束时间展示。
- 对 `409 / 2004` 可以提示“本次训练已结束”，不要再次重试结束请求。
- 对 `404 / 2003` 可以提示“训练不存在或服务已重启”，因为当前版本使用内存存储。
- 对 `400 / 2002` 应检查前端路由参数，避免把非数字 ID 传给接口。
- 前端统一按 `{ code, message, data }` 解析成功响应，按 `{ code, message }` 解析错误响应。

## 推荐调用流程

```text
1. GET /api/v1/scenarios
2. 用户选择场景
3. POST /api/v1/sessions
4. 渲染 opening_message
5. POST /api/v1/sessions/:id/messages
6. GET /api/v1/sessions/:id
7. POST /api/v1/sessions/:id/finish
```

## 验证命令

启动服务：

```bash
go run ./cmd/server
```

接口验证：

```bash
curl -X POST http://localhost:8080/api/v1/sessions \
  -H "Content-Type: application/json" \
  -d '{"scenario_id":1}'

curl http://localhost:8080/api/v1/sessions/1

curl -X POST http://localhost:8080/api/v1/sessions/1/finish

curl http://localhost:8080/api/v1/sessions/1

curl -X POST http://localhost:8080/api/v1/sessions/1/finish
```

自动化测试：

```bash
go test ./internal/router ./internal/service
```
