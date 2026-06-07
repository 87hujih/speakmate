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
  mapHistoryRecord,
  mapReport,
  mapScenarioDetail,
  mapScenarioSummary,
  mapSessionDetailToTrainingSession,
} from "./adapters";

type ScenarioClient = Pick<typeof apiClient, "listScenarios" | "getScenario" | "createSession">;
type TrainingClient = Pick<typeof apiClient, "getSession" | "getScenario" | "listSessionCorrections" | "getSessionScore">;
type MessageClient = TrainingClient & Pick<typeof apiClient, "sendTextMessage">;
type AudioClient = TrainingClient & Pick<typeof apiClient, "uploadAudioMessage">;
type FinishClient = Pick<typeof apiClient, "finishSession">;
type ReportClient = Pick<typeof apiClient, "getReport">;
type GenerateReportClient = Pick<typeof apiClient, "generateReport">;
type HistoryClient = Pick<typeof apiClient, "listHistory">;

function hasApiCode(error: unknown, code: number) {
  return error instanceof ApiError && error.code === code;
}

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

export async function loadScenarioChoices(client: ScenarioClient = apiClient) {
  const summaries = await client.listScenarios();
  const details = await scenarioDetails(summaries, client);

  return details.map(mapScenarioDetail);
}

export async function createTrainingSession(scenarioId: number, client: Pick<typeof apiClient, "createSession"> = apiClient) {
  return client.createSession(scenarioId);
}

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

export async function sendTrainingText(sessionId: number, content: string, client: MessageClient = apiClient) {
  const result = await client.sendTextMessage(sessionId, content);
  const state = await loadTrainingSessionState(sessionId, client, result.next_goal);

  return {
    result,
    ...state,
  };
}

export async function sendTrainingAudio(sessionId: number, file: File, client: AudioClient = apiClient) {
  const result = await client.uploadAudioMessage(sessionId, file);
  const state = await loadTrainingSessionState(sessionId, client, result.next_goal);

  return {
    result,
    ...state,
  };
}

export async function finishTrainingSession(sessionId: number, client: FinishClient = apiClient) {
  return client.finishSession(sessionId);
}

export type ReportState = { status: "ready"; report: ReturnType<typeof mapReport> } | { status: "missing" };

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

export async function generateReportState(sessionId: number, client: GenerateReportClient = apiClient) {
  const report: BackendReport = await client.generateReport(sessionId);

  return mapReport(report);
}

export async function loadHistoryState(page: number, pageSize: number, client: HistoryClient = apiClient) {
  const result: BackendHistoryListResult = await client.listHistory(page, pageSize);

  return {
    records: result.items.map(mapHistoryRecord),
    page: result.page || page,
    pageSize: result.page_size || pageSize,
    total: result.total,
  };
}

export function mapScenarioFallback(scenario: BackendScenarioSummary) {
  return mapScenarioSummary(scenario);
}
