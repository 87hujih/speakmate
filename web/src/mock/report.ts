import type { BetterExpression, PracticePlanItem, TrainingReport } from "../types";
import { scenarios } from "./scenarios";
import { corrections, interviewSession, scoreDimensions } from "./training";

const betterExpressions: BetterExpression[] = [
  {
    before: "I made a robot project.",
    after: "I built a robotics project focused on motion control.",
  },
  {
    before: "My work is make the backend service.",
    after: "I was responsible for building the backend service and integrating it with Redis and MySQL.",
  },
];

const nextPracticePlan: PracticePlanItem[] = [
  {
    title: "用 STAR 法重答项目经历",
    description: "从 Situation、Task、Action、Result 四个维度组织回答。",
  },
  {
    title: "重点练习时态",
    description: "复盘 have done、worked on、built 等项目表达句式。",
  },
  {
    title: "增加个人贡献",
    description: "用 I was responsible for 明确说明你的角色和产出。",
  },
  {
    title: "减少泛词使用",
    description: "少用 good、thing、do，替换为具体动词和技术动作。",
  },
];

export const interviewReport: TrainingReport = {
  sessionId: interviewSession.sessionId,
  scenario: scenarios[0],
  durationLabel: "04:32",
  turnCount: 8,
  issueCount: 6,
  completionRate: 85,
  totalScore: 78,
  grade: "Good / 100",
  summary: "本次训练重点覆盖项目经历表达、时态准确性和技术追问回答，整体表达可被理解。",
  scores: scoreDimensions,
  majorProblems: ["语法准确度需要加强，优先检查动词形式和时态。", "项目贡献描述偏短，需要补充个人职责和结果。"],
  frequentErrors: corrections.slice(0, 2),
  betterExpressions,
  nextPracticePlan,
  createdAt: "2026-06-07 14:30",
};

export const reports: Record<string, TrainingReport> = {
  [interviewReport.sessionId]: interviewReport,
  [scenarios[1].sessionId]: {
    ...interviewReport,
    sessionId: scenarios[1].sessionId,
    scenario: scenarios[1],
    totalScore: 84,
    grade: "Great / 100",
    summary: "餐厅点餐流程完整，礼貌表达基本自然，可继续补充特殊需求表达。",
  },
  [scenarios[2].sessionId]: {
    ...interviewReport,
    sessionId: scenarios[2].sessionId,
    scenario: scenarios[2],
    totalScore: 81,
    grade: "Good / 100",
    summary: "工作会议表达有清晰主线，但观点展开和例子支撑还可以更充分。",
  },
};

export function getReport(sessionId: string) {
  return reports[sessionId] ?? interviewReport;
}
