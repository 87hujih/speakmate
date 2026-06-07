import {
  API_BASE_URL,
  type BackendCorrectionSummary,
  type BackendMessage,
  type BackendScoreSummary,
} from "./client";

export type AudioWebSocketEvent =
  | { type: "start"; content_type?: string }
  | { type: "partial_transcript"; transcript: string; sequence: number }
  | {
      type: "final_transcript";
      transcript: string;
      user_message: BackendMessage;
      ai_message: BackendMessage;
      stage: string;
      next_goal: string;
      turn_count: number;
    }
  | ({ type: "correction" } & BackendCorrectionSummary)
  | ({ type: "score_updated" } & BackendScoreSummary)
  | { type: "end"; reason?: string }
  | { type: "error"; code?: string; message: string };

interface RawAudioWebSocketEvent {
  type?: string;
  payload?: Record<string, unknown>;
}

export function createAudioWebSocketUrl(sessionId: number | string, apiBaseUrl = API_BASE_URL, origin?: string) {
  const base = apiBaseUrl.replace(/\/$/, "");
  const path = `${base}/sessions/${sessionId}/audio/ws`;

  if (/^wss?:\/\//i.test(path)) {
    return path;
  }
  if (/^https?:\/\//i.test(path)) {
    return path.replace(/^http/i, "ws");
  }

  const browserOrigin = origin ?? (typeof window === "undefined" ? "" : window.location.origin);
  if (!browserOrigin) {
    return path;
  }

  const protocol = browserOrigin.startsWith("https://") ? "wss" : "ws";
  const host = browserOrigin.replace(/^https?:\/\//i, "");
  const normalizedPath = path.startsWith("/") ? path : `/${path}`;

  return `${protocol}://${host}${normalizedPath}`;
}

export function parseAudioWebSocketEvent(rawData: string): AudioWebSocketEvent {
  let parsed: RawAudioWebSocketEvent;
  try {
    parsed = JSON.parse(rawData) as RawAudioWebSocketEvent;
  } catch {
    return { type: "error", message: rawData || "audio websocket event invalid" };
  }

  const payload = parsed.payload ?? {};
  switch (parsed.type) {
    case "start":
      return { type: "start", content_type: stringValue(payload.content_type) };
    case "partial_transcript":
      return {
        type: "partial_transcript",
        transcript: stringValue(payload.transcript) ?? "",
        sequence: numberValue(payload.sequence) ?? 0,
      };
    case "final_transcript":
      return {
        type: "final_transcript",
        transcript: stringValue(payload.transcript) ?? "",
        user_message: payload.user_message as BackendMessage,
        ai_message: payload.ai_message as BackendMessage,
        stage: stringValue(payload.stage) ?? "",
        next_goal: stringValue(payload.next_goal) ?? "",
        turn_count: numberValue(payload.turn_count) ?? 0,
      };
    case "correction":
      return {
        type: "correction",
        has_errors: Boolean(payload.has_errors),
        error_count: numberValue(payload.error_count) ?? 0,
      };
    case "score_updated":
      return {
        type: "score_updated",
        total_score: numberValue(payload.total_score) ?? 0,
        grammar: numberValue(payload.grammar) ?? 0,
        expression: numberValue(payload.expression) ?? 0,
      };
    case "end":
      return { type: "end", reason: stringValue(payload.reason) };
    case "error":
      return {
        type: "error",
        code: stringValue(payload.code),
        message: stringValue(payload.message) ?? "audio websocket error",
      };
    default:
      return { type: "error", message: "audio websocket event unsupported" };
  }
}

function stringValue(value: unknown) {
  return typeof value === "string" ? value : undefined;
}

function numberValue(value: unknown) {
  return typeof value === "number" && Number.isFinite(value) ? value : undefined;
}
