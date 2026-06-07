import { describe, expect, it, vi } from "vitest";
import { ApiError, type BackendScenario, type BackendSessionDetail } from "./client";
import { loadReportState, loadTrainingSessionState } from "./loaders";

const scenario: BackendScenario = {
  id: 1,
  code: "interview",
  name: "英语面试",
  description: "练习自我介绍、项目经历和技术追问",
  difficulty: "medium",
  ai_role: "技术面试官",
  user_goal: "清晰介绍自己的背景和项目经历。",
  opening_message: "Could you introduce yourself?",
  stages: [{ name: "自我介绍", description: "介绍背景。" }],
  rubric: [{ name: "语法准确度", description: "检查时态。" }],
};

const session: BackendSessionDetail = {
  session_id: 7,
  session_no: "S202606070001",
  scenario: {
    id: 1,
    code: "interview",
    name: "英语面试",
    description: "练习自我介绍、项目经历和技术追问",
    difficulty: "medium",
  },
  status: "running",
  turn_count: 0,
  messages: [],
  created_at: "2026-06-07T03:00:00Z",
  ended_at: null,
};

describe("api loaders", () => {
  it("loads a training session while treating missing feedback as empty state", async () => {
    const client = {
      getSession: vi.fn(async () => session),
      getScenario: vi.fn(async () => scenario),
      listSessionCorrections: vi.fn(async () => {
        throw new ApiError("correction not found", 4002, 404);
      }),
      getSessionScore: vi.fn(async () => {
        throw new ApiError("score not found", 4003, 404);
      }),
    };

    const result = await loadTrainingSessionState(7, client);

    expect(result.session.sessionId).toBe("7");
    expect(result.session.corrections).toEqual([]);
    expect(result.session.currentScore).toBe(0);
    expect(client.getScenario).toHaveBeenCalledWith(1);
  });

  it("returns a missing report state when the backend has no report yet", async () => {
    const client = {
      getReport: vi.fn(async () => {
        throw new ApiError("report not found", 5003, 404);
      }),
    };

    const result = await loadReportState(7, client);

    expect(result.status).toBe("missing");
  });
});
