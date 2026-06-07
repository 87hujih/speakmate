import { API_BASE_URL } from "./client";

export type SessionStreamEvent =
  | { type: "ai_message_delta"; message_id?: number; content: string }
  | { type: "ai_message_done"; message_id?: number; content?: string; stage?: string }
  | { type: "correction_done" }
  | { type: "score_updated" }
  | { type: "report_done" }
  | { type: "error"; code?: string; message: string };

interface SessionStreamHandlers {
  onEvent: (event: SessionStreamEvent) => void;
  onError?: (message: string) => void;
}

export function createSessionStreamUrl(sessionId: number | string, apiBaseUrl = API_BASE_URL) {
  return `${apiBaseUrl.replace(/\/$/, "")}/sessions/${sessionId}/stream`;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}

function stringValue(value: unknown) {
  return typeof value === "string" ? value : undefined;
}

function numberValue(value: unknown) {
  return typeof value === "number" && Number.isFinite(value) ? value : undefined;
}

export function parseSessionStreamEvent(type: SessionStreamEvent["type"], rawData: string): SessionStreamEvent {
  if (!rawData) {
    if (type === "ai_message_delta") {
      return { type, content: "" };
    }
    if (type === "error") {
      return { type, message: "stream error" };
    }

    return { type } as SessionStreamEvent;
  }

  try {
    const parsed: unknown = JSON.parse(rawData);
    const payload = isRecord(parsed) && isRecord(parsed.payload) ? parsed.payload : isRecord(parsed) ? parsed : {};

    if (type === "ai_message_delta") {
      return {
        type,
        message_id: numberValue(payload.message_id),
        content: stringValue(payload.delta) ?? stringValue(payload.content) ?? "",
      };
    }
    if (type === "ai_message_done") {
      return {
        type,
        message_id: numberValue(payload.message_id),
        content: stringValue(payload.content),
        stage: stringValue(payload.stage),
      };
    }
    if (type === "error") {
      return {
        type,
        code: stringValue(payload.code),
        message: stringValue(payload.message) ?? "stream error",
      };
    }

    return { type } as SessionStreamEvent;
  } catch {
    if (type === "ai_message_delta") {
      return { type, content: rawData };
    }

    if (type === "error") {
      return { type, message: rawData };
    }

    return { type } as SessionStreamEvent;
  }
}

export function connectSessionStream(sessionId: number | string, handlers: SessionStreamHandlers) {
  if (!("EventSource" in window)) {
    handlers.onError?.("当前浏览器不支持实时流，已使用普通消息接口。");
    return () => undefined;
  }

  const eventSource = new EventSource(createSessionStreamUrl(sessionId));
  const eventTypes: SessionStreamEvent["type"][] = [
    "ai_message_delta",
    "ai_message_done",
    "correction_done",
    "score_updated",
    "report_done",
    "error",
  ];

  eventTypes.forEach((type) => {
    eventSource.addEventListener(type, (event) => {
      handlers.onEvent(parseSessionStreamEvent(type, event.data));
    });
  });
  eventSource.onerror = () => {
    handlers.onError?.("实时流暂不可用，当前训练继续使用普通消息接口。");
  };

  return () => eventSource.close();
}
