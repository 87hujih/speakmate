# Module 4: Feedback API 接口文档

本文档说明当前已实现的反馈查询接口。前端可以用它查询单条消息纠错、整场训练纠错列表和当前 Session 评分。

当前版本只实现查询 API。反馈数据来自内存 Feedback Repository，服务重启后数据会丢失。消息发送流程尚未把 Correction Agent 和 Scoring Agent 接入持久保存链路，因此刚启动服务直接查询通常会返回 `correction not found` 或 `score not found`。后续完成反馈生成链路后，这些接口会返回真实纠错和评分数据。

## 基本信息

| 项目 | 说明 |
|---|---|
| API 前缀 | `/api/v1` |
| 响应格式 | JSON |
| 成功响应结构 | `{ "code": 0, "message": "success", "data": ... }` |
| 错误响应结构 | `{ "code": 业务错误码, "message": "错误说明" }` |
| 当前数据来源 | 后端内存数据 |
| 是否需要登录 | 当前版本不需要 |
| 前置条件 | 需要后端已经保存对应纠错或评分结果 |

## 错误码

| HTTP 状态码 | 业务错误码 | message | 说明 |
|---|---:|---|---|
| `400` | `2002` | `invalid session id` | 路径参数 `:session_id` 不是合法正整数 |
| `400` | `4001` | `invalid feedback request` | 路径参数 `:message_id` 不是合法正整数，或反馈查询参数非法 |
| `404` | `4002` | `correction not found` | 没有找到对应消息或 Session 的纠错结果 |
| `404` | `4003` | `score not found` | 没有找到对应 Session 的当前评分 |
| `500` | `500` | `internal server error` | 非预期服务端错误 |

## 纠错结果字段

单条纠错和纠错列表中的元素都使用同一结构。

| 字段 | 类型 | 说明 |
|---|---|---|
| `message_id` | number | 被纠错的用户消息 ID |
| `session_id` | number | 所属 Session ID |
| `original_text` | string | 用户原始表达 |
| `corrected_text` | string | 推荐修正后的表达 |
| `errors` | array | 具体错误列表 |
| `errors[].type` | string | 错误类型，可选值见下表 |
| `errors[].span` | string | 原句中的问题片段 |
| `errors[].suggestion` | string | 推荐替换表达 |
| `errors[].explanation` | string | 中文解释 |
| `better_expressions` | array | 更自然或更贴合场景的推荐表达 |

错误类型：

| type | 说明 |
|---|---|
| `grammar` | 语法错误 |
| `vocabulary` | 用词不准确 |
| `expression` | 表达不自然 |
| `structure` | 句子结构问题 |
| `scenario` | 不符合当前场景 |

## 评分结果字段

| 字段 | 类型 | 说明 |
|---|---|---|
| `message_id` | number | 评分对应的最近一条用户消息 ID |
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

## 获取单条消息纠错结果

### 接口路径

```http
GET /api/v1/messages/:message_id/corrections
```

### 请求参数

路径参数：

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `message_id` | number | 是 | 用户消息 ID，必须是正整数 |

请求体：无。

Query 参数：无。

### 成功响应字段

HTTP 状态码：`200`

`data` 是单条纠错结果，字段见 [纠错结果字段](#纠错结果字段)。

示例响应：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "message_id": 10,
    "session_id": 1,
    "original_text": "I am study computer science and I have did a project.",
    "corrected_text": "I am studying computer science, and I have done a project.",
    "errors": [
      {
        "type": "grammar",
        "span": "am study",
        "suggestion": "am studying",
        "explanation": "be 动词后应接现在分词。"
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

Message ID 非法：

```http
GET /api/v1/messages/abc/corrections
```

```json
{
  "code": 4001,
  "message": "invalid feedback request"
}
```

纠错结果不存在：

```http
GET /api/v1/messages/999/corrections
```

```json
{
  "code": 4002,
  "message": "correction not found"
}
```

### curl 示例

```bash
curl http://localhost:8080/api/v1/messages/10/corrections
curl http://localhost:8080/api/v1/messages/999/corrections
curl http://localhost:8080/api/v1/messages/abc/corrections
```

## 获取某次训练的全部纠错结果

### 接口路径

```http
GET /api/v1/sessions/:session_id/corrections
```

### 请求参数

路径参数：

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `session_id` | number | 是 | Session ID，必须是正整数 |

请求体：无。

Query 参数：无。

### 成功响应字段

HTTP 状态码：`200`

`data` 是纠错结果数组。数组元素字段见 [纠错结果字段](#纠错结果字段)。

示例响应：

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "message_id": 10,
      "session_id": 1,
      "original_text": "I am study computer science.",
      "corrected_text": "I am studying computer science.",
      "errors": [
        {
          "type": "grammar",
          "span": "am study",
          "suggestion": "am studying",
          "explanation": "be 动词后应接现在分词。"
        }
      ],
      "better_expressions": [
        "I major in computer science."
      ]
    },
    {
      "message_id": 12,
      "session_id": 1,
      "original_text": "I worked on a project.",
      "corrected_text": "I worked on a project.",
      "errors": [],
      "better_expressions": [
        "I contributed to a robotics project."
      ]
    }
  ]
}
```

### 错误响应

Session ID 非法：

```http
GET /api/v1/sessions/abc/corrections
```

```json
{
  "code": 2002,
  "message": "invalid session id"
}
```

纠错结果不存在：

```http
GET /api/v1/sessions/999/corrections
```

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

### 接口路径

```http
GET /api/v1/sessions/:session_id/scores
```

### 请求参数

路径参数：

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `session_id` | number | 是 | Session ID，必须是正整数 |

请求体：无。

Query 参数：无。

### 成功响应字段

HTTP 状态码：`200`

`data` 是当前 Session 的评分结果，字段见 [评分结果字段](#评分结果字段)。

示例响应：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "message_id": 12,
    "session_id": 1,
    "fluency": 75,
    "grammar": 72,
    "expression": 80,
    "vocabulary": 76,
    "completion": 85,
    "total_score": 78,
    "comment": "用户能够表达核心意思，但存在时态和动词形式错误。"
  }
}
```

### 错误响应

Session ID 非法：

```http
GET /api/v1/sessions/abc/scores
```

```json
{
  "code": 2002,
  "message": "invalid session id"
}
```

评分结果不存在：

```http
GET /api/v1/sessions/999/scores
```

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

## 前端使用建议

- 训练页右侧反馈面板可以调用 `GET /api/v1/sessions/:session_id/corrections` 渲染累计纠错列表。
- 用户点击某条用户消息时，可以调用 `GET /api/v1/messages/:message_id/corrections` 展示单条详细纠错。
- 评分卡片可以调用 `GET /api/v1/sessions/:session_id/scores` 展示当前综合分和五个分项分。
- 当前版本查询不到反馈时会返回 `404 / 4002` 或 `404 / 4003`。前端应显示空状态，例如“本轮反馈生成中”或“暂无评分”，不要当成页面级错误。
- 对 `400 / 2002` 和 `400 / 4001`，前端应检查路由参数或本地保存的 ID，避免传入非数字、空字符串或 `0`。
- `errors` 可能为空数组，表示当前表达没有明显错误。前端不要用 `errors.length === 0` 判断接口失败。
- `better_expressions` 可能为空数组，前端应兼容不展示推荐表达区域。
- `total_score` 是后端计算后的整数，前端不要重复计算综合分，只负责展示。
- 当前数据是内存存储，服务重启后查询可能返回 not found。开发联调时不要把它当成用户数据丢失问题。
- 后续消息发送接口接入反馈生成后，可以在 `POST /api/v1/sessions/:id/messages` 成功后再拉取 corrections 和 scores，或等待前端统一轮询策略。

## 推荐调用流程

```text
1. POST /api/v1/sessions
2. POST /api/v1/sessions/:id/messages
3. 将 user_message 和 ai_message 追加到前端消息列表
4. GET /api/v1/messages/:message_id/corrections
5. GET /api/v1/sessions/:session_id/corrections
6. GET /api/v1/sessions/:session_id/scores
7. 如果返回 not found，展示暂无反馈或生成中状态
```

## 验证命令

启动服务：

```bash
go run ./cmd/server
```

接口验证：

```bash
curl http://localhost:8080/api/v1/messages/1/corrections
curl http://localhost:8080/api/v1/sessions/1/corrections
curl http://localhost:8080/api/v1/sessions/1/scores
```

自动化测试：

```bash
go test ./...
```
