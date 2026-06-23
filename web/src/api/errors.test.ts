import { describe, expect, it } from "vitest";
import { ApiError } from "./client";
import { demoErrorMessage } from "./errors";

describe("demoErrorMessage", () => {
  it("turns real LLM failures into actionable configuration guidance", () => {
    const message = demoErrorMessage(new ApiError("对话 AI 回复失败", 3003, 502), "消息发送失败");

    expect(message).toContain("真实 AI 回复失败");
    expect(message).toContain("LLM");
    expect(message).toContain("LLM_API_KEY");
    expect(message).not.toContain("LLM_FALLBACK_TO_MOCK");
    expect(message).not.toContain("mock");
  });

  it("explains report preconditions in user-facing Chinese", () => {
    expect(demoErrorMessage(new ApiError("训练尚未结束", 5002, 409), "报告生成失败")).toBe(
      "请先结束训练，再生成课后报告。",
    );
    expect(demoErrorMessage(new ApiError("报告缺少反馈数据", 5004, 409), "报告生成失败")).toBe(
      "暂无足够反馈数据，请至少完成一轮输入并等待纠错评分生成后再生成报告。",
    );
  });

  it("maps infrastructure and abuse-control errors to clear recovery steps", () => {
    expect(demoErrorMessage(new ApiError("实时事件发布失败", 8002, 503), "请求失败")).toContain("Redis");
    expect(demoErrorMessage(new ApiError("请求超时", 9002, 504), "请求失败")).toBe("请求超时，请稍后重试或检查外部服务配置。");
    expect(demoErrorMessage(new ApiError("请求过于频繁", 9004, 429), "请求失败")).toBe("请求过于频繁，请稍等后再试。");
  });

  it("falls back to the original error message or provided default", () => {
    expect(demoErrorMessage(new ApiError("自定义后端错误", 9999, 500), "请求失败")).toBe("自定义后端错误");
    expect(demoErrorMessage("非错误对象", "请求失败")).toBe("请求失败");
  });
});
