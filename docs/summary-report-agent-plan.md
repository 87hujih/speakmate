# 课后报告与 Summary Agent 实施计划

## Goal

基于一次训练中的场景信息、Session 生命周期、消息历史、纠错结果和当前评分，生成结构化课后报告，补齐文本训练学习闭环：

```text
选择场景 -> 创建训练 -> 对话练习 -> 纠错评分 -> 结束训练 -> 生成课后报告 -> 重复查询报告
```

## Data Dependencies

报告生成依赖以下数据：

| 数据 | 来源 | 用途 |
|---|---|---|
| `scenario` | Scenario Service | 报告展示场景摘要，并约束 Summary Agent 按场景目标总结 |
| `session` | Session Repository | 校验训练存在和状态，读取轮次、起止时间和消息历史 |
| `messages` | Session.Messages | 汇总用户表达、AI 追问和对话阶段 |
| `corrections` | Feedback Repository | 提取主要问题、高频错误和更好表达 |
| `score` | Feedback Repository | 展示分项评分、综合分和评分理由 |

第一版要求已结束训练才能生成报告。查询已生成报告不要求重新调用 Summary Agent。

## Report Model

新增 `model.Report`，字段如下：

| 字段 | 类型 | 说明 |
|---|---|---|
| `session_id` | `int` | 报告所属 Session |
| `scenario` | `ReportScenario` | 场景摘要，包含 `id/code/name/difficulty` |
| `duration_seconds` | `int` | `ended_at - created_at`，缺失或异常时为 `0` |
| `turn_count` | `int` | 训练轮次 |
| `total_score` | `int` | 当前综合分 |
| `scores` | `ScoreResult` | 当前分项评分 |
| `summary` | `string` | 总评 |
| `major_problems` | `[]string` | 主要问题 |
| `frequent_errors` | `[]string` | 高频错误 |
| `better_expressions` | `[]string` | 更自然表达 |
| `next_practice_plan` | `[]string` | 下一步练习建议 |
| `created_at` | `time.Time` | 报告生成时间 |

数组字段返回空数组而不是 `null`。

## Repository

新增 Report Repository 接口：

```go
type ReportRepository interface {
    Save(report model.Report) error
    FindBySessionID(sessionID int) (model.Report, error)
}
```

内存实现以 `session_id` 为唯一键，`Save` 支持覆盖。`FindBySessionID` 在缺失时返回 `repository.ErrReportNotFound`。Repository 测试覆盖：

- 保存并查询；
- 覆盖同一 `session_id` 报告；
- 查询不存在报告；
- 保存后修改原始切片不会污染仓库数据；
- 查询结果被调用方修改不会污染仓库数据。

## Summary Agent

新增 Summary Agent 抽象：

```go
type SummaryAgent interface {
    Summarize(input SummaryInput) (SummaryOutput, error)
}
```

`SummaryInput` 包含：

- `Scenario`
- `Session`
- `Messages`
- `Corrections`
- `Score`

`SummaryOutput` 包含：

- `Summary`
- `MajorProblems`
- `FrequentErrors`
- `BetterExpressions`
- `NextPracticePlan`
- `Raw`

### Mock Summary Agent

Mock 输出必须稳定，并包含：

- 总评；
- 高频错误；
- 更好表达；
- 下一步练习建议。

Mock 会优先从纠错结果提取错误和更好表达，没有纠错数据时返回通用建议，自动测试不请求真实 LLM。

### LLM Summary Agent

新增 OpenAI-compatible LLM Summary Agent，复用现有 `llm.Client`：

- 配置完整、全局 Mock 关闭、`SUMMARY_USE_MOCK=false` 时才启用 LLM；
- 配置缺失或 provider 不支持时默认使用 Mock；
- LLM 调用、JSON 解析或字段校验失败时 fallback 到 Mock；
- Agent 单测使用 Fake client，不请求真实模型。

## Prompt Contract

Summary Prompt 约束：

- 只输出合法 JSON；
- 不输出 Markdown；
- 不虚构训练中不存在的消息、评分或错误；
- 反馈要具体、可执行，面向英语口语练习；
- 高频错误必须来自纠错结果或对纠错结果的归纳；
- 下一步练习建议要和当前场景目标相关。

LLM JSON 结构：

```json
{
  "summary": "本次训练总评",
  "major_problems": ["主要问题"],
  "frequent_errors": ["高频错误"],
  "better_expressions": ["更好表达"],
  "next_practice_plan": ["下一步练习建议"]
}
```

所有字符串字段必须非空，数组字段必须存在。

## Service

新增 Report Service：

- `GenerateReport(sessionID int, ctx context.Context) (model.Report, error)`
- `GetReport(sessionID int) (model.Report, error)`

生成流程：

1. 校验 `session_id` 为正整数；
2. 查询 Session，不存在返回 `ErrSessionNotFound`；
3. 要求 Session 已结束，否则返回 `ErrSessionNotFinished`；
4. 查询 Scenario；
5. 查询 Session 纠错列表，缺失返回 `ErrReportFeedbackMissing`；
6. 查询当前评分，缺失返回 `ErrReportFeedbackMissing`；
7. 调用 Summary Agent；
8. 组装 Report 并保存，重复生成覆盖旧报告；
9. 返回保存后的 Report。

查询流程：

1. 校验 `session_id`；
2. 按 `session_id` 查询报告；
3. 未生成报告返回 `ErrReportNotFound`。

## API

新增路由：

```http
POST /api/v1/sessions/:id/report
GET  /api/v1/sessions/:id/report
```

成功响应复用统一结构：

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
      "difficulty": "intermediate"
    },
    "duration_seconds": 120,
    "turn_count": 2,
    "total_score": 77,
    "scores": {},
    "summary": "...",
    "major_problems": [],
    "frequent_errors": [],
    "better_expressions": [],
    "next_practice_plan": [],
    "created_at": "2026-06-07T03:00:00Z"
  }
}
```

## Error Codes

| HTTP | code | message | 场景 |
|---:|---:|---|---|
| 400 | 5001 | `invalid report request` | 路径 ID 非正整数 |
| 404 | 2003 | `session not found` | Session 不存在 |
| 409 | 5002 | `session not finished` | Session 尚未结束 |
| 404 | 5003 | `report not found` | 查询时报告尚未生成 |
| 409 | 5004 | `report feedback missing` | 缺少纠错或评分数据 |
| 502 | 5005 | `summary agent failed` | Summary Agent 失败且无 fallback |
| 500 | 500 | `internal server error` | 未预期错误 |

## Tests

### Repository

- 保存并查询；
- 覆盖；
- 不存在报告；
- 切片 clone。

### Agent

- Mock 输出稳定且包含总评、错误、表达和计划；
- Prompt 包含场景、消息、纠错和评分；
- LLM Agent 解析合法 JSON；
- LLM Agent 拒绝非法 JSON 或缺失字段；
- LLM Agent 失败时 fallback 到 Mock。

### Service

- 已结束训练可以生成报告；
- 重复生成覆盖旧报告；
- 查询已生成报告；
- 非法 session id；
- Session 不存在；
- Session 未结束；
- 缺失纠错；
- 缺失评分；
- Summary Agent 失败。

### Handler / Router

- `POST /api/v1/sessions/:id/report` 正常生成；
- `GET /api/v1/sessions/:id/report` 正常查询；
- 非法 ID；
- 不存在 Session；
- 未生成报告查询；
- 缺失反馈数据。

## Done Definition

- 用户可以结束训练后生成课后报告；
- 报告可以重复查询；
- 报告内容包含评分、纠错、高频错误和下一步建议；
- Mock/Fake 测试不依赖真实 LLM；
- 可用完整测试命令通过；
- 普通 `go test ./...` 若仍受本地 Go cache 权限限制，需要在任务记录和最终说明中明确记录。
