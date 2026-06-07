# History API 接口文档

本文档说明训练历史记录查询接口。前端可以用它展示历史训练列表，并进入单次训练详情或报告页。

## 基本信息

| 项目 | 说明 |
|---|---|
| API 前缀 | `/api/v1` |
| 响应格式 | JSON |
| 成功响应结构 | `{ "code": 0, "message": "success", "data": ... }` |
| 错误响应结构 | `{ "code": 业务错误码, "message": "错误说明" }` |
| 数据来源 | `STORAGE_MODE=memory` 时使用内存仓库；`STORAGE_MODE=mysql` 时使用 MySQL |
| 是否需要登录 | 当前版本不需要 |

## 分页规则

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|---|---|---|---:|---|
| `page` | number | 否 | `1` | 页码，传入时必须为正整数 |
| `page_size` | number | 否 | `20` | 每页条数，传入时必须为正整数，最大按 `100` 处理 |

分页响应统一返回：

| 字段 | 类型 | 说明 |
|---|---|---|
| `items` | array | 当前页历史记录 |
| `page` | number | 当前页码 |
| `page_size` | number | 当前每页条数 |
| `total` | number | 符合条件的总记录数 |

## 历史记录字段

| 字段 | 类型 | 说明 |
|---|---|---|
| `session_id` | number | Session 内部 ID |
| `session_no` | string | 展示或排查用编号 |
| `user_id` | number | 用户 ID |
| `scenario` | object | 场景摘要 |
| `scenario.id` | number | 场景 ID |
| `scenario.code` | string | 场景编码 |
| `scenario.name` | string | 场景名称 |
| `scenario.description` | string | 场景简介 |
| `scenario.difficulty` | string | 场景难度 |
| `status` | string | `running` 或 `finished` |
| `turn_count` | number | 已完成对话轮次 |
| `total_score` | number/null | 当前综合分；尚未评分时为 `null` |
| `report_status` | string | `generated` 或 `not_generated` |
| `created_at` | string | Session 创建时间，RFC3339 格式 |
| `ended_at` | string/null | Session 结束时间，未结束时为 `null` |

## 错误码

| HTTP 状态码 | 业务错误码 | message | 说明 |
|---|---:|---|---|
| `400` | `6001` | `invalid history request` | 分页参数非法，或 `user_id` 不是正整数 |
| `500` | `500` | `internal server error` | 非预期服务端错误 |

## 查询全部训练历史

```http
GET /api/v1/sessions?page=1&page_size=20
```

### 成功响应

HTTP 状态码：`200`

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "items": [
      {
        "session_id": 1,
        "session_no": "S202606070001",
        "user_id": 42,
        "scenario": {
          "id": 1,
          "code": "interview",
          "name": "英语面试",
          "description": "练习自我介绍、项目经历和技术追问",
          "difficulty": "medium"
        },
        "status": "finished",
        "turn_count": 1,
        "total_score": 77,
        "report_status": "generated",
        "created_at": "2026-06-07T03:00:00Z",
        "ended_at": "2026-06-07T03:05:00Z"
      }
    ],
    "page": 1,
    "page_size": 20,
    "total": 1
  }
}
```

空列表时 `items` 返回空数组，`total` 为 `0`。

### curl 示例

```bash
curl 'http://localhost:8080/api/v1/sessions?page=1&page_size=20'
curl 'http://localhost:8080/api/v1/sessions?page=0&page_size=20'
```

## 按用户查询训练历史

```http
GET /api/v1/users/:user_id/sessions?page=1&page_size=20
```

### 路径参数

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `user_id` | number | 是 | 用户 ID，必须为正整数 |

### 成功响应

响应结构同 `GET /api/v1/sessions`，但只返回指定用户的训练记录。

### curl 示例

```bash
curl 'http://localhost:8080/api/v1/users/42/sessions?page=1&page_size=20'
curl 'http://localhost:8080/api/v1/users/abc/sessions'
```

## 查看训练详情和报告

历史列表只返回摘要。进入详情页时继续调用：

```http
GET /api/v1/sessions/:id
GET /api/v1/sessions/:id/report
```

如果 `report_status` 是 `generated`，前端可以展示“查看报告”入口；如果是 `not_generated`，可以展示“生成报告”入口。

## 验证命令

```bash
go test ./...
```
