/** ApiResponse 描述后端统一响应体。 */
export interface ApiResponse<T> {
  code: number;
  message: string;
  data?: T;
}

/** BackendScenarioSummary 描述后端场景摘要结构。 */
export interface BackendScenarioSummary {
  id: number;
  code: string;
  name: string;
  description: string;
  difficulty: string;
}

/** BackendScenario 描述后端场景详情结构。 */
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

/** BackendSessionStatus 表示后端 Session 生命周期状态。 */
export type BackendSessionStatus = "running" | "finished";
/** BackendMessageRole 表示后端消息发送方角色。 */
export type BackendMessageRole = "user" | "ai";

/** BackendMessage 描述后端消息结构。 */
export interface BackendMessage {
  id: number;
  session_id: number;
  role: BackendMessageRole;
  content: string;
  stage: string;
  created_at: string;
}

/** BackendSessionCreateResult 描述创建 Session 的后端返回值。 */
export interface BackendSessionCreateResult {
  session_id: number;
  session_no: string;
  scenario_id: number;
  status: BackendSessionStatus;
  opening_message: string;
}

/** BackendSessionDetail 描述训练 Session 详情返回值。 */
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

/** BackendSessionFinishResult 描述结束 Session 的后端返回值。 */
export interface BackendSessionFinishResult {
  session_id: number;
  status: BackendSessionStatus;
  turn_count: number;
  ended_at: string;
}

/** BackendCorrectionSummary 描述消息响应中的纠错摘要。 */
export interface BackendCorrectionSummary {
  has_errors: boolean;
  error_count: number;
}

/** BackendScoreSummary 描述消息响应中的评分摘要。 */
export interface BackendScoreSummary {
  total_score: number;
  grammar: number;
  expression: number;
}

/** BackendSendMessageResult 描述文本消息发送结果。 */
export interface BackendSendMessageResult {
  user_message: BackendMessage;
  ai_message: BackendMessage;
  stage: string;
  next_goal: string;
  turn_count: number;
  correction_summary: BackendCorrectionSummary;
  score_summary: BackendScoreSummary;
}

/** BackendUploadAudioResult 描述音频消息上传结果。 */
export interface BackendUploadAudioResult extends BackendSendMessageResult {
  transcript: string;
}

/** BackendCorrectionError 描述后端单条纠错问题。 */
export interface BackendCorrectionError {
  type: "grammar" | "vocabulary" | "expression" | "structure" | "scenario";
  span: string;
  suggestion: string;
  explanation: string;
}

/** BackendCorrectionResult 描述后端纠错结果。 */
export interface BackendCorrectionResult {
  message_id: number;
  session_id: number;
  original_text: string;
  corrected_text: string;
  errors: BackendCorrectionError[] | null;
  better_expressions: string[] | null;
}

/** BackendScoreResult 描述后端评分结果。 */
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

/** BackendReport 描述后端课后报告结构。 */
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

/** BackendHistoryItem 描述后端历史列表条目。 */
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

/** BackendHistoryListResult 描述后端历史分页返回值。 */
export interface BackendHistoryListResult {
  items: BackendHistoryItem[];
  page: number;
  page_size: number;
  total: number;
}

/** BackendHistoryInsightSummary 描述后端历史洞察汇总。 */
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

/** BackendHistoryScoreTrendPoint 描述历史评分趋势点。 */
export interface BackendHistoryScoreTrendPoint {
  date: string;
  average_score: number;
  session_count: number;
}

/** BackendScenarioTrend 描述场景维度训练趋势。 */
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

/** BackendFrequentErrorInsight 描述高频错误洞察。 */
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

/** BackendNextPracticeRecommendation 描述下一次训练推荐。 */
export interface BackendNextPracticeRecommendation {
  type: "scenario_repractice" | "continue_session" | string;
  reason: string;
  scenario: BackendScenarioSummary | null;
  session_id: number;
  focus: string;
}

/** BackendHistoryInsights 描述历史洞察接口返回值。 */
export interface BackendHistoryInsights {
  summary: BackendHistoryInsightSummary;
  score_trend: BackendHistoryScoreTrendPoint[] | null;
  scenario_trends: BackendScenarioTrend[] | null;
  frequent_errors: BackendFrequentErrorInsight[] | null;
  next_recommendation: BackendNextPracticeRecommendation | null;
}

/** ApiError 保存统一 API 错误码、HTTP 状态和错误提示。 */
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

/** buildApiUrl 基于统一 API 前缀生成请求地址。 */
function buildApiUrl(path: string) {
  const base = API_BASE_URL.replace(/\/$/, "");
  const normalizedPath = path.startsWith("/") ? path : `/${path}`;

  return `${base}${normalizedPath}`;
}

/** readPayload 读取并解析统一 API 响应体。 */
async function readPayload<T>(response: Response): Promise<ApiResponse<T>> {
  try {
    return (await response.json()) as ApiResponse<T>;
  } catch {
    return {
      code: response.ok ? 0 : response.status,
      message: response.statusText || "请求失败",
    };
  }
}

/** request 发送 JSON API 请求并校验统一响应结构。 */
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
    throw new ApiError(payload.message || "请求失败", payload.code, response.status);
  }
  if (payload.data === undefined) {
    throw new ApiError("响应数据缺失", payload.code, response.status);
  }

  return payload.data;
}

/** requestForm 发送表单 API 请求并校验统一响应结构。 */
async function requestForm<T>(path: string, body: FormData): Promise<T> {
  const response = await fetch(buildApiUrl(path), {
    method: "POST",
    body,
  });
  const payload = await readPayload<T>(response);

  if (!response.ok || payload.code !== 0) {
    throw new ApiError(payload.message || "请求失败", payload.code, response.status);
  }
  if (payload.data === undefined) {
    throw new ApiError("响应数据缺失", payload.code, response.status);
  }

  return payload.data;
}

/** withPagination 为列表路径附加分页参数。 */
function withPagination(path: string, page: number, pageSize: number) {
  const params = new URLSearchParams();
  params.set("page", String(page));
  params.set("page_size", String(pageSize));

  return `${path}?${params.toString()}`;
}

function withHistoryInsightsParams(days: number, userId?: number) {
  const params = new URLSearchParams();
  params.set("days", String(days));
  if (userId !== undefined) {
    params.set("user_id", String(userId));
  }

  return `/history/insights?${params.toString()}`;
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
  getHistoryInsights: (days = 30, userId?: number) => request<BackendHistoryInsights>(withHistoryInsightsParams(days, userId)),
};
