import { describe, expect, it, vi } from "vitest";
import { ApiError, type BackendHistoryInsights, type BackendScenario, type BackendSessionDetail } from "./client";
import { loadHistoryInsights, loadReportState, loadTrainingSessionState, sendTrainingAudio } from "./loaders";

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

const backendInsights: BackendHistoryInsights = {
  summary: {
    days: 7,
    total_sessions: 1,
    finished_sessions: 1,
    running_sessions: 0,
    scored_sessions: 1,
    generated_reports: 1,
    average_score: 77,
    previous_average_score: null,
    score_delta: null,
  },
  score_trend: [],
  scenario_trends: [],
  frequent_errors: [],
  next_recommendation: null,
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

  it("uploads audio and reloads the training state with the returned next goal", async () => {
    const file = new File(["audio"], "answer.webm", { type: "audio/webm" });
    const client = {
      uploadAudioMessage: vi.fn(async () => ({
        transcript: "I am study computer science and I have did a project.",
        user_message: {
          id: 10,
          session_id: 7,
          role: "user" as const,
          content: "I am study computer science and I have did a project.",
          stage: "自我介绍",
          created_at: "2026-06-07T03:00:00Z",
        },
        ai_message: {
          id: 11,
          session_id: 7,
          role: "ai" as const,
          content: "Could you explain your role?",
          stage: "项目经历",
          created_at: "2026-06-07T03:00:01Z",
        },
        stage: "项目经历",
        next_goal: "ask project details",
        turn_count: 1,
        correction_summary: { has_errors: true, error_count: 2 },
        score_summary: { total_score: 77, grammar: 72, expression: 80 },
      })),
      getSession: vi.fn(async () => ({
        ...session,
        turn_count: 1,
        messages: [
          {
            id: 10,
            session_id: 7,
            role: "user" as const,
            content: "I am study computer science and I have did a project.",
            stage: "自我介绍",
            created_at: "2026-06-07T03:00:00Z",
          },
        ],
      })),
      getScenario: vi.fn(async () => scenario),
      listSessionCorrections: vi.fn(async () => []),
      getSessionScore: vi.fn(async () => {
        throw new ApiError("score not found", 4003, 404);
      }),
    };

    const result = await sendTrainingAudio(7, file, client);

    expect(client.uploadAudioMessage).toHaveBeenCalledWith(7, file);
    expect(result.result.transcript).toBe("I am study computer science and I have did a project.");
    expect(result.session.coachSummary).toBe("ask project details");
  });

  it("loads history insights for the selected window", async () => {
    const client = {
      getHistoryInsights: vi.fn(async () => backendInsights),
    };

    const result = await loadHistoryInsights(7, client);

    expect(client.getHistoryInsights).toHaveBeenCalledWith(7);
    expect(result.summary.days).toBe(7);
  });
});
