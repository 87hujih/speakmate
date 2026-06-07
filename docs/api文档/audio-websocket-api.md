# Audio WebSocket API 接口文档

本文档说明实时音频分片接口。服务端会缓存客户端发送的音频分片；客户端发送 `end` 后，服务端使用当前配置的 ASR Provider 生成 final transcript，并复用现有 `SendMessage` 训练链路。

```text
WebSocket 录音分片 -> end -> ASR Provider final transcript -> SendMessage -> AI 回复 -> 纠错评分
```

当前支持两种 ASR 模式：

| 模式 | partial_transcript | final_transcript |
|---|---|---|
| Mock ASR | 每个分片后返回稳定 mock partial | 使用 Mock ASR |
| 腾讯云 ASR | 不请求腾讯云实时识别，返回空 partial 占位和 sequence | 客户端 `end` 后用累计音频调用 `FlashRecognizer` |

当前版本不做音素级发音评分，也不实现腾讯云实时 partial transcript。真实实时识别需要后续单独接入腾讯云 `SpeechRecognizer`。

## 基本信息

| 项目 | 说明 |
|---|---|
| API 前缀 | `/api/v1` |
| WebSocket 地址 | `GET /api/v1/sessions/:id/audio/ws` |
| 是否需要登录 | 当前版本不需要 |
| 前置条件 | 必须先创建 `running` Session |
| 最大音频大小 | 单条连接累计 `10MB` |
| 后端接受音频类型 | 与单段上传一致：`audio/webm`、`audio/wav`、`audio/mp4`、`audio/ogg` 等 |
| 腾讯云真实识别限制 | `webm` 当前不支持；建议使用 `ogg-opus`、`m4a/mp4` 或 `wav` |

## 客户端事件

客户端可以发送 JSON 文本帧控制流程。`audio_chunk` 也支持二进制帧；浏览器 `MediaRecorder` 可以在 `start` 后直接发送每个 `Blob` 分片。

### start

```json
{
  "type": "start",
  "payload": {
    "content_type": "audio/ogg"
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
    "content_type": "audio/ogg"
  }
}
```

### partial_transcript

每个有效音频分片后返回 `partial_transcript` 事件。Mock 模式会带稳定 mock 文本；腾讯云真实模式下不会对每个 chunk 请求 `FlashRecognizer`，因此 `transcript` 可以为空，前端应把它视为占位进度事件。

```json
{
  "type": "partial_transcript",
  "session_id": 1,
  "payload": {
    "transcript": "",
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
    "transcript": "I built a speech practice app with Go.",
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
    "code": "asr_client_failed",
    "message": "asr client failed"
  }
}
```

常见错误码：

| code | message | 说明 |
|---|---|---|
| `invalid_audio_request` | `invalid audio request` | 事件格式非法、顺序非法或缺少必要字段 |
| `audio_file_required` | `audio file is required` | 分片为空或结束时没有音频 |
| `audio_file_too_large` | `audio file too large` | 累计音频超过 `10MB` |
| `audio_file_type_unsupported` | `audio file type unsupported` | `start.payload.content_type` 不在后端支持列表中 |
| `asr_client_failed` | `asr client failed` | ASR 转写失败，包括腾讯云鉴权失败、请求失败、真实模式下收到 `webm` 等 |
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
- 展示 `partial_transcript.payload.transcript` 时允许空字符串；真实腾讯云模式下它不是实时识别结果。
- 收到 `final_transcript` 后刷新训练详情；收到 `correction` / `score_updated` 后刷新反馈面板。
- 收到 `asr_client_failed` 时允许用户重新录音；如果浏览器只能录 `webm`，真实腾讯云模式下需要换支持 `ogg/mp4/wav` 的浏览器或等待后端转码能力。
- WebSocket 不可用或连接失败时，保留单段上传 fallback。

## 验证命令

```bash
go test ./...
cd web && npm test && npm run build
```
