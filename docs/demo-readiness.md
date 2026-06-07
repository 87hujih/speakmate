# SpeakMate AI Demo Readiness

## 固定 Demo Flow

目标时长：3 到 5 分钟。

```text
打开前端
  -> 选择英语面试
  -> 创建训练
  -> 输入或录音一句带错误的英文
  -> AI 追问
  -> 查看纠错
  -> 查看评分
  -> 结束训练
  -> 生成报告
  -> 查看历史记录
```

推荐演示输入：

```text
I am study computer science and I have did a project about robot control.
```

## Mock 模式演示

Mock 模式不依赖 API Key、MySQL 或 Redis，适合现场兜底。

后端：

```powershell
$env:STORAGE_MODE="memory"
$env:REDIS_ENABLED="false"
$env:LLM_USE_MOCK="true"
$env:LLM_FALLBACK_TO_MOCK="true"
$env:CORRECTION_USE_MOCK="true"
$env:SCORING_USE_MOCK="true"
$env:SUMMARY_USE_MOCK="true"
$env:ASR_USE_MOCK="true"
go run ./cmd/server
```

前端：

```powershell
cd web
$env:VITE_API_BASE_URL="http://localhost:8080/api/v1"
npm run dev
```

API 脚本验证：

```powershell
powershell -ExecutionPolicy Bypass -File scripts/demo-mock.ps1
```

前端演示步骤：

1. 打开 `http://localhost:5173`。
2. 点击“开始英语面试训练”。
3. 输入推荐 Demo 句子并发送。
4. 观察 AI 追问、右侧纠错、综合评分和分项评分。
5. 点击结束训练，进入报告页。
6. 点击生成课后报告，查看总评、高频错误、表达升级和练习建议。
7. 打开历史记录，确认本次训练和报告状态可见。

## 真实 LLM / ASR / Redis 联调

真实服务模式依赖本机或部署环境变量。不要把真实密钥写进仓库。

```powershell
$env:STORAGE_MODE="mysql"
$env:MYSQL_DSN="speakmate:speakmate@tcp(127.0.0.1:3306)/speakmate?parseTime=true&loc=UTC"
$env:REDIS_ENABLED="true"
$env:REDIS_ADDR="127.0.0.1:6379"
$env:LLM_USE_MOCK="false"
$env:LLM_PROVIDER="openai-compatible"
$env:LLM_BASE_URL="https://your-provider.example.com/v1"
$env:LLM_API_KEY="your-api-key"
$env:LLM_MODEL="your-model"
$env:LLM_FALLBACK_TO_MOCK="true"
$env:ASR_USE_MOCK="false"
$env:ASR_PROVIDER="tencent"
$env:TENCENT_ASR_APP_ID="your-app-id"
$env:TENCENT_ASR_SECRET_ID="your-secret-id"
$env:TENCENT_ASR_SECRET_KEY="your-secret-key"
```

联调前启动依赖并执行迁移：

```powershell
docker compose up -d mysql redis
go run ./cmd/migrate -dir migrations
go run ./cmd/server
```

文本链路脚本：

```powershell
powershell -ExecutionPolicy Bypass -File scripts/demo-real-services.ps1
```

带 ASR 音频文件：

```powershell
powershell -ExecutionPolicy Bypass -File scripts/demo-real-services.ps1 -AudioFile .\answer.ogg -AudioContentType audio/ogg
```

SSE 验证：

```powershell
curl.exe -N http://localhost:8080/api/v1/sessions/<session_id>/stream
```

另开终端发送消息，观察 `ai_message_delta`、`ai_message_done`、`correction_done`、`score_updated`。

## MySQL 持久化验收

1. 启动 MySQL 并执行 migration。
2. 设置 `STORAGE_MODE=mysql` 和 `MYSQL_DSN`。
3. 完成一次 Demo Flow 并生成报告。
4. 停止并重新启动后端，不清空 MySQL volume。
5. 打开历史记录，确认训练记录仍在。
6. 打开报告页，确认报告仍可查询。

## Redis 验收

`REDIS_ENABLED=true` 时后端启动会 ping Redis，失败会直接启动失败。Demo 中可以验证：

- 训练发送消息后，SSE 能收到事件；
- Redis 中存在 `session:{id}:state`、`session:{id}:messages`、`session:{id}:events`；
- WebSocket 录音后，Redis 中存在 `ws:{session_id}:connection`；
- Redis 短期 key 过期不影响 MySQL 历史记录和报告。

## 讲解重点

- SpeakMate 是场景训练，不是开放闲聊。
- Mock 模式保证现场无 Key 也能完整演示。
- 真实 LLM streaming、ASR、Redis 都是配置启用，不硬编码真实服务返回。
- MySQL 保存长期 Session、Message、Correction、Score、Report；Redis 只保存短期状态和事件。
- 错误路径有明确提示：API Key 错、ASR 失败、Redis 不可用、请求超时、请求过大或过频繁。

## Demo 兜底策略

- 无 API Key：保持 `LLM_USE_MOCK=true`、`ASR_USE_MOCK=true`。
- LLM Key 错误但要继续演示：设置 `LLM_FALLBACK_TO_MOCK=true`。
- ASR 真实模式失败：改用文本输入或切回 `ASR_USE_MOCK=true`。
- Redis 不可用：本地演示可设置 `REDIS_ENABLED=false`；真实 Redis 联调时保持失败即显式报错。
- MySQL 不可用：Mock 演示可使用 `STORAGE_MODE=memory`，持久化验收必须使用 MySQL。
