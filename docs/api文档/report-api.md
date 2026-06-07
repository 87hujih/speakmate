# Report API 接口文档

本文档说明当前已实现的课后报告接口。前端可以在用户结束训练后生成结构化报告，并重复查询已生成报告。

当前版本支持内存和 MySQL 两种 Report Repository。`STORAGE_MODE=memory` 时服务重启后 Session、消息、纠错、评分和报告都会丢失；`STORAGE_MODE=mysql` 时会持久化到 MySQL。报告生成依赖已保存的纠错和当前评分，因此需要先完成至少一轮消息发送并成功生成反馈。

## 基本信息

| 项目 | 说明 |
|---|---|
| API 前缀 | `/api/v1` |
| 响应格式 | JSON |
| 成功响应结构 | `{ "code": 0, "message": "success", "data": ... }` |
| 错误响应结构 | `{ "code": 业务错误码, "message": "错误说明" }` |
| 当前数据来源 | `STORAGE_MODE=memory` 时使用内存仓库；`STORAGE_MODE=mysql` 时使用 MySQL |
| 是否需要登录 | 当前版本不需要 |
| 生成前置条件 | Session 必须是 `finished`，且已有纠错和评分数据 |

## 当前链路

```text
创建 Session
  -> 发送至少一条消息
  -> 同步生成纠错和评分
  -> 结束 Session
  -> 生成课后报告
  -> 查询已生成报告
```

报告生成会读取：

- Scenario 摘要；
- Session 状态、轮次、起止时间和消息历史；
- Session 下全部纠错结果；
- Session 当前评分；
- Summary Agent 输出。

## Mock / LLM 配置

本地开发和自动测试默认使用 Mock / Fake，不请求真实 LLM：

```bash
LLM_USE_MOCK=true
SUMMARY_USE_MOCK=true
go run ./cmd/server
```

切换真实 OpenAI-compatible Summary Agent 时，需要关闭全局 Mock 和 Summary Mock，并提供完整模型配置：

```bash
APP_PORT=8080
LLM_PROVIDER=openai-compatible
LLM_BASE_URL=https://api.example.com/v1
LLM_API_KEY=replace-with-your-api-key
LLM_MODEL=replace-with-your-model
LLM_TIMEOUT_SECONDS=30
LLM_USE_MOCK=false
SUMMARY_USE_MOCK=false
go run ./cmd/server
```

配置规则：

- `LLM_USE_MOCK=true` 时，Summary Agent 使用 Mock。
- `SUMMARY_USE_MOCK=true` 时，强制使用 Mock Summary Agent。
- `LLM_BASE_URL`、`LLM_API_KEY`、`LLM_MODEL` 任一缺失时，Summary Agent 回退到 Mock。
- 当前仅支持 `LLM_PROVIDER=openai-compatible` 的真实模型接入。
- LLM Summary Agent 带 Mock fallback，模型调用或 JSON 解析失败时会尽量返回 Mock 结果。

## 错误码

| HTTP 状态码 | 业务错误码 | message | 说明 |
|---|---:|---|---|
| `400` | `5001` | `invalid report request` | 路径参数 `:id` 不是合法正整数 |
| `404` | `2003` | `session not found` | Session 不存在 |
| `409` | `5002` | `session not finished` | Session 尚未结束，不能生成报告 |
| `404` | `5003` | `report not found` | 查询时报告尚未生成 |
| `409` | `5004` | `report feedback missing` | 缺少纠错或评分数据 |
| `502` | `5005` | `summary agent failed` | Summary Agent 失败且没有可用降级 |
| `500` | `500` | `internal server error` | 非预期服务端错误 |

## 报告字段

| 字段 | 类型 | 说明 |
|---|---|---|
| `session_id` | number | 报告所属 Session ID |
| `scenario` | object | 场景摘要 |
| `scenario.id` | number | 场景 ID |
| `scenario.code` | string | 场景编码 |
| `scenario.name` | string | 场景名称 |
| `scenario.difficulty` | string | 场景难度 |
| `duration_seconds` | number | 训练持续秒数 |
| `turn_count` | number | 训练轮次 |
| `total_score` | number | 当前综合分 |
| `scores` | object | 当前分项评分，结构同 Feedback API 的评分结果 |
| `summary` | string | 总评 |
| `major_problems` | array | 主要问题 |
| `frequent_errors` | array | 高频错误 |
| `better_expressions` | array | 更自然表达 |
| `next_practice_plan` | array | 下一步练习建议 |
| `created_at` | string | 报告生成时间，RFC3339 格式 |

数组字段可能为空数组，不会返回 `null`。

## 生成课后报告

```http
POST /api/v1/sessions/:id/report
```

### 路径参数

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `id` | number | 是 | Session ID，必须是正整数 |

### 请求参数

无 query 参数，无 request body。

### 成功响应

HTTP 状态码：`200`

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "session_id": 1,
    "scenario": {
      "id": 1,
      "code": "interview",
      "name": "英语面试",
      "difficulty": "medium"
    },
    "duration_seconds": 180,
    "turn_count": 1,
    "total_score": 77,
    "scores": {
      "message_id": 1,
      "session_id": 1,
      "fluency": 75,
      "grammar": 72,
      "expression": 80,
      "vocabulary": 76,
      "completion": 85,
      "total_score": 77,
      "comment": "用户能够表达核心意思，但存在时态和动词形式错误。"
    },
    "summary": "英语面试训练完成 1 轮，当前综合评分 77。用户能够表达核心意思，但存在时态和动词形式错误。",
    "major_problems": [
      "语法准确度需要加强，优先检查动词形式和时态。"
    ],
    "frequent_errors": [
      "am study -> am studying",
      "have did -> have done"
    ],
    "better_expressions": [
      "I major in computer science.",
      "I worked on a robotics project."
    ],
    "next_practice_plan": [
      "用 STAR 结构重写一次项目经历回答。",
      "准备 3 个关于个人贡献和技术难点的英文追问回答。"
    ],
    "created_at": "2026-06-07T03:05:00Z"
  }
}
```

重复调用同一 Session 的生成接口会覆盖旧报告并返回最新报告。

### curl 示例

```bash
curl -X POST http://localhost:8080/api/v1/sessions/1/report
```

## 查询课后报告

```http
GET /api/v1/sessions/:id/report
```

### 成功响应

HTTP 状态码：`200`

响应结构同生成接口。

### 错误响应

未生成报告：

```json
{
  "code": 5003,
  "message": "report not found"
}
```

### curl 示例

```bash
curl http://localhost:8080/api/v1/sessions/1/report
curl http://localhost:8080/api/v1/sessions/999/report
curl http://localhost:8080/api/v1/sessions/abc/report
```

## 完整联调示例

以下示例假设服务刚启动，内存 ID 从 `1` 开始。实际联调时以前一个响应中的 `data.session_id` 为准。

```bash
curl -X POST http://localhost:8080/api/v1/sessions \
  -H "Content-Type: application/json" \
  -d '{"scenario_id":1}'

curl -X POST http://localhost:8080/api/v1/sessions/1/messages \
  -H "Content-Type: application/json" \
  -d '{"content":"I am study computer science and I have did a project."}'

curl -X POST http://localhost:8080/api/v1/sessions/1/finish

curl -X POST http://localhost:8080/api/v1/sessions/1/report

curl http://localhost:8080/api/v1/sessions/1/report
```

## 前端使用建议

- 用户点击“结束训练”后，先调用 `POST /api/v1/sessions/:id/finish`。
- 结束成功后调用 `POST /api/v1/sessions/:id/report` 生成报告。
- 报告页刷新时调用 `GET /api/v1/sessions/:id/report`。
- 收到 `404 / 5003` 时展示“报告尚未生成”，可提供生成按钮。
- 收到 `409 / 5002` 时提示用户先结束训练。
- 收到 `409 / 5004` 时提示“暂无足够反馈数据”，通常需要用户先完成至少一轮消息练习。
- `scores.total_score` 和顶层 `total_score` 当前一致，列表卡片可直接使用顶层字段。
- `frequent_errors`、`better_expressions`、`next_practice_plan` 都应兼容空数组。
- `memory` 模式下服务重启后 not found 属于预期开发行为；`mysql` 模式下应优先检查报告是否已生成并落库。

## 验证命令

```bash
go test ./...
```
