export interface ApiResponse<T> {
  code: number;
  message: string;
  data?: T;
}

export interface BackendScenarioSummary {
  id: number;
  code: string;
  name: string;
  description: string;
  difficulty: string;
}

export interface BackendScenario extends BackendScenarioSummary {
  ai_role: string;
  user_goal: string;
  opening_message: string;
  stages: Array<{
    name: string;
    description: string;
  }>;
  rubric: Array<{
    name: string;
    description: string;
  }>;
}

export type BackendSessionStatus = "running" | "finished";
export type BackendMessageRole = "user" | "ai";

export interface BackendMessage {
  id: number;
  session_id: number;
  role: BackendMessageRole;
  content: string;
  stage: string;
  created_at: string;
}

export interface BackendSessionCreateResult {
  session_id: number;
  session_no: string;
  scenario_id: number;
  status: BackendSessionStatus;
  opening_message: string;
}

export interface BackendSessionDetail {
  session_id: number;
  session_no: string;
  scenario: BackendScenarioSummary;
  status: BackendSessionStatus;
  turn_count: number;
  messages: BackendMessage[];
  created_at: string;
  ended_at: string | null;
}

export interface BackendSessionFinishResult {
  session_id: number;
  status: BackendSessionStatus;
  turn_count: number;
  ended_at: string;
}

export interface BackendCorrectionSummary {
  has_errors: boolean;
  error_count: number;
}

export interface BackendScoreSummary {
  total_score: number;
  grammar: number;
  expression: number;
}

export interface BackendSendMessageResult {
  user_message: BackendMessage;
  ai_message: BackendMessage;
  stage: string;
  next_goal: string;
  turn_count: number;
  correction_summary: BackendCorrectionSummary;
  score_summary: BackendScoreSummary;
}

export interface BackendUploadAudioResult extends BackendSendMessageResult {
  transcript: string;
}

export interface BackendCorrectionError {
  type: "grammar" | "vocabulary" | "expression" | "structure" | "scenario";
  span: string;
  suggestion: string;
  explanation: string;
}

export interface BackendCorrectionResult {
  message_id: number;
  session_id: number;
  original_text: string;
  corrected_text: string;
  errors: BackendCorrectionError[] | null;
  better_expressions: string[] | null;
}

export interface BackendScoreResult {
  message_id: number;
  session_id: number;
  fluency: number;
  grammar: number;
  expression: number;
  vocabulary: number;
  completion: number;
  total_score: number;
  comment: string;
}

export interface BackendReport {
  session_id: number;
  scenario: {
    id: number;
    code: string;
    name: string;
    difficulty: string;
  };
  duration_seconds: number;
  turn_count: number;
  total_score: number;
  scores: BackendScoreResult;
  summary: string;
  major_problems: string[] | null;
  frequent_errors: string[] | null;
  better_expressions: string[] | null;
  next_practice_plan: string[] | null;
  created_at: string;
}

export interface BackendHistoryItem {
  session_id: number;
  session_no: string;
  user_id: number;
  scenario: BackendScenarioSummary;
  status: BackendSessionStatus;
  turn_count: number;
  total_score: number | null;
  report_status: "generated" | "not_generated";
  created_at: string;
  ended_at: string | null;
}

export interface BackendHistoryListResult {
  items: BackendHistoryItem[];
  page: number;
  page_size: number;
  total: number;
}

export interface BackendHistoryInsightSummary {
  days: number;
  total_sessions: number;
  finished_sessions: number;
  running_sessions: number;
  scored_sessions: number;
  generated_reports: number;
  average_score: number | null;
  previous_average_score: number | null;
  score_delta: number | null;
}

export interface BackendHistoryScoreTrendPoint {
  date: string;
  average_score: number;
  session_count: number;
}

export interface BackendScenarioTrend {
  scenario: BackendScenarioSummary;
  session_count: number;
  scored_sessions: number;
  average_score: number | null;
  first_score: number | null;
  latest_score: number | null;
  score_delta: number | null;
  last_trained_at: string;
}

export interface BackendFrequentErrorInsight {
  key: string;
  title: string;
  category: "grammar" | "expression" | "vocabulary" | string;
  suggestion: string;
  count: number;
  latest_evidence: string;
  last_seen_at: string;
  source_session_id: number;
}

export interface BackendNextPracticeRecommendation {
  type: "scenario_repractice" | "continue_session" | string;
  reason: string;
  scenario: BackendScenarioSummary | null;
  session_id: number;
  focus: string;
}

export interface BackendHistoryInsights {
  summary: BackendHistoryInsightSummary;
  score_trend: BackendHistoryScoreTrendPoint[];
  scenario_trends: BackendScenarioTrend[];
  frequent_errors: BackendFrequentErrorInsight[];
  next_recommendation: BackendNextPracticeRecommendation | null;
}

export class ApiError extends Error {
  code: number;
  status: number;

  constructor(message: string, code: number, status: number) {
    super(message);
    this.name = "ApiError";
    this.code = code;
    this.status = status;
  }
}

export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? "/api/v1";

function buildApiUrl(path: string) {
  const base = API_BASE_URL.replace(/\/$/, "");
  const normalizedPath = path.startsWith("/") ? path : `/${path}`;

  return `${base}${normalizedPath}`;
}

async function readPayload<T>(response: Response): Promise<ApiResponse<T>> {
  try {
    return (await response.json()) as ApiResponse<T>;
  } catch {
    return {
      code: response.ok ? 0 : response.status,
      message: response.statusText || "request failed",
    };
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(buildApiUrl(path), {
    headers: {
      "Content-Type": "application/json",
      ...init?.headers,
    },
    ...init,
  });
  const payload = await readPayload<T>(response);

  if (!response.ok || payload.code !== 0) {
    throw new ApiError(payload.message || "request failed", payload.code, response.status);
  }
  if (payload.data === undefined) {
    throw new ApiError("response data missing", payload.code, response.status);
  }

  return payload.data;
}

async function requestForm<T>(path: string, body: FormData): Promise<T> {
  const response = await fetch(buildApiUrl(path), {
    method: "POST",
    body,
  });
  const payload = await readPayload<T>(response);

  if (!response.ok || payload.code !== 0) {
    throw new ApiError(payload.message || "request failed", payload.code, response.status);
  }
  if (payload.data === undefined) {
    throw new ApiError("response data missing", payload.code, response.status);
  }

  return payload.data;
}

function withPagination(path: string, page: number, pageSize: number) {
  const params = new URLSearchParams();
  params.set("page", String(page));
  params.set("page_size", String(pageSize));

  return `${path}?${params.toString()}`;
}

export const apiClient = {
  get: <T>(path: string) => request<T>(path),
  post: <T>(path: string, body?: unknown) =>
    request<T>(path, {
      method: "POST",
      body: body === undefined ? undefined : JSON.stringify(body),
    }),
  listScenarios: () => request<BackendScenarioSummary[]>("/scenarios"),
  getScenario: (scenarioId: number) => request<BackendScenario>(`/scenarios/${scenarioId}`),
  createSession: (scenarioId: number, userId?: number) =>
    request<BackendSessionCreateResult>("/sessions", {
      method: "POST",
      body: JSON.stringify(userId === undefined ? { scenario_id: scenarioId } : { scenario_id: scenarioId, user_id: userId }),
    }),
  getSession: (sessionId: number) => request<BackendSessionDetail>(`/sessions/${sessionId}`),
  finishSession: (sessionId: number) =>
    request<BackendSessionFinishResult>(`/sessions/${sessionId}/finish`, {
      method: "POST",
    }),
  sendTextMessage: (sessionId: number, content: string) =>
    request<BackendSendMessageResult>(`/sessions/${sessionId}/messages`, {
      method: "POST",
      body: JSON.stringify({ content }),
    }),
  uploadAudioMessage: (sessionId: number, file: File) => {
    const body = new FormData();
    body.append("audio", file);

    return requestForm<BackendUploadAudioResult>(`/sessions/${sessionId}/audio`, body);
  },
  listSessionCorrections: (sessionId: number) => request<BackendCorrectionResult[]>(`/sessions/${sessionId}/corrections`),
  getMessageCorrections: (messageId: number) => request<BackendCorrectionResult>(`/messages/${messageId}/corrections`),
  getSessionScore: (sessionId: number) => request<BackendScoreResult>(`/sessions/${sessionId}/scores`),
  generateReport: (sessionId: number) =>
    request<BackendReport>(`/sessions/${sessionId}/report`, {
      method: "POST",
    }),
  getReport: (sessionId: number) => request<BackendReport>(`/sessions/${sessionId}/report`),
  listHistory: (page = 1, pageSize = 20) => request<BackendHistoryListResult>(withPagination("/sessions", page, pageSize)),
  listUserHistory: (userId: number, page = 1, pageSize = 20) =>
    request<BackendHistoryListResult>(withPagination(`/users/${userId}/sessions`, page, pageSize)),
  getHistoryInsights: (days = 30, userId?: number) =>
    request<BackendHistoryInsights>(
      userId === undefined ? `/history/insights?days=${days}` : `/history/insights?days=${days}&user_id=${userId}`,
    ),
};
