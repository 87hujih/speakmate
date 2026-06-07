import { describe, expect, it } from "vitest";
import { ApiError } from "./client";
import { demoErrorMessage } from "./errors";

describe("demoErrorMessage", () => {
  it("turns real LLM failures into actionable configuration guidance", () => {
    const message = demoErrorMessage(new ApiError("conversation agent failed", 3003, 502), "消息发送失败");

    expect(message).toContain("LLM");
    expect(message).toContain("LLM_API_KEY");
    expect(message).toContain("LLM_FALLBACK_TO_MOCK");
  });

  it("explains report preconditions in user-facing Chinese", () => {
    expect(demoErrorMessage(new ApiError("session not finished", 5002, 409), "报告生成失败")).toBe(
      "请先结束训练，再生成课后报告。",
    );
    expect(demoErrorMessage(new ApiError("report feedback missing", 5004, 409), "报告生成失败")).toBe(
      "暂无足够反馈数据，请至少完成一轮输入并等待纠错评分生成后再生成报告。",
    );
  });

  it("maps infrastructure and abuse-control errors to clear recovery steps", () => {
    expect(demoErrorMessage(new ApiError("stream event publish failed", 8002, 503), "请求失败")).toContain("Redis");
    expect(demoErrorMessage(new ApiError("request timeout", 9002, 504), "请求失败")).toBe("请求超时，请稍后重试或检查外部服务配置。");
    expect(demoErrorMessage(new ApiError("rate limit exceeded", 9004, 429), "请求失败")).toBe("请求过于频繁，请稍等后再试。");
  });

  it("falls back to the original error message or provided default", () => {
    expect(demoErrorMessage(new ApiError("custom backend error", 9999, 500), "请求失败")).toBe("custom backend error");
    expect(demoErrorMessage("not an error", "请求失败")).toBe("请求失败");
  });
});
