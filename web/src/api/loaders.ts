import {
  ApiError,
  apiClient,
  type BackendCorrectionResult,
  type BackendHistoryListResult,
  type BackendReport,
  type BackendScenario,
  type BackendScenarioSummary,
  type BackendScoreResult,
  type BackendSessionDetail,
} from "./client";
import {
  mapHistoryInsights,
  mapHistoryRecord,
  mapReport,
  mapScenarioDetail,
  mapScenarioSummary,
  mapSessionDetailToTrainingSession,
} from "./adapters";

/** ScenarioClient 限定首页场景加载需要的 API 能力。 */
type ScenarioClient = Pick<typeof apiClient, "listScenarios" | "getScenario" | "createSession">;
/** TrainingClient 限定训练状态加载需要的 API 能力。 */
type TrainingClient = Pick<typeof apiClient, "getSession" | "getScenario" | "listSessionCorrections" | "getSessionScore">;
/** MessageClient 限定文本发送需要的 API 能力。 */
type MessageClient = TrainingClient & Pick<typeof apiClient, "sendTextMessage">;
/** AudioClient 限定音频发送需要的 API 能力。 */
type AudioClient = TrainingClient & Pick<typeof apiClient, "uploadAudioMessage">;
/** FinishClient 限定结束训练需要的 API 能力。 */
type FinishClient = Pick<typeof apiClient, "finishSession">;
/** ReportClient 限定报告查询需要的 API 能力。 */
type ReportClient = Pick<typeof apiClient, "getReport">;
/** GenerateReportClient 限定报告生成需要的 API 能力。 */
type GenerateReportClient = Pick<typeof apiClient, "generateReport">;
/** HistoryClient 限定历史列表需要的 API 能力。 */
type HistoryClient = Pick<typeof apiClient, "listHistory">;
/** HistoryInsightsClient 限定历史洞察需要的 API 能力。 */
type HistoryInsightsClient = Pick<typeof apiClient, "getHistoryInsights">;

/** hasApiCode 判断未知错误是否为指定 API 错误码。 */
function hasApiCode(error: unknown, code: number) {
  return error instanceof ApiError && error.code === code;
}

/** optionalCorrections 在纠错缺失时返回空列表。 */
async function optionalCorrections(client: TrainingClient, sessionId: number): Promise<BackendCorrectionResult[]> {
  try {
    return await client.listSessionCorrections(sessionId);
  } catch (error) {
    if (hasApiCode(error, 4002)) {
      return [];
    }

    throw error;
  }
}

/** optionalScore 在评分缺失时返回空状态。 */
async function optionalScore(client: TrainingClient, sessionId: number): Promise<BackendScoreResult | null> {
  try {
    return await client.getSessionScore(sessionId);
  } catch (error) {
    if (hasApiCode(error, 4003)) {
      return null;
    }

    throw error;
  }
}

/** scenarioDetails 并发加载场景详情，失败时使用摘要兜底。 */
async function scenarioDetails(
  summaries: BackendScenarioSummary[],
  client: Pick<typeof apiClient, "getScenario">,
): Promise<BackendScenario[]> {
  const results = await Promise.allSettled(summaries.map((scenario) => client.getScenario(scenario.id)));

  return results.map((result, index) => {
    if (result.status === "fulfilled") {
      return result.value;
    }

    return {
      ...summaries[index],
      ai_role: "AI 教练",
      user_goal: summaries[index].description,
      opening_message: "",
      stages: [],
      rubric: [],
    };
  });
}

/** loadScenarioChoices 加载首页可选训练场景。 */
export async function loadScenarioChoices(client: ScenarioClient = apiClient) {
  const summaries = await client.listScenarios();
  const details = await scenarioDetails(summaries, client);

  return details.map(mapScenarioDetail);
}

/** createTrainingSession 调用后端创建训练 Session。 */
export async function createTrainingSession(scenarioId: number, client: Pick<typeof apiClient, "createSession"> = apiClient) {
  return client.createSession(scenarioId);
}

/** loadTrainingSessionState 加载训练页所需的完整状态。 */
export async function loadTrainingSessionState(sessionId: number, client: TrainingClient = apiClient, nextGoal?: string) {
  const session = await client.getSession(sessionId);
  const [scenario, corrections, score] = await Promise.all([
    client.getScenario(session.scenario.id),
    optionalCorrections(client, sessionId),
    optionalScore(client, sessionId),
  ]);

  return {
    session: mapSessionDetailToTrainingSession({
      session,
      scenario,
      corrections,
      score,
      nextGoal,
    }),
    raw: {
      session,
      scenario,
      corrections,
      score,
    },
  };
}

/** sendTrainingText 发送文本消息并刷新训练状态。 */
export async function sendTrainingText(sessionId: number, content: string, client: MessageClient = apiClient) {
  const result = await client.sendTextMessage(sessionId, content);
  const state = await loadTrainingSessionState(sessionId, client, result.next_goal);

  return {
    result,
    ...state,
  };
}

/** sendTrainingAudio 上传音频消息并刷新训练状态。 */
export async function sendTrainingAudio(sessionId: number, file: File, client: AudioClient = apiClient) {
  const result = await client.uploadAudioMessage(sessionId, file);
  const state = await loadTrainingSessionState(sessionId, client, result.next_goal);

  return {
    result,
    ...state,
  };
}

/** finishTrainingSession 结束当前训练 Session。 */
export async function finishTrainingSession(sessionId: number, client: FinishClient = apiClient) {
  return client.finishSession(sessionId);
}

/** ReportState 表示报告页已生成或缺失两种状态。 */
export type ReportState = { status: "ready"; report: ReturnType<typeof mapReport> } | { status: "missing" };

/** loadReportState 加载报告状态并区分未生成报告。 */
export async function loadReportState(sessionId: number, client: ReportClient = apiClient): Promise<ReportState> {
  try {
    return {
      status: "ready",
      report: mapReport(await client.getReport(sessionId)),
    };
  } catch (error) {
    if (hasApiCode(error, 5003)) {
      return { status: "missing" };
    }

    throw error;
  }
}

/** generateReportState 触发报告生成并转换展示模型。 */
export async function generateReportState(sessionId: number, client: GenerateReportClient = apiClient) {
  const report: BackendReport = await client.generateReport(sessionId);

  return mapReport(report);
}

/** loadHistoryState 加载历史记录分页数据。 */
export async function loadHistoryState(page: number, pageSize: number, client: HistoryClient = apiClient) {
  const result: BackendHistoryListResult = await client.listHistory(page, pageSize);

  return {
    records: result.items.map(mapHistoryRecord),
    page: result.page || page,
    pageSize: result.page_size || pageSize,
    total: result.total,
  };
}

/** loadHistoryInsights 加载历史洞察数据。 */
export async function loadHistoryInsights(days = 30, client: HistoryInsightsClient = apiClient) {
  return mapHistoryInsights(await client.getHistoryInsights(days));
}

/** mapScenarioFallback 将缺失详情的场景摘要转换为可展示模型。 */
export function mapScenarioFallback(scenario: BackendScenarioSummary) {
  return mapScenarioSummary(scenario);
}
