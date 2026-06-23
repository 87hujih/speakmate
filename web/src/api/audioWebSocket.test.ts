import { describe, expect, it } from "vitest";
import { createAudioWebSocketUrl, parseAudioWebSocketEvent } from "./audioWebSocket";

describe("audio websocket api helpers", () => {
  it("builds ws URLs from relative and absolute API bases", () => {
    expect(createAudioWebSocketUrl(7, "/api/v1", "http://localhost:5173")).toBe(
      "ws://localhost:5173/api/v1/sessions/7/audio/ws",
    );
    expect(createAudioWebSocketUrl(7, "https://api.example.test/api/v1")).toBe(
      "wss://api.example.test/api/v1/sessions/7/audio/ws",
    );
  });

  it("flattens partial, final, correction and error event payloads", () => {
    expect(
      parseAudioWebSocketEvent(
        JSON.stringify({
          type: "partial_transcript",
          session_id: 7,
          payload: { transcript: "I am study", sequence: 1 },
        }),
      ),
    ).toEqual({ type: "partial_transcript", transcript: "I am study", sequence: 1 });

    expect(
      parseAudioWebSocketEvent(
        JSON.stringify({
          type: "final_transcript",
          session_id: 7,
          payload: {
            transcript: "I am study computer science and I have did a project.",
            user_message: {
              id: 1,
              session_id: 7,
              role: "user",
              content: "I am study computer science and I have did a project.",
              stage: "自我介绍",
              created_at: "2026-06-07T03:00:00Z",
            },
            ai_message: {
              id: 2,
              session_id: 7,
              role: "ai",
              content: "Could you explain your role?",
              stage: "项目经历",
              created_at: "2026-06-07T03:00:01Z",
            },
            stage: "项目经历",
            next_goal: "ask project details",
            turn_count: 1,
          },
        }),
      ),
    ).toMatchObject({
      type: "final_transcript",
      transcript: "I am study computer science and I have did a project.",
      turn_count: 1,
    });

    expect(
      parseAudioWebSocketEvent(JSON.stringify({ type: "correction", payload: { has_errors: true, error_count: 2 } })),
    ).toEqual({ type: "correction", has_errors: true, error_count: 2 });

    expect(parseAudioWebSocketEvent(JSON.stringify({ type: "error", payload: { message: "请上传音频文件" } }))).toEqual({
      type: "error",
      message: "请上传音频文件",
    });
  });
});
