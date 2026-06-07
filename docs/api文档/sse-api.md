# SSE Stream API 接口文档

本文档说明 Session 级 SSE 流式事件接口。前端可以用 `EventSource` 监听一次训练中的 AI 回复、纠错、评分和报告生成事件，同时继续保留普通 JSON API 作为完整结果来源。

## 基本信息

| 项目 | 说明 |
|---|---|
| API 前缀 | `/api/v1` |
| 连接方式 | Server-Sent Events |
| 响应 `Content-Type` | `text/event-stream; charset=utf-8` |
| 是否需要登录 | 当前版本不需要 |
| 是否回放历史事件 | 不回放。只推送连接建立后的实时事件 |
| 心跳 | 服务端定期发送 `: ping` 注释帧 |

## 建立连接

```http
GET /api/v1/sessions/:id/stream
```

路径参数：

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `id` | number | 是 | Session ID，必须是正整数 |

非法 `id` 会在建立 SSE 前返回普通 JSON 错误：

```json
{
  "code": 2002,
  "message": "invalid session id"
}
```

## 事件帧格式

每个业务事件都使用 SSE 标准格式：

```text
event: ai_message_done
data: {"type":"ai_message_done","session_id":1,"payload":{"message_id":2,"content":"...","stage":"项目经历"},"created_at":"2026-06-07T03:00:00Z"}
```

`event` 行等于业务事件类型。`data` 行是完整事件 JSON：

| 字段 | 类型 | 说明 |
|---|---|---|
| `type` | string | 事件类型 |
| `session_id` | number | 所属 Session ID |
| `payload` | object | 事件数据，不同事件类型字段不同 |
| `created_at` | string | 事件创建时间，RFC3339 格式 |

## 事件类型

| 事件 | 触发时机 | payload |
|---|---|---|
| `ai_message_delta` | AI 回复保存后模拟发送回复分片 | `{ "message_id": number, "delta": string }` |
| `ai_message_done` | AI 完整回复保存后 | `{ "message_id": number, "content": string, "stage": string }` |
| `correction_done` | 本轮纠错保存后 | `{ "message_id": number, "has_errors": boolean, "error_count": number }` |
| `score_updated` | 当前评分保存后 | `{ "message_id": number, "total_score": number, "grammar": number, "expression": number }` |
| `report_done` | 课后报告生成并保存后 | `{ "total_score": number, "summary": string }` |
| `error` | 消息反馈或报告生成失败时 | `{ "code": string, "message": string }` |

第一版不要求真实 LLM streaming。`ai_message_delta` 会先以单个完整回复文本作为模拟分片发送，后续可替换为真实模型 token/句子分片，事件结构保持不变。

## 客户端示例

```js
const source = new EventSource('/api/v1/sessions/1/stream');

source.addEventListener('ai_message_delta', (event) => {
  const data = JSON.parse(event.data);
  appendAssistantDelta(data.payload.delta);
});

source.addEventListener('ai_message_done', (event) => {
  const data = JSON.parse(event.data);
  replaceAssistantMessage(data.payload);
});

source.addEventListener('correction_done', () => {
  refreshCorrections();
});

source.addEventListener('score_updated', (event) => {
  const data = JSON.parse(event.data);
  updateScore(data.payload);
});

source.addEventListener('report_done', () => {
  refreshReport();
});

source.addEventListener('error', (event) => {
  const data = JSON.parse(event.data);
  showError(data.payload.message);
});
```

## 推荐调用流程

```text
1. POST /api/v1/sessions 创建 Session
2. GET /api/v1/sessions/:id/stream 建立 SSE
3. POST /api/v1/sessions/:id/messages 发送文本消息
4. SSE 接收 ai_message_delta、ai_message_done、correction_done、score_updated
5. POST /api/v1/sessions/:id/finish 结束训练
6. POST /api/v1/sessions/:id/report 生成报告
7. SSE 接收 report_done
```

普通 JSON 接口仍然返回完整结果。前端应把 SSE 作为实时体验增强，不应只依赖 SSE 保存最终状态。

## curl 验证

终端 A 建立 SSE 连接：

```bash
curl -N http://localhost:8080/api/v1/sessions/1/stream
```

终端 B 发送消息或生成报告：

```bash
curl -X POST http://localhost:8080/api/v1/sessions/1/messages \
  -H "Content-Type: application/json" \
  -d '{"content":"I am study computer science and I have did a project."}'

curl -X POST http://localhost:8080/api/v1/sessions/1/report
```

客户端主动断开连接后，服务端会取消订阅并释放对应连接资源。

## 验证命令

```bash
go test ./...
```
