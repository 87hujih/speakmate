# Feedback API 接口文档

本文档说明当前已实现的 AI 纠错与评分接口。前端可以通过消息发送响应拿到轻量反馈摘要，也可以查询单条消息纠错、整场训练纠错列表和当前 Session 评分。

当前版本使用内存 Feedback Repository。服务重启后 Session、消息、纠错和评分都会丢失。多轮发送时，纠错列表会按 Session 累积，`GET /api/v1/sessions/:id/scores` 返回最近一轮成功保存的当前评分。

## 基本信息

| 项目 | 说明 |
|---|---|
| API 前缀 | `/api/v1` |
| 响应格式 | JSON |
| 成功响应结构 | `{ "code": 0, "message": "success", "data": ... }` |
| 错误响应结构 | `{ "code": 业务错误码, "message": "错误说明" }` |
| 当前数据来源 | 后端内存数据 |
| 是否需要登录 | 当前版本不需要 |
| 反馈生成时机 | `POST /api/v1/sessions/:id/messages` 成功保存用户消息和 AI 消息后同步生成 |

## 当前链路

```text
用户发送消息
  -> 保存用户消息
  -> Conversation Agent 生成 AI 回复
  -> 保存 AI 消息
  -> Correction Agent 生成纠错结果
  -> Scoring Agent 基于纠错结果生成评分
  -> 保存纠错结果和当前 Session 评分
  -> 返回 AI 回复、纠错摘要和评分摘要
```

默认 `FEEDBACK_FAIL_OPEN=true`。反馈生成失败时，主对话链路优先成功，响应中的反馈摘要可能为空或只有纠错摘要；设置为 `false` 后，反馈失败会返回 `502 / 3004`。

## Mock / LLM 配置

本地开发和自动测试默认使用 Mock / Fake，不请求真实 LLM：

```bash
LLM_USE_MOCK=true
CORRECTION_USE_MOCK=true
SCORING_USE_MOCK=true
FEEDBACK_FAIL_OPEN=true
go run ./cmd/server
```

切换真实 OpenAI-compatible LLM 时，需要同时关闭全局 Mock 和反馈 Mock，并提供完整模型配置：

```bash
APP_PORT=8080
LLM_PROVIDER=openai-compatible
LLM_BASE_URL=https://api.example.com/v1
LLM_API_KEY=replace-with-your-api-key
LLM_MODEL=replace-with-your-model
LLM_TIMEOUT_SECONDS=30
LLM_USE_MOCK=false
CORRECTION_USE_MOCK=false
SCORING_USE_MOCK=false
FEEDBACK_FAIL_OPEN=true
go run ./cmd/server
```

配置规则：

- `LLM_USE_MOCK=true` 时，Conversation / Correction / Scoring 都使用 Mock。
- `CORRECTION_USE_MOCK=true` 时，纠错强制使用 Mock Correction Agent。
- `SCORING_USE_MOCK=true` 时，评分强制使用 Mock Scoring Agent。
- `LLM_BASE_URL`、`LLM_API_KEY`、`LLM_MODEL` 任一缺失时，反馈 Agent 回退到 Mock。
- 当前仅支持 `LLM_PROVIDER=openai-compatible` 的真实模型接入。
- LLM Correction / Scoring Agent 带 Mock fallback，模型调用或 JSON 解析失败时会尽量返回 Mock 结果。
- `FEEDBACK_FAIL_OPEN=true` 时反馈失败不阻断消息发送；`false` 时返回 `502 / 3004 feedback agent failed`。

## 错误码

| HTTP 状态码 | 业务错误码 | message | 说明 |
|---|---:|---|---|
| `400` | `2002` | `invalid session id` | 路径参数 `:id` 不是合法正整数 |
| `400` | `3001` | `invalid message request` | 消息请求体不是合法 JSON，或缺少 `content` |
| `400` | `3002` | `message content is required` | `content` 去除首尾空白后为空 |
| `404` | `2003` | `session not found` | Session 不存在 |
| `409` | `2004` | `session already finished` | Session 已结束，不允许继续发送消息 |
| `502` | `3004` | `feedback agent failed` | `FEEDBACK_FAIL_OPEN=false` 且反馈生成失败 |
| `400` | `4001` | `invalid feedback request` | `:message_id` 非法，或反馈查询参数非法 |
| `404` | `4002` | `correction not found` | 没有找到对应消息或 Session 的纠错结果 |
| `404` | `4003` | `score not found` | 没有找到对应 Session 的当前评分 |
| `500` | `500` | `internal server error` | 非预期服务端错误 |

## 字段结构

### 纠错结果

| 字段 | 类型 | 说明 |
|---|---|---|
| `message_id` | number | 被纠错的用户消息 ID |
| `session_id` | number | 所属 Session ID |
| `original_text` | string | 用户原始表达 |
| `corrected_text` | string | 推荐修正后的表达 |
| `errors` | array | 具体错误列表，可能为空数组 |
| `errors[].type` | string | 错误类型：`grammar`、`vocabulary`、`expression`、`structure`、`scenario` |
| `errors[].span` | string | 原句中的问题片段 |
| `errors[].suggestion` | string | 推荐替换表达 |
| `errors[].explanation` | string | 中文解释 |
| `better_expressions` | array | 更自然或更贴合场景的推荐表达，可能为空数组 |

### 评分结果

| 字段 | 类型 | 说明 |
|---|---|---|
| `message_id` | number | 当前评分对应的最近一条用户消息 ID |
| `session_id` | number | 所属 Session ID |
| `fluency` | number | 流利度，0 到 100 |
| `grammar` | number | 语法准确度，0 到 100 |
| `expression` | number | 表达自然度，0 到 100 |
| `vocabulary` | number | 词汇丰富度，0 到 100 |
| `completion` | number | 场景完成度，0 到 100 |
| `total_score` | number | 综合分，0 到 100 |
| `comment` | string | 中文评分说明 |

综合分计算规则：

```text
total_score =
0.25 * fluency
+ 0.25 * grammar
+ 0.20 * expression
+ 0.15 * vocabulary
+ 0.15 * completion
```

## 发送消息并生成反馈

```http
POST /api/v1/sessions/:id/messages
```

### 请求参数

路径参数：

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `id` | number | 是 | Session ID，必须是正整数 |

请求体：

```json
{
  "content": "I am study computer science and I have did a project."
}
```

### 成功响应

HTTP 状态码：`200`

`data` 中除 `user_message`、`ai_message`、`stage`、`next_goal`、`turn_count` 外，还包含：

| 字段 | 类型 | 说明 |
|---|---|---|
| `correction_summary.has_errors` | boolean | 本轮纠错是否发现问题 |
| `correction_summary.error_count` | number | 本轮错误数量 |
| `score_summary.total_score` | number | 本轮综合分 |
| `score_summary.grammar` | number | 本轮语法分 |
| `score_summary.expression` | number | 本轮表达自然度分 |

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
      "content": "I am study computer science and I have did a project.",
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
    "turn_count": 1,
    "correction_summary": {
      "has_errors": true,
      "error_count": 2
    },
    "score_summary": {
      "total_score": 77,
      "grammar": 72,
      "expression": 80
    }
  }
}
```

### curl 示例

```bash
curl -X POST http://localhost:8080/api/v1/sessions/1/messages \
  -H "Content-Type: application/json" \
  -d '{"content":"I am study computer science and I have did a project."}'
```

## 获取单条消息纠错结果

```http
GET /api/v1/messages/:message_id/corrections
```

### 成功响应

HTTP 状态码：`200`

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "message_id": 1,
    "session_id": 1,
    "original_text": "I am study computer science and I have did a project.",
    "corrected_text": "I am studying computer science, and I have done a project.",
    "errors": [
      {
        "type": "grammar",
        "span": "am study",
        "suggestion": "am studying",
        "explanation": "be 动词后应接现在分词。"
      },
      {
        "type": "grammar",
        "span": "have did",
        "suggestion": "have done",
        "explanation": "现在完成时中 have 后应接过去分词 done。"
      }
    ],
    "better_expressions": [
      "I major in computer science.",
      "I worked on a robotics project."
    ]
  }
}
```

### 错误响应

```json
{
  "code": 4002,
  "message": "correction not found"
}
```

### curl 示例

```bash
curl http://localhost:8080/api/v1/messages/1/corrections
curl http://localhost:8080/api/v1/messages/999/corrections
curl http://localhost:8080/api/v1/messages/abc/corrections
```

## 获取某次训练的全部纠错结果

```http
GET /api/v1/sessions/:id/corrections
```

### 成功响应

HTTP 状态码：`200`

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "message_id": 1,
      "session_id": 1,
      "original_text": "I am study computer science and I have did a project.",
      "corrected_text": "I am studying computer science, and I have done a project.",
      "errors": [
        {
          "type": "grammar",
          "span": "am study",
          "suggestion": "am studying",
          "explanation": "be 动词后应接现在分词。"
        },
        {
          "type": "grammar",
          "span": "have did",
          "suggestion": "have done",
          "explanation": "现在完成时中 have 后应接过去分词 done。"
        }
      ],
      "better_expressions": [
        "I major in computer science.",
        "I worked on a robotics project."
      ]
    }
  ]
}
```

### 错误响应

```json
{
  "code": 4002,
  "message": "correction not found"
}
```

### curl 示例

```bash
curl http://localhost:8080/api/v1/sessions/1/corrections
curl http://localhost:8080/api/v1/sessions/999/corrections
curl http://localhost:8080/api/v1/sessions/abc/corrections
```

## 获取某次训练当前评分

```http
GET /api/v1/sessions/:id/scores
```

### 成功响应

HTTP 状态码：`200`

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "message_id": 1,
    "session_id": 1,
    "fluency": 75,
    "grammar": 72,
    "expression": 80,
    "vocabulary": 76,
    "completion": 85,
    "total_score": 77,
    "comment": "用户能够表达核心意思，但存在时态和动词形式错误。"
  }
}
```

### 错误响应

```json
{
  "code": 4003,
  "message": "score not found"
}
```

### curl 示例

```bash
curl http://localhost:8080/api/v1/sessions/1/scores
curl http://localhost:8080/api/v1/sessions/999/scores
curl http://localhost:8080/api/v1/sessions/abc/scores
```

## 完整联调示例

以下示例假设服务刚启动，内存 ID 从 `1` 开始。实际联调时以前一个响应中的 `data.session_id` 和 `data.user_message.id` 为准。

```bash
curl -X POST http://localhost:8080/api/v1/sessions \
  -H "Content-Type: application/json" \
  -d '{"scenario_id":1}'

curl -X POST http://localhost:8080/api/v1/sessions/1/messages \
  -H "Content-Type: application/json" \
  -d '{"content":"I am study computer science and I have did a project."}'

curl http://localhost:8080/api/v1/messages/1/corrections
curl http://localhost:8080/api/v1/sessions/1/corrections
curl http://localhost:8080/api/v1/sessions/1/scores
```

## 前端使用建议

- 发送消息成功后，先使用响应中的 `correction_summary` 和 `score_summary` 更新轻量反馈区。
- 需要展开详情时，再调用 `GET /api/v1/messages/:message_id/corrections`。
- 训练页右侧反馈面板可以调用 `GET /api/v1/sessions/:id/corrections` 渲染累计纠错列表。
- 评分卡片调用 `GET /api/v1/sessions/:id/scores`，前端不要重复计算 `total_score`。
- 查询接口返回 `404 / 4002` 或 `404 / 4003` 时，展示“暂无反馈”或“反馈生成中”，不要作为页面级错误。
- `errors` 和 `better_expressions` 都可能为空数组，前端应兼容空状态。
- 当前存储是内存数据，服务重启后 not found 属于预期开发行为。

## 验证命令

```bash
go test ./...
```
