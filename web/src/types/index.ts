export type ScenarioCode = string;

export type Difficulty = string;

export interface ScenarioStage {
  name: string;
  description: string;
}

export interface ScenarioRubric {
  name: string;
  description: string;
}

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

export type MessageRole = "ai" | "user";

export type VoiceStatus = "idle" | "recording" | "recognizing" | "thinking";

export interface MessageMeta {
  asrConfidence?: number;
  wpm?: number;
  pauses?: number;
}

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

export type ConversationMessage = ChatMessage;

export interface TrainingTask {
  label: string;
  status: "done" | "active" | "pending";
}

export interface ScoreDimension {
  key: "fluency" | "grammar" | "expression" | "vocabulary" | "completion";
  name: string;
  score: number;
  description: string;
}

export interface Correction {
  title: string;
  category: "grammar" | "expression" | "vocabulary";
  original: string;
  suggestion: string;
  explanation: string;
  issues?: string[];
}

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

export interface BetterExpression {
  before: string;
  after: string;
}

export interface PracticePlanItem {
  title: string;
  description: string;
}

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
