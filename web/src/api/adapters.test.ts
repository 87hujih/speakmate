import { describe, expect, it } from "vitest";
import {
  mapCorrections,
  mapHistoryRecord,
  mapReport,
  mapScenarioDetail,
  mapSessionDetailToTrainingSession,
  mapSessionScore,
} from "./adapters";
import type {
  BackendCorrectionResult,
  BackendHistoryItem,
  BackendReport,
  BackendScenario,
  BackendSessionDetail,
  BackendScoreResult,
} from "./client";

const scenarioDetail: BackendScenario = {
  id: 1,
  code: "interview",
  name: "英语面试",
  description: "练习自我介绍、项目经历和技术追问",
  difficulty: "medium",
  ai_role: "技术面试官",
  user_goal: "清晰介绍自己的背景和项目经历。",
  opening_message: "Could you introduce yourself?",
  stages: [
    { name: "自我介绍", description: "介绍背景。" },
    { name: "项目经历", description: "说明项目。" },
  ],
  rubric: [{ name: "语法准确度", description: "检查时态和句式。" }],
};

const score: BackendScoreResult = {
  message_id: 10,
  session_id: 7,
  fluency: 75,
  grammar: 72,
  expression: 80,
  vocabulary: 76,
  completion: 85,
  total_score: 77,
  comment: "语法需要加强。",
};

describe("api adapters", () => {
  it("maps backend scenario detail into UI scenario fields", () => {
    const scenario = mapScenarioDetail(scenarioDetail);

    expect(scenario.aiRole).toBe("技术面试官");
    expect(scenario.difficultyLabel).toBe("中等");
    expect(scenario.goals).toEqual(["介绍背景。", "说明项目。"]);
    expect(scenario.sessionId).toBe("");
  });

  it("maps session detail into a training session with opening message and derived tasks", () => {
    const sessionDetail: BackendSessionDetail = {
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

    const session = mapSessionDetailToTrainingSession({
      session: sessionDetail,
      scenario: scenarioDetail,
      corrections: [],
      score,
      now: new Date("2026-06-07T03:02:30Z"),
    });

    expect(session.sessionId).toBe("7");
    expect(session.messages).toHaveLength(1);
    expect(session.messages[0].content).toBe("Could you introduce yourself?");
    expect(session.currentStage).toBe("自我介绍");
    expect(session.tasks.map((task) => task.status)).toEqual(["active", "pending"]);
    expect(session.durationLabel).toBe("02:30");
    expect(session.currentScore).toBe(77);
  });

  it("flattens backend correction errors for the feedback panel", () => {
    const corrections: BackendCorrectionResult[] = [
      {
        message_id: 10,
        session_id: 7,
        original_text: "I am study computer science.",
        corrected_text: "I am studying computer science.",
        errors: [
          {
            type: "grammar",
            span: "am study",
            suggestion: "am studying",
            explanation: "be 动词后应接现在分词。",
          },
        ],
        better_expressions: ["I major in computer science."],
      },
    ];

    const mapped = mapCorrections(corrections);

    expect(mapped[0]).toMatchObject({
      title: "be 动词后应接现在分词。",
      category: "grammar",
      original: "am study",
      suggestion: "am studying",
    });
    expect(mapped[0].issues).toEqual(["am study -> am studying"]);
  });

  it("maps null correction arrays to empty feedback instead of throwing", () => {
    const sessionDetail: BackendSessionDetail = {
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
      turn_count: 1,
      messages: [
        {
          id: 10,
          session_id: 7,
          role: "user",
          content: "I study computer science.",
          stage: "自我介绍",
          created_at: "2026-06-07T03:01:00Z",
        },
      ],
      created_at: "2026-06-07T03:00:00Z",
      ended_at: null,
    };
    const corrections = [
      {
        message_id: 10,
        session_id: 7,
        original_text: "I study computer science.",
        corrected_text: "I study computer science.",
        errors: null,
        better_expressions: null,
      },
    ] as unknown as BackendCorrectionResult[];

    const session = mapSessionDetailToTrainingSession({
      session: sessionDetail,
      scenario: scenarioDetail,
      corrections,
      now: new Date("2026-06-07T03:02:00Z"),
    });

    expect(session.corrections).toEqual([
      {
        title: "表达优化建议",
        category: "expression",
        original: "I study computer science.",
        suggestion: "I study computer science.",
        explanation: "AI 给出了更自然的表达方式。",
        issues: [],
      },
    ]);
    expect(session.naturalExpression).toBe("完成一轮输入后，这里会显示更自然的替代表达。");
  });

  it("orders mapped corrections with the latest user message first for realtime feedback", () => {
    const sessionDetail: BackendSessionDetail = {
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
      turn_count: 2,
      messages: [
        {
          id: 10,
          session_id: 7,
          role: "user",
          content: "I am study computer science.",
          stage: "自我介绍",
          created_at: "2026-06-07T03:01:00Z",
        },
        {
          id: 12,
          session_id: 7,
          role: "user",
          content: "I have did a project.",
          stage: "项目经历",
          created_at: "2026-06-07T03:02:00Z",
        },
      ],
      created_at: "2026-06-07T03:00:00Z",
      ended_at: null,
    };
    const corrections: BackendCorrectionResult[] = [
      {
        message_id: 10,
        session_id: 7,
        original_text: "I am study computer science.",
        corrected_text: "I am studying computer science.",
        errors: [
          {
            type: "grammar",
            span: "am study",
            suggestion: "am studying",
            explanation: "be 动词后应接现在分词。",
          },
        ],
        better_expressions: ["I major in computer science."],
      },
      {
        message_id: 12,
        session_id: 7,
        original_text: "I have did a project.",
        corrected_text: "I have done a project.",
        errors: [
          {
            type: "grammar",
            span: "have did",
            suggestion: "have done",
            explanation: "现在完成时中 have 后应接过去分词。",
          },
        ],
        better_expressions: ["I completed a project that solved a real user problem."],
      },
    ];

    const session = mapSessionDetailToTrainingSession({
      session: sessionDetail,
      scenario: scenarioDetail,
      corrections,
      now: new Date("2026-06-07T03:03:00Z"),
    });

    expect(session.corrections[0]).toMatchObject({
      original: "have did",
      suggestion: "have done",
    });
    expect(session.naturalExpression).toBe("I completed a project that solved a real user problem.");
  });

  it("maps backend scores to all five UI dimensions", () => {
    expect(mapSessionScore(score).map((item) => [item.key, item.score])).toEqual([
      ["fluency", 75],
      ["grammar", 72],
      ["expression", 80],
      ["vocabulary", 76],
      ["completion", 85],
    ]);
  });

  it("maps backend report string arrays into report card structures", () => {
    const report: BackendReport = {
      session_id: 7,
      scenario: { id: 1, code: "interview", name: "英语面试", difficulty: "medium" },
      duration_seconds: 180,
      turn_count: 1,
      total_score: 77,
      scores: score,
      summary: "本次训练完成 1 轮。",
      major_problems: ["语法准确度需要加强。"],
      frequent_errors: ["am study -> am studying"],
      better_expressions: ["I major in computer science."],
      next_practice_plan: ["用 STAR 结构重写项目经历。"],
      created_at: "2026-06-07T03:05:00Z",
    };

    const mapped = mapReport(report);

    expect(mapped.sessionId).toBe("7");
    expect(mapped.durationLabel).toBe("03:00");
    expect(mapped.issueCount).toBe(1);
    expect(mapped.frequentErrors[0].suggestion).toBe("am studying");
    expect(mapped.betterExpressions[0].after).toBe("I major in computer science.");
    expect(mapped.nextPracticePlan[0].title).toBe("练习建议 1");
  });

  it("maps evidence-rich backend report strings into readable report cards", () => {
    const report: BackendReport = {
      session_id: 7,
      scenario: { id: 1, code: "interview", name: "英语面试", difficulty: "medium" },
      duration_seconds: 180,
      turn_count: 1,
      total_score: 77,
      scores: score,
      summary: "本次训练完成 1 轮。",
      major_problems: ["语法准确度需要加强。证据：am study -> am studying"],
      frequent_errors: ["am study -> am studying | 原因：be 动词后应接现在分词。 | 证据：“I am study computer science.”"],
      better_expressions: ["I am study computer science. -> I major in computer science."],
      next_practice_plan: [
        "任务：把 “am study” 改写为 “am studying” 并各造 2 个新句子 | 验收：新句子必须正确使用 “am studying”。",
      ],
      created_at: "2026-06-07T03:05:00Z",
    };

    const mapped = mapReport(report);

    expect(mapped.frequentErrors[0]).toMatchObject({
      original: "am study",
      suggestion: "am studying",
      explanation: "原因：be 动词后应接现在分词。 | 证据：“I am study computer science.”",
    });
    expect(mapped.betterExpressions[0]).toEqual({
      before: "I am study computer science.",
      after: "I major in computer science.",
    });
    expect(mapped.nextPracticePlan[0]).toEqual({
      title: "把 “am study” 改写为 “am studying” 并各造 2 个新句子",
      description: "验收：新句子必须正确使用 “am studying”。",
    });
  });

  it("maps null backend report arrays to empty UI arrays", () => {
    const report: BackendReport = {
      session_id: 7,
      scenario: { id: 1, code: "interview", name: "英语面试", difficulty: "medium" },
      duration_seconds: 180,
      turn_count: 1,
      total_score: 77,
      scores: score,
      summary: "本次训练完成 1 轮。",
      major_problems: null,
      frequent_errors: null,
      better_expressions: null,
      next_practice_plan: null,
      created_at: "2026-06-07T03:05:00Z",
    };

    const mapped = mapReport(report);

    expect(mapped.issueCount).toBe(0);
    expect(mapped.majorProblems).toEqual([]);
    expect(mapped.frequentErrors).toEqual([]);
    expect(mapped.betterExpressions).toEqual([]);
    expect(mapped.nextPracticePlan).toEqual([]);
  });

  it("maps history records with missing score and running state", () => {
    const record: BackendHistoryItem = {
      session_id: 7,
      session_no: "S202606070001",
      user_id: 1,
      scenario: {
        id: 1,
        code: "interview",
        name: "英语面试",
        description: "练习自我介绍、项目经历和技术追问",
        difficulty: "medium",
      },
      status: "running",
      turn_count: 0,
      total_score: null,
      report_status: "not_generated",
      created_at: "2026-06-07T03:00:00Z",
      ended_at: null,
    };

    const mapped = mapHistoryRecord(record);

    expect(mapped.sessionId).toBe("7");
    expect(mapped.score).toBe(0);
    expect(mapped.majorProblem).toBe("暂无评分");
    expect(mapped.durationLabel).toBe("进行中");
  });
});
