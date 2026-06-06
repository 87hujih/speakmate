
# Module 1: Scenario API 接口文档

本文档说明当前已实现的训练场景查询接口。前端可以用它完成场景选择页和场景详情页的数据加载。

## 基本信息

| 项目 | 说明 |
|---|---|
| API 前缀 | `/api/v1` |
| 响应格式 | JSON |
| 成功响应结构 | `{ "code": 0, "message": "success", "data": ... }` |
| 错误响应结构 | `{ "code": 业务错误码, "message": "错误说明" }` |
| 当前数据来源 | 后端内存内置数据 |
| 是否需要登录 | 当前版本不需要 |

## 错误码

| HTTP 状态码 | 业务错误码 | message | 说明 |
|---|---:|---|---|
| `400` | `1002` | `invalid scenario id` | `:id` 不是合法正整数 |
| `404` | `1001` | `scenario not found` | 场景不存在 |

## 场景列表

获取所有可训练场景。用于前端场景选择页。

```http
GET /api/v1/scenarios
```

### 请求参数

无 query 参数，无 request body。

### 成功响应

HTTP 状态码：`200`

`data` 是场景摘要数组。列表接口只返回前端选择场景所需的基础字段。

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | number | 场景 ID |
| `code` | string | 场景编码，后续可用于选择 Agent 或 Mock 回复策略 |
| `name` | string | 场景名称 |
| `description` | string | 场景简介 |
| `difficulty` | string | 难度，当前为 `easy` 或 `medium` |

示例：

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "code": "interview",
      "name": "英语面试",
      "description": "练习自我介绍、项目经历和技术追问",
      "difficulty": "medium"
    },
    {
      "id": 2,
      "code": "restaurant",
      "name": "餐厅点餐",
      "description": "练习询问菜单、表达偏好和处理特殊需求",
      "difficulty": "easy"
    },
    {
      "id": 3,
      "code": "meeting",
      "name": "工作会议",
      "description": "练习表达观点、澄清问题和总结结论",
      "difficulty": "medium"
    }
  ]
}
```

### curl 示例

```bash
curl http://localhost:8080/api/v1/scenarios
```

## 场景详情

获取单个训练场景的完整信息。用于前端进入训练前展示 AI 角色、训练目标、开场白、训练阶段和评分维度。

```http
GET /api/v1/scenarios/:id
```

### 路径参数

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `id` | number | 是 | 场景 ID，必须是正整数 |

### 成功响应

HTTP 状态码：`200`

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | number | 场景 ID |
| `code` | string | 场景编码 |
| `name` | string | 场景名称 |
| `description` | string | 场景简介 |
| `difficulty` | string | 难度 |
| `ai_role` | string | AI 在该场景中扮演的角色 |
| `user_goal` | string | 用户在该场景中要完成的训练目标 |
| `opening_message` | string | AI 开场白，后续创建训练 Session 时可复用 |
| `stages` | array | 训练阶段列表 |
| `stages[].name` | string | 阶段名称 |
| `stages[].description` | string | 阶段说明 |
| `rubric` | array | 评分维度列表 |
| `rubric[].name` | string | 评分维度名称 |
| `rubric[].description` | string | 评分维度说明 |

示例：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "code": "interview",
    "name": "英语面试",
    "description": "练习自我介绍、项目经历和技术追问",
    "difficulty": "medium",
    "ai_role": "技术面试官",
    "user_goal": "清晰介绍自己的背景、项目经历和协作方式，并能回答常见技术追问。",
    "opening_message": "Hello, welcome to the interview. Could you start by briefly introducing yourself and telling me about one project you are proud of?",
    "stages": [
      {
        "name": "自我介绍",
        "description": "介绍教育背景、技术方向和当前目标。"
      },
      {
        "name": "项目经历",
        "description": "说明项目目标、个人职责和关键成果。"
      },
      {
        "name": "技术追问",
        "description": "回答实现细节、技术选择和问题排查过程。"
      },
      {
        "name": "结尾提问",
        "description": "向面试官提出关于团队、岗位或项目的问题。"
      }
    ],
    "rubric": [
      {
        "name": "流利度",
        "description": "回答是否连贯，是否能自然展开说明。"
      },
      {
        "name": "语法准确度",
        "description": "时态、主谓一致和句子结构是否准确。"
      },
      {
        "name": "表达自然度",
        "description": "是否使用面试场景中自然、得体的表达。"
      },
      {
        "name": "场景完成度",
        "description": "是否完成自我介绍、项目说明和反问环节。"
      }
    ]
  }
}
```

### 错误响应

#### 场景 ID 非法

请求：

```http
GET /api/v1/scenarios/abc
```

响应：

```json
{
  "code": 1002,
  "message": "invalid scenario id"
}
```

#### 场景不存在

请求：

```http
GET /api/v1/scenarios/999
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
curl http://localhost:8080/api/v1/scenarios/1
curl http://localhost:8080/api/v1/scenarios/999
curl http://localhost:8080/api/v1/scenarios/abc
```

## 前端使用建议

- 场景选择页调用 `GET /api/v1/scenarios` 渲染卡片或列表。
- 用户点击场景后，使用对应 `id` 调用 `GET /api/v1/scenarios/:id`。
- `opening_message` 后续可作为训练开始时 AI 的第一条消息。
- `stages` 可用于展示训练流程或进度提示。
- `rubric` 可用于展示本场景的评分标准。
- 前端应按 HTTP 状态码区分错误类型，并展示 `message`。

## 当前内置场景

| id | code | name | difficulty | ai_role |
|---:|---|---|---|---|
| `1` | `interview` | 英语面试 | `medium` | 技术面试官 |
| `2` | `restaurant` | 餐厅点餐 | `easy` | 餐厅服务员 |
| `3` | `meeting` | 工作会议 | `medium` | 项目同事 |

## 验证命令

启动服务：

```bash
go run ./cmd/server
```

接口验证：

```bash
curl http://localhost:8080/api/v1/scenarios
curl http://localhost:8080/api/v1/scenarios/1
curl http://localhost:8080/api/v1/scenarios/999
curl http://localhost:8080/api/v1/scenarios/abc
```

自动化测试：

```bash
go test ./internal/router
```
