# Audio Upload API 接口文档

本文档说明单段音频上传接口。浏览器也可以使用实时 WebSocket 音频分片接口，见 [audio-websocket-api.md](audio-websocket-api.md)。
单段上传后端使用 Mock ASR 生成稳定 transcript，然后复用现有文本消息训练链路。

```text
浏览器录音 -> 上传音频 -> Mock ASR 转文本 -> SendMessage -> AI 回复 -> 纠错评分
```

当前版本不做音素级发音评分，也不请求真实 ASR 服务。

## 基本信息

| 项目 | 说明 |
|---|---|
| API 前缀 | `/api/v1` |
| 请求格式 | `multipart/form-data` |
| 音频字段名 | `audio` |
| 成功响应结构 | `{ "code": 0, "message": "success", "data": ... }` |
| 错误响应结构 | `{ "code": 业务错误码, "message": "错误说明" }` |
| 是否需要登录 | 当前版本不需要 |
| 前置条件 | 必须先通过 `POST /api/v1/sessions` 创建 `running` Session |

## 能力边界

| 能力 | 当前状态 |
|---|---|
| 单段录音上传 | 支持 |
| ASR 转写 | Mock 实现，输出稳定文本 |
| 转写后进入训练链路 | 支持，复用 `SendMessage` |
| 纠错和评分 | 支持，行为与文本消息一致 |
| 真实 ASR Provider | 未接入 |
| WebSocket 音频分片 | 支持，见 [audio-websocket-api.md](audio-websocket-api.md) |
| 音素级发音评分 | 未包含在本接口 |

## 文件限制

| 限制项 | 说明 |
|---|---|
| 最大大小 | `10MB` |
| 支持类型 | `audio/webm`、`audio/wav`、`audio/wave`、`audio/x-wav`、`audio/mpeg`、`audio/mp3`、`audio/mp4`、`audio/ogg`、`audio/x-m4a` |
| 空文件 | 返回 `400 / 7002` |
| 不支持类型 | 返回 `400 / 7004` |

## 上传音频

```http
POST /api/v1/sessions/:id/audio
```

### 路径参数

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `id` | number | 是 | Session ID，必须是正整数 |

### 请求参数

请求体类型：`multipart/form-data`

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `audio` | file | 是 | 浏览器录制或选择的音频文件 |

### 成功响应

HTTP 状态码：`200`

返回字段与文本消息接口基本一致，并额外包含 `transcript`。

| 字段 | 类型 | 说明 |
|---|---|---|
| `transcript` | string | ASR 转写文本，已去除首尾空白 |
| `user_message` | object | 使用 transcript 保存的用户消息 |
| `ai_message` | object | Conversation Agent 生成并保存的 AI 回复 |
| `stage` | string | 当前响应对应的训练阶段 |
| `next_goal` | string | Agent 给出的下一步追问目标 |
| `turn_count` | number | 上传成功后的 Session 对话轮次 |
| `correction_summary` | object | 本轮纠错摘要 |
| `score_summary` | object | 本轮评分摘要 |

示例响应：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "transcript": "I am study computer science and I have did a project.",
    "user_message": {
      "id": 1,
      "session_id": 1,
      "role": "user",
      "content": "I am study computer science and I have did a project.",
      "stage": "自我介绍",
      "created_at": "2026-06-07T03:00:00Z"
    },
    "ai_message": {
      "id": 2,
      "session_id": 1,
      "role": "ai",
      "content": "That project sounds relevant. Could you explain your role in the project and one technical challenge you solved?",
      "stage": "项目经历",
      "created_at": "2026-06-07T03:00:00Z"
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

## 错误码

| HTTP 状态码 | 业务错误码 | message | 说明 |
|---|---:|---|---|
| `400` | `2002` | `invalid session id` | 路径参数 `:id` 不是合法正整数 |
| `400` | `7001` | `invalid audio request` | multipart 请求格式非法 |
| `400` | `7002` | `audio file is required` | 缺少 `audio` 文件或文件为空 |
| `413` | `7003` | `audio file too large` | 文件超过 `10MB` |
| `400` | `7004` | `audio file type unsupported` | 音频类型不在支持列表中 |
| `502` | `7005` | `asr client failed` | ASR Client 转写失败 |
| `400` | `7006` | `audio transcript is required` | ASR 没有返回有效文本 |
| `404` | `2003` | `session not found` | Session 不存在 |
| `409` | `2004` | `session already finished` | Session 已结束，不允许继续发送音频 |
| `502` | `3003` | `conversation agent failed` | Conversation Agent 生成回复失败且没有可用降级 |
| `502` | `3004` | `feedback agent failed` | `FEEDBACK_FAIL_OPEN=false` 且纠错或评分生成失败 |
| `500` | `500` | `internal server error` | 非预期服务端错误 |

## curl 示例

```bash
curl -X POST http://localhost:8080/api/v1/sessions \
  -H "Content-Type: application/json" \
  -d '{"scenario_id":1}'

curl -X POST http://localhost:8080/api/v1/sessions/1/audio \
  -F "audio=@answer.webm;type=audio/webm"
```

## 前端使用建议

- 使用 `MediaRecorder` 录制单段音频，停止后把 `Blob` 包装成 `File`。
- 使用 `FormData` 上传，字段名固定为 `audio`；不要手动设置 `Content-Type`，让浏览器生成 multipart boundary。
- 上传中禁用文本发送和录音按钮，避免同一个 Session 并发推进多轮。
- 成功后用 `transcript` 展示转写文本，用 `user_message` / `ai_message` 更新对话 UI。
- 收到 `409 / 2004` 后禁用录音入口，并提示训练已结束。
- SSE 已连接时，音频上传同样会触发 `ai_message_delta`、`ai_message_done`、`correction_done`、`score_updated` 事件。
- 如果浏览器支持 WebSocket，训练页会优先使用实时音频分片；连接失败时可回退到本单段上传接口。

## 验证命令

```bash
go test ./...
```
