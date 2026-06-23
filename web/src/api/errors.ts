import { ApiError } from "./client";

/** demoErrorMessage 将后端错误码转换为面向 Demo 用户的中文提示。 */
export function demoErrorMessage(error: unknown, fallback: string) {
  if (!(error instanceof Error)) {
    return fallback;
  }
  if (!(error instanceof ApiError)) {
    return error.message || fallback;
  }

  switch (error.code) {
    case 3003:
      return "真实 AI 回复失败，请检查 LLM_API_KEY、LLM_BASE_URL、LLM_MODEL 和模型服务可用性后重试。";
    case 3004:
      return "AI 纠错或评分生成失败，请检查反馈模型配置，或开启 FEEDBACK_FAIL_OPEN 后重试。";
    case 5002:
      return "请先结束训练，再生成课后报告。";
    case 5004:
      return "暂无足够反馈数据，请至少完成一轮输入并等待纠错评分生成后再生成报告。";
    case 5005:
      return "课后报告生成失败，请检查 LLM_API_KEY、LLM_BASE_URL、LLM_MODEL，或开启 SUMMARY_USE_MOCK 后重试。";
    case 7005:
      return "语音识别失败，请检查 ASR 配置、API Key 或浏览器录音格式后重试。";
    case 8001:
      return "训练短期状态服务不可用，请检查 Redis 配置或关闭 REDIS_ENABLED 后重试。";
    case 8002:
      return "实时事件发布失败，请检查 Redis 连接或暂时关闭 REDIS_ENABLED 后重试。";
    case 9002:
      return "请求超时，请稍后重试或检查外部服务配置。";
    case 9003:
      return "请求内容过大，请缩短输入或压缩音频后重试。";
    case 9004:
      return "请求过于频繁，请稍等后再试。";
    default:
      return error.message || fallback;
  }
}
