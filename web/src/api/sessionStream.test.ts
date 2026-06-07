import { describe, expect, it } from "vitest";
import { createSessionStreamUrl, parseSessionStreamEvent } from "./sessionStream";

describe("session stream helpers", () => {
  it("builds stream URLs from the API base URL", () => {
    expect(createSessionStreamUrl(7, "/api/v1")).toBe("/api/v1/sessions/7/stream");
    expect(createSessionStreamUrl(7, "http://localhost:8080/api/v1/")).toBe("http://localhost:8080/api/v1/sessions/7/stream");
  });

  it("flattens backend event payloads", () => {
    expect(
      parseSessionStreamEvent(
        "ai_message_delta",
        JSON.stringify({ type: "ai_message_delta", session_id: 7, payload: { message_id: 0, delta: "What " } }),
      ),
    ).toEqual({ type: "ai_message_delta", message_id: 0, content: "What " });

    expect(
      parseSessionStreamEvent(
        "ai_message_done",
        JSON.stringify({
          type: "ai_message_done",
          session_id: 7,
          payload: { message_id: 11, content: "What was your role?", stage: "项目经历" },
        }),
      ),
    ).toEqual({ type: "ai_message_done", message_id: 11, content: "What was your role?", stage: "项目经历" });

    expect(
      parseSessionStreamEvent(
        "error",
        JSON.stringify({ type: "error", session_id: 7, payload: { code: "conversation_agent_failed", message: "conversation agent failed" } }),
      ),
    ).toEqual({ type: "error", code: "conversation_agent_failed", message: "conversation agent failed" });
  });
});
