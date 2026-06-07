import type {
  BackendCorrectionError,
  BackendCorrectionResult,
  BackendHistoryItem,
  BackendMessage,
  BackendReport,
  BackendScenario,
  BackendScenarioSummary,
  BackendScoreResult,
  BackendSessionDetail,
} from "./client";
import type { Correction, HistoryRecord, Scenario, ScoreDimension, TrainingReport, TrainingSession, TrainingTask } from "../types";
import { scoreTone } from "../utils/format";

const defaultScoreDescriptions: Record<ScoreDimension["key"], string> = {
  fluency: "表达是否连贯，回答是否能自然展开。",
  grammar: "时态、主谓一致和句子结构是否准确。",
  expression: "表达是否自然、得体，是否符合场景。",
  vocabulary: "是否使用了丰富且贴合场景的词汇。",
  completion: "是否完成当前场景的核心任务。",
};

const scoreNames: Record<ScoreDimension["key"], string> = {
  fluency: "流利度",
  grammar: "语法准确度",
  expression: "表达自然度",
  vocabulary: "词汇丰富度",
  completion: "场景完成度",
};

const difficultyLabels: Record<string, string> = {
  easy: "简单",
  medium: "中等",
  hard: "困难",
};

export function formatDurationSeconds(seconds: number) {
  const safeSeconds = Math.max(0, Math.floor(seconds));
  const minutes = Math.floor(safeSeconds / 60);
  const remainingSeconds = safeSeconds % 60;

  return `${String(minutes).padStart(2, "0")}:${String(remainingSeconds).padStart(2, "0")}`;
}

function formatDateTime(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }

  return date.toLocaleString("zh-CN", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  });
}

function formatMessageTime(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }

  return date.toLocaleTimeString("zh-CN", {
    hour: "2-digit",
    minute: "2-digit",
  });
}

function secondsBetween(start: string, end: string | null, now: Date) {
  const startDate = new Date(start);
  const endDate = end ? new Date(end) : now;
  if (Number.isNaN(startDate.getTime()) || Number.isNaN(endDate.getTime())) {
    return 0;
  }

  return Math.max(0, Math.floor((endDate.getTime() - startDate.getTime()) / 1000));
}

function mapScenarioBase(scenario: BackendScenarioSummary | BackendScenario): Scenario {
  const detail = scenario as Partial<BackendScenario>;
  const stages = detail.stages ?? [];
  const rubric = detail.rubric ?? [];

  return {
    id: scenario.id,
    code: scenario.code,
    name: scenario.name,
    englishName: scenario.code,
    description: scenario.description,
    difficulty: scenario.difficulty,
    difficultyLabel: difficultyLabels[scenario.difficulty] ?? scenario.difficulty,
    aiRole: detail.ai_role ?? "AI 教练",
    userGoal: detail.user_goal ?? scenario.description,
    openingMessage: detail.opening_message ?? "",
    goals: stages.length ? stages.map((stage) => stage.description) : [scenario.description],
    stages,
    rubric,
    sessionId: "",
  };
}

export function mapScenarioSummary(scenario: BackendScenarioSummary) {
  return mapScenarioBase(scenario);
}

export function mapScenarioDetail(scenario: BackendScenario) {
  return mapScenarioBase(scenario);
}

function mapMessage(message: BackendMessage, scenario: Scenario) {
  const isUser = message.role === "user";

  return {
    id: message.id,
    role: message.role,
    speaker: isUser ? "You" : `AI ${scenario.aiRole}`,
    content: message.content,
    stage: message.stage,
    createdAt: formatMessageTime(message.created_at),
  };
}

export function mapSessionScore(score?: BackendScoreResult | null): ScoreDimension[] {
  const values: Record<ScoreDimension["key"], number> = {
    fluency: score?.fluency ?? 0,
    grammar: score?.grammar ?? 0,
    expression: score?.expression ?? 0,
    vocabulary: score?.vocabulary ?? 0,
    completion: score?.completion ?? 0,
  };

  return (Object.keys(scoreNames) as ScoreDimension["key"][]).map((key) => ({
    key,
    name: scoreNames[key],
    score: values[key],
    description: score?.comment || defaultScoreDescriptions[key],
  }));
}

function toCorrectionCategory(type: BackendCorrectionError["type"]): Correction["category"] {
  if (type === "grammar" || type === "vocabulary" || type === "expression") {
    return type;
  }

  return "expression";
}

export function mapCorrections(corrections: BackendCorrectionResult[]): Correction[] {
  return corrections.flatMap((correction) => {
    if (correction.errors.length === 0) {
      return [
        {
          title: "表达优化建议",
          category: "expression",
          original: correction.original_text,
          suggestion: correction.corrected_text || correction.better_expressions[0] || correction.original_text,
          explanation: "AI 给出了更自然的表达方式。",
          issues: correction.better_expressions,
        },
      ];
    }

    return correction.errors.map((error) => ({
      title: error.explanation || `${error.span} 表达建议`,
      category: toCorrectionCategory(error.type),
      original: error.span || correction.original_text,
      suggestion: error.suggestion || correction.corrected_text,
      explanation: error.explanation,
      issues: error.span && error.suggestion ? [`${error.span} -> ${error.suggestion}`] : correction.better_expressions,
    }));
  });
}

function deriveTasks(scenario: Scenario, currentStage: string, status: BackendSessionDetail["status"]): TrainingTask[] {
  if (scenario.stages.length === 0) {
    return [{ label: currentStage, status: status === "finished" ? "done" : "active" }];
  }

  const activeIndex = Math.max(
    0,
    scenario.stages.findIndex((stage) => stage.name === currentStage),
  );

  return scenario.stages.map((stage, index) => {
    let taskStatus: TrainingTask["status"] = "pending";
    if (status === "finished" || index < activeIndex) {
      taskStatus = "done";
    } else if (index === activeIndex) {
      taskStatus = "active";
    }

    return {
      label: stage.name,
      status: taskStatus,
    };
  });
}

function deriveProgress(tasks: TrainingSession["tasks"]) {
  if (tasks.length === 0) {
    return 0;
  }

  const doneCount = tasks.filter((task) => task.status === "done").length;
  return Math.round((doneCount / tasks.length) * 100);
}

function latestBetterExpression(corrections: BackendCorrectionResult[]) {
  for (let index = corrections.length - 1; index >= 0; index -= 1) {
    const expressions = corrections[index].better_expressions;
    const expression = expressions[expressions.length - 1];
    if (expression) {
      return expression;
    }
  }

  return "完成一轮输入后，这里会显示更自然的替代表达。";
}

export function mapSessionDetailToTrainingSession({
  session,
  scenario,
  corrections = [],
  score = null,
  nextGoal,
  now = new Date(),
}: {
  session: BackendSessionDetail;
  scenario?: BackendScenario;
  corrections?: BackendCorrectionResult[];
  score?: BackendScoreResult | null;
  nextGoal?: string;
  now?: Date;
}): TrainingSession {
  const uiScenario = scenario ? mapScenarioDetail(scenario) : mapScenarioSummary(session.scenario);
  const latestMessage = session.messages[session.messages.length - 1];
  const currentStage = latestMessage?.stage || uiScenario.stages[0]?.name || "准备开始";
  const messages = session.messages.map((message) => mapMessage(message, uiScenario));

  if (messages.length === 0 && uiScenario.openingMessage) {
    messages.push({
      id: 0,
      role: "ai",
      speaker: `AI ${uiScenario.aiRole}`,
      content: uiScenario.openingMessage,
      stage: currentStage,
      createdAt: "刚刚",
    });
  }

  const tasks = deriveTasks(uiScenario, currentStage, session.status);
  const scoreDimensions = mapSessionScore(score);

  return {
    sessionId: String(session.session_id),
    sessionNo: session.session_no,
    scenario: uiScenario,
    status: session.status,
    liveStatus: session.status === "running" ? "普通 JSON 对话" : "训练已结束",
    durationLabel: formatDurationSeconds(secondsBetween(session.created_at, session.ended_at, now)),
    turnCount: session.turn_count,
    progress: deriveProgress(tasks),
    currentStage,
    voiceStatus: "idle",
    tasks,
    focusTags: uiScenario.rubric.length ? uiScenario.rubric.map((item) => item.name) : ["表达完整性", "语法准确性", "场景完成度"],
    messages,
    currentScore: score?.total_score ?? 0,
    coachSummary: nextGoal || score?.comment || "发送第一条文本后，AI 会给出纠错和评分反馈。",
    scores: scoreDimensions,
    corrections: mapCorrections(corrections),
    naturalExpression: latestBetterExpression(corrections),
  };
}

function mapReportScenario(report: BackendReport): Scenario {
  return {
    ...mapScenarioSummary({
      id: report.scenario.id,
      code: report.scenario.code,
      name: report.scenario.name,
      description: `${report.scenario.name}训练报告`,
      difficulty: report.scenario.difficulty,
    }),
    sessionId: String(report.session_id),
  };
}

function parseFrequentError(error: string, index: number): Correction {
  const [rawOriginal, rawSuggestion] = error.split("->").map((part) => part.trim());

  return {
    title: `高频错误 ${index + 1}`,
    category: "grammar",
    original: rawOriginal || error,
    suggestion: rawSuggestion || error,
    explanation: error,
    issues: [error],
  };
}

function stringArray(value: string[] | null | undefined) {
  return Array.isArray(value) ? value : [];
}

export function mapReport(report: BackendReport): TrainingReport {
  const majorProblems = stringArray(report.major_problems);
  const frequentErrors = stringArray(report.frequent_errors);
  const betterExpressions = stringArray(report.better_expressions);
  const nextPracticePlan = stringArray(report.next_practice_plan);

  return {
    sessionId: String(report.session_id),
    scenario: mapReportScenario(report),
    durationLabel: formatDurationSeconds(report.duration_seconds),
    turnCount: report.turn_count,
    issueCount: frequentErrors.length,
    completionRate: report.scores.completion,
    totalScore: report.total_score,
    grade: `${scoreTone(report.total_score)} / 100`,
    summary: report.summary,
    scores: mapSessionScore(report.scores),
    majorProblems,
    frequentErrors: frequentErrors.map(parseFrequentError),
    betterExpressions: betterExpressions.map((expression) => ({
      before: "原表达见纠错项",
      after: expression,
    })),
    nextPracticePlan: nextPracticePlan.map((item, index) => ({
      title: `练习建议 ${index + 1}`,
      description: item,
    })),
    createdAt: formatDateTime(report.created_at),
  };
}

export function mapHistoryRecord(record: BackendHistoryItem): HistoryRecord {
  return {
    sessionId: String(record.session_id),
    sessionNo: record.session_no,
    scenario: mapScenarioSummary(record.scenario),
    status: record.status,
    score: record.total_score ?? 0,
    trainedAt: formatDateTime(record.created_at),
    durationLabel: record.ended_at ? formatDurationSeconds(secondsBetween(record.created_at, record.ended_at, new Date())) : "进行中",
    turnCount: record.turn_count,
    majorProblem: record.total_score === null ? "暂无评分" : record.status === "finished" ? "已完成训练" : "训练进行中",
    reportStatus: record.report_status,
  };
}
