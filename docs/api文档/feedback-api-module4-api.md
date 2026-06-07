# Module 4: Feedback API 接口文档

此文件保留为旧入口，最新反馈 API 契约已迁移到 [feedback-api.md](feedback-api.md)。

当前 Module 4 第一版已完成：

- `POST /api/v1/sessions/:id/messages` 会同步生成纠错摘要和评分摘要；
- `GET /api/v1/messages/:message_id/corrections` 查询单条消息纠错；
- `GET /api/v1/sessions/:id/corrections` 查询整场训练纠错列表；
- `GET /api/v1/sessions/:id/scores` 查询当前 Session 评分；
- 默认使用 Mock Agent，配置完整且关闭 Mock 后可使用 OpenAI-compatible LLM；
- 默认 `FEEDBACK_FAIL_OPEN=true`，反馈失败不阻断主对话链路。

请以 [feedback-api.md](feedback-api.md) 为前端联调和后续开发的正式文档。
