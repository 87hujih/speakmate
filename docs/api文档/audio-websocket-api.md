# Audio WebSocket API 接口文档

本文档说明实时音频分片接口。当前版本使用 Mock ASR 生成稳定 partial/final transcript；final transcript 会复用现有 `SendMessage` 训练链路。

```text
WebSocket 录音分片 -> Mock ASR partial transcript -> end -> final transcript -> SendMessage -> AI 回复 -> 纠错评分
```

当前版本不请求真实 ASR 服务，也不做音素级发音评分。单段上传接口仍见 [audio-api.md](audio-api.md)。

## 基本信息

| 项目 | 说明 |
|---|---|
| API 前缀 | `/api/v1` |
| WebSocket 地址 | `GET /api/v1/sessions/:id/audio/ws` |
| 是否需要登录 | 当前版本不需要 |
| 前置条件 | 必须先创建 `running` Session |
| 最大音频大小 | 单条连接累计 `10MB` |
| 支持音频类型 | 与单段上传一致：`audio/webm`、`audio/wav`、`audio/mp4`、`audio/ogg` 等 |

## 客户端事件

客户端可以发送 JSON 文本帧控制流程。`audio_chunk` 也支持二进制帧；浏览器 `MediaRecorder` 可以在 `start` 后直接发送每个 `Blob` 分片。

### start

```json
{
  "type": "start",
  "payload": {
    "content_type": "audio/webm"
  }
}
```

### audio_chunk

JSON 文本帧格式：

```json
{
  "type": "audio_chunk",
  "payload": {
    "sequence": 1,
    "audio_base64": "AQID"
  }
}
```

二进制帧格式：在 `start` 之后直接发送音频分片二进制内容，服务端会按连接内顺序自动计数。

### end

```json
{
  "type": "end"
}
```

## 服务端事件

服务端事件统一结构：

```json
{
  "type": "partial_transcript",
  "session_id": 1,
  "payload": {},
  "created_at": "2026-06-07T03:00:00Z"
}
```

### start

表示服务端已接受本次音频流。

```json
{
  "type": "start",
  "session_id": 1,
  "payload": {
    "content_type": "audio/webm"
  }
}
```

### partial_transcript

每个有效音频分片后返回稳定 partial transcript。

```json
{
  "type": "partial_transcript",
  "session_id": 1,
  "payload": {
    "transcript": "I am study",
    "sequence": 1
  }
}
```

### final_transcript

客户端发送 `end` 后，服务端生成 final transcript 并进入训练消息链路。

```json
{
  "type": "final_transcript",
  "session_id": 1,
  "payload": {
    "transcript": "I am study computer science and I have did a project.",
    "user_message": {},
    "ai_message": {},
    "stage": "项目经历",
    "next_goal": "ask user to describe personal project contribution",
    "turn_count": 1
  }
}
```

### correction

```json
{
  "type": "correction",
  "session_id": 1,
  "payload": {
    "has_errors": true,
    "error_count": 2
  }
}
```

### score_updated

```json
{
  "type": "score_updated",
  "session_id": 1,
  "payload": {
    "total_score": 77,
    "grammar": 72,
    "expression": 80
  }
}
```

### error

```json
{
  "type": "error",
  "session_id": 1,
  "payload": {
    "code": "audio_file_required",
    "message": "audio file is required"
  }
}
```

常见错误码：

| code | message | 说明 |
|---|---|---|
| `invalid_audio_request` | `invalid audio request` | 事件格式非法、顺序非法或缺少必要字段 |
| `audio_file_required` | `audio file is required` | 分片为空或结束时没有音频 |
| `audio_file_too_large` | `audio file too large` | 累计音频超过 `10MB` |
| `audio_file_type_unsupported` | `audio file type unsupported` | `start.payload.content_type` 不支持 |
| `asr_client_failed` | `asr client failed` | ASR 转写失败 |
| `audio_transcript_required` | `audio transcript is required` | ASR 没有返回有效文本 |
| `session_not_found` | `session not found` | Session 不存在 |
| `session_already_finished` | `session already finished` | Session 已结束 |

### end

服务端发送 `end` 后会用 WebSocket normal close `1000` 关闭连接。

```json
{
  "type": "end",
  "session_id": 1,
  "payload": {
    "reason": "client_end"
  }
}
```

## 关闭语义

- 客户端正常完成录音时发送 `end`。
- 服务端完成 final transcript、纠错摘要和评分摘要推送后发送 `end` 事件，并以 close code `1000` 关闭连接。
- 客户端直接断开时，本次连接丢弃未完成音频，不会生成 final transcript。
- WebSocket 失败不影响 `POST /api/v1/sessions/:id/messages` 和 `POST /api/v1/sessions/:id/audio`。

## 前端接入建议

- 录音开始后先发送 `start`，再按 `MediaRecorder.start(timeslice)` 产生的分片发送二进制帧。
- 展示 `partial_transcript.payload.transcript` 作为实时转写。
- 收到 `final_transcript` 后刷新训练详情；收到 `correction` / `score_updated` 后刷新反馈面板。
- WebSocket 不可用或连接失败时，保留单段上传 fallback。

## 验证命令

```bash
go test ./...
cd web && npm test && npm run build
```
