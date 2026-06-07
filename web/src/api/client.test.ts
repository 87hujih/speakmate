import { beforeEach, describe, expect, it, vi } from "vitest";
import { ApiError, apiClient } from "./client";

describe("apiClient", () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it("posts a text message and unwraps the API data payload", async () => {
    const fetchMock = vi.fn(async () =>
      new Response(
        JSON.stringify({
          code: 0,
          message: "success",
          data: {
            user_message: {
              id: 1,
              session_id: 7,
              role: "user",
              content: "hello",
              stage: "自我介绍",
              created_at: "2026-06-07T03:00:00Z",
            },
            ai_message: {
              id: 2,
              session_id: 7,
              role: "ai",
              content: "Nice to meet you.",
              stage: "项目经历",
              created_at: "2026-06-07T03:00:01Z",
            },
            stage: "项目经历",
            next_goal: "ask project details",
            turn_count: 1,
            correction_summary: { has_errors: false, error_count: 0 },
            score_summary: { total_score: 86, grammar: 88, expression: 84 },
          },
        }),
      ),
    );
    vi.stubGlobal("fetch", fetchMock);

    const result = await apiClient.sendTextMessage(7, "hello");

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/sessions/7/messages",
      expect.objectContaining({
        method: "POST",
        body: JSON.stringify({ content: "hello" }),
      }),
    );
    expect(result.turn_count).toBe(1);
  });

  it("throws ApiError with backend code and message", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => new Response(JSON.stringify({ code: 3002, message: "message content is required" }), { status: 400 })),
    );

    await expect(apiClient.sendTextMessage(7, " ")).rejects.toMatchObject({
      code: 3002,
      message: "message content is required",
      status: 400,
    });
  });
});
