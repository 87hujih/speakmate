import type { HistoryRecord } from "../types";
import { interviewScenario, scenarios } from "./scenarios";

export const historyRecords: HistoryRecord[] = [
  {
    sessionId: interviewScenario.sessionId,
    sessionNo: "S202606070001",
    scenario: interviewScenario,
    status: "finished",
    score: 78,
    trainedAt: "2026-06-05 20:30",
    durationLabel: "04:32",
    turnCount: 8,
    majorProblem: "时态错误、项目表达不够具体",
    reportStatus: "generated",
  },
  {
    sessionId: scenarios[1].sessionId,
    sessionNo: "S202606070002",
    scenario: scenarios[1],
    status: "finished",
    score: 84,
    trainedAt: "2026-06-06 10:12",
    durationLabel: "05:18",
    turnCount: 7,
    majorProblem: "礼貌表达可以更自然",
    reportStatus: "generated",
  },
  {
    sessionId: scenarios[2].sessionId,
    sessionNo: "S202606070003",
    scenario: scenarios[2],
    status: "finished",
    score: 81,
    trainedAt: "2026-06-06 21:45",
    durationLabel: "06:04",
    turnCount: 9,
    majorProblem: "观点表达较短，缺少例子支撑",
    reportStatus: "generated",
  },
];
