/** ScenarioCode 表示训练场景业务编码。 */
export type ScenarioCode = string;

/** Difficulty 表示训练场景难度。 */
export type Difficulty = string;

/** ScenarioStage 描述训练场景中的阶段。 */
export interface ScenarioStage {
  name: string;
  description: string;
}

/** ScenarioRubric 描述训练场景中的评分维度。 */
export interface ScenarioRubric {
  name: string;
  description: string;
}

/** Scenario 描述前端使用的训练场景模型。 */
export interface Scenario {
  id: number;
  code: ScenarioCode;
  name: string;
  englishName: string;
  description: string;
  difficulty: Difficulty;
  difficultyLabel: string;
  aiRole: string;
  userGoal: string;
  openingMessage: string;
  goals: string[];
  stages: ScenarioStage[];
  rubric: ScenarioRubric[];
  sessionId: string;
}

/** MessageRole 表示前端聊天消息角色。 */
export type MessageRole = "ai" | "user";

/** VoiceStatus 表示训练页语音控件状态。 */
export type VoiceStatus = "idle" | "recording" | "recognizing" | "thinking" | "speaking";

/** MessageMeta 描述消息附带的评分和纠错摘要。 */
export interface MessageMeta {
  asrConfidence?: number;
  wpm?: number;
  pauses?: number;
}

/** ChatMessage 描述聊天面板中的消息。 */
export interface ChatMessage {
  id: number;
  role: MessageRole;
  speaker: string;
  content: string;
  stage: string;
  createdAt: string;
  isTyping?: boolean;
  meta?: MessageMeta;
}

/** ConversationMessage 保留聊天消息的兼容类型别名。 */
export type ConversationMessage = ChatMessage;

/** TrainingTask 描述训练任务进度项。 */
export interface TrainingTask {
  label: string;
  status: "done" | "active" | "pending";
}

/** ScoreDimension 描述单个评分维度。 */
export interface ScoreDimension {
  key: "fluency" | "grammar" | "expression" | "vocabulary" | "completion";
  name: string;
  score: number;
  description: string;
}

/** Correction 描述前端展示的纠错卡片数据。 */
export interface Correction {
  title: string;
  category: "grammar" | "expression" | "vocabulary";
  original: string;
  suggestion: string;
  explanation: string;
  issues?: string[];
}

/** TrainingSession 描述训练页完整状态。 */
export interface TrainingSession {
  sessionId: string;
  sessionNo: string;
  scenario: Scenario;
  status: "running" | "finished";
  liveStatus: string;
  durationLabel: string;
  turnCount: number;
  progress: number;
  currentStage: string;
  voiceStatus: VoiceStatus;
  tasks: TrainingTask[];
  focusTags: string[];
  messages: ChatMessage[];
  currentScore: number;
  coachSummary: string;
  scores: ScoreDimension[];
  corrections: Correction[];
  naturalExpression: string;
}

/** BetterExpression 描述报告中的更自然表达建议。 */
export interface BetterExpression {
  before: string;
  after: string;
}

/** PracticePlanItem 描述报告中的后续练习项。 */
export interface PracticePlanItem {
  title: string;
  description: string;
}

/** TrainingReport 描述报告页展示模型。 */
export interface TrainingReport {
  sessionId: string;
  scenario: Scenario;
  durationLabel: string;
  turnCount: number;
  issueCount: number;
  completionRate: number;
  totalScore: number;
  grade: string;
  summary: string;
  scores: ScoreDimension[];
  majorProblems: string[];
  frequentErrors: Correction[];
  betterExpressions: BetterExpression[];
  nextPracticePlan: PracticePlanItem[];
  createdAt: string;
}

/** HistoryRecord 描述历史记录卡片数据。 */
export interface HistoryRecord {
  sessionId: string;
  sessionNo: string;
  scenario: Scenario;
  status: "running" | "finished";
  score: number;
  trainedAt: string;
  durationLabel: string;
  turnCount: number;
  majorProblem: string;
  reportStatus: "generated" | "not_generated";
}

export interface HistoryInsightSummary {
  days: number;
  totalSessions: number;
  finishedSessions: number;
  runningSessions: number;
  scoredSessions: number;
  generatedReports: number;
  averageScore: number | null;
  previousAverageScore: number | null;
  scoreDelta: number | null;
}

export interface HistoryScoreTrendPoint {
  date: string;
  averageScore: number;
  sessionCount: number;
}

export interface ScenarioTrend {
  scenario: Scenario;
  sessionCount: number;
  scoredSessions: number;
  averageScore: number | null;
  firstScore: number | null;
  latestScore: number | null;
  scoreDelta: number | null;
  lastTrainedAt: string;
}

export interface FrequentErrorInsight {
  key: string;
  title: string;
  category: "grammar" | "expression" | "vocabulary" | string;
  suggestion: string;
  count: number;
  latestEvidence: string;
  lastSeenAt: string;
  sourceSessionId: string;
}

export interface NextPracticeRecommendation {
  type: "scenario_repractice" | "continue_session" | string;
  reason: string;
  scenario: Scenario | null;
  sessionId: string;
  focus: string;
}

export interface HistoryInsights {
  summary: HistoryInsightSummary;
  scoreTrend: HistoryScoreTrendPoint[];
  scenarioTrends: ScenarioTrend[];
  frequentErrors: FrequentErrorInsight[];
  nextRecommendation: NextPracticeRecommendation | null;
}
