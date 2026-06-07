# SpeakMate AI Troubleshooting

## 快速定位

先确认三个入口：

```powershell
go test ./...
cd web
npm test
npm run build
```

服务健康检查：

```powershell
curl.exe http://localhost:8080/health
```

前端 API 地址：

```powershell
$env:VITE_API_BASE_URL="http://localhost:8080/api/v1"
```

## 后端启动失败

### `mysql dsn required`

原因：`STORAGE_MODE=mysql` 但 `MYSQL_DSN` 为空。

处理：

```powershell
$env:MYSQL_DSN="speakmate:speakmate@tcp(127.0.0.1:3306)/speakmate?parseTime=true&loc=UTC"
```

或切回本地 Mock：

```powershell
$env:STORAGE_MODE="memory"
```

### `redis addr required` 或 Redis unavailable

原因：`REDIS_ENABLED=true` 但 Redis 地址缺失或不可连接。

处理：

```powershell
docker compose up -d redis
$env:REDIS_ADDR="127.0.0.1:6379"
```

本地演示兜底：

```powershell
$env:REDIS_ENABLED="false"
```

## API Key 和真实 LLM

### 无 API Key

Mock 演示应保持：

```powershell
$env:LLM_USE_MOCK="true"
$env:LLM_FALLBACK_TO_MOCK="true"
$env:CORRECTION_USE_MOCK="true"
$env:SCORING_USE_MOCK="true"
$env:SUMMARY_USE_MOCK="true"
```

### API Key 错误或模型地址错误

常见前端提示：

```text
AI 回复服务调用失败，请检查 LLM_API_KEY、LLM_BASE_URL、LLM_MODEL，或开启 LLM_FALLBACK_TO_MOCK 后重试。
```

处理：

- 确认 `LLM_BASE_URL` 是 OpenAI-compatible `/v1` 地址；
- 确认 `LLM_API_KEY` 没有复制空格；
- 确认 `LLM_MODEL` 是供应商支持的模型；
- 现场演示需要兜底时设置 `LLM_FALLBACK_TO_MOCK=true`。

## ASR 问题

### 语音识别失败

常见前端提示：

```text
语音识别失败，请检查浏览器录音格式或后端 ASR 配置后重试。
```

Mock 演示：

```powershell
$env:ASR_USE_MOCK="true"
$env:ASR_PROVIDER="mock"
```

腾讯云真实模式需要：

```powershell
$env:ASR_USE_MOCK="false"
$env:ASR_PROVIDER="tencent"
$env:TENCENT_ASR_APP_ID="your-app-id"
$env:TENCENT_ASR_SECRET_ID="your-secret-id"
$env:TENCENT_ASR_SECRET_KEY="your-secret-key"
$env:TENCENT_ASR_ENGINE_TYPE="16k_en"
```

当前腾讯云真实识别建议使用 `audio/ogg`、`audio/mp4`、`audio/wav`。`audio/webm` 在真实腾讯云模式下可能失败，本地 Demo 可使用 Mock ASR。

## MySQL 持久化

初始化：

```powershell
docker compose up -d mysql
$env:MYSQL_DSN="speakmate:speakmate@tcp(127.0.0.1:3306)/speakmate?parseTime=true&loc=UTC"
go run ./cmd/migrate -dir migrations
```

验证：

1. `STORAGE_MODE=mysql` 启动后端；
2. 完成一次训练并生成报告；
3. 重启后端；
4. 打开历史页和报告页确认数据仍在。

如果历史为空，检查是否误用了 `STORAGE_MODE=memory` 或清空了 Compose volume。

## Redis 和 SSE / WebSocket

Redis 只保存短期状态和事件，不保存历史报告。

常见错误：

```text
session state store failed
stream event publish failed
```

处理：

- 检查 `REDIS_ENABLED`；
- 检查 `REDIS_ADDR`；
- 检查 Redis 容器健康状态；
- 本地 Mock 演示可设置 `REDIS_ENABLED=false`。

SSE 连接：

```powershell
curl.exe -N http://localhost:8080/api/v1/sessions/<session_id>/stream
```

WebSocket 被拒绝时先检查 Origin 是否在 `CORS_ALLOWED_ORIGINS` 中。浏览器前端默认需要：

```env
CORS_ALLOWED_ORIGINS=http://localhost:5173,http://127.0.0.1:5173
```

## CORS

前端请求被浏览器拦截时，确认：

- `VITE_API_BASE_URL` 指向后端 `/api/v1`；
- 后端 `CORS_ALLOWED_ORIGINS` 包含前端 Origin；
- 如果前后端同源通过 Nginx 代理，前端可使用 `VITE_API_BASE_URL=/api/v1`。

## Timeout、请求体大小和限流

默认安全配置：

```env
REQUEST_TIMEOUT_SECONDS=30
REQUEST_BODY_LIMIT_BYTES=12582912
RATE_LIMIT_REQUESTS=120
RATE_LIMIT_WINDOW_SECONDS=60
```

错误含义：

| HTTP | code | message | 处理 |
|---|---:|---|---|
| 504 | 9002 | request timeout | 检查外部 LLM / ASR 响应时间，或调大 timeout |
| 413 | 9003 | request body too large | 缩短文本、压缩音频，或调大 body limit |
| 429 | 9004 | rate limit exceeded | 稍后重试，或调大限流窗口 |

SSE 和 WebSocket 不套普通请求超时，但仍受 CORS / Origin 和限流保护。

## Docker Compose

启动完整环境：

```powershell
docker compose up --build
```

访问：

```text
前端：http://localhost:5173
后端：http://localhost:8080/health
```

如果 `backend` 未启动，按顺序检查：

1. `mysql` 是否 healthy；
2. `migrate` 是否成功结束；
3. `redis` 是否 healthy；
4. 后端日志是否出现 MySQL / Redis 配置错误。

如果本机 `3306`、`6379`、`8080` 或 `5173` 已被占用，可以只修改宿主端口：

```powershell
$env:COMPOSE_MYSQL_PORT="33306"
$env:COMPOSE_REDIS_PORT="36379"
$env:BACKEND_HOST_PORT="18080"
$env:FRONTEND_HOST_PORT="15173"
docker compose up --build
```

此时访问：

```text
前端：http://localhost:15173
后端：http://localhost:18080/health
```

清空本地数据：

```powershell
docker compose down -v
```

## 密钥安全

- `.env` 和 `.env.*` 已被 `.gitignore` 忽略；
- `.env.example` 只保留空值和示例；
- 日志会脱敏常见 `api_key`、`token`、`password`、`secret`、Bearer token 和 MySQL DSN 密码；
- 不要把真实 Key 写进 README、测试输出、截图或 issue。
