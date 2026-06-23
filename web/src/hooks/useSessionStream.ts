import { useEffect, useRef } from "react";
import { connectSessionStream, type SessionStreamEvent } from "../api/sessionStream";
import type { TrainingSession } from "../types";

interface UseSessionStreamOptions {
  sessionId: number | null;
  sessionStatus?: TrainingSession["status"];
  appendAIStreamDelta: (event: Extract<SessionStreamEvent, { type: "ai_message_delta" }>) => void;
  reload: (nextGoal?: string) => Promise<void>;
  speakAIReply: (messageId: number, content: string) => void;
  setStreamNotice: (message: string) => void;
  connect?: typeof connectSessionStream;
}

/** useSessionStream 连接训练 Session SSE 并把事件分发到页面状态。 */
export function useSessionStream({
  sessionId,
  sessionStatus,
  appendAIStreamDelta,
  reload,
  speakAIReply,
  setStreamNotice,
  connect = connectSessionStream,
}: UseSessionStreamOptions) {
  const streamErrorShown = useRef(false);

  useEffect(() => {
    streamErrorShown.current = false;
  }, [sessionId]);

  useEffect(() => {
    if (!sessionId || sessionStatus !== "running") {
      return undefined;
    }

    function handleStreamEvent(event: SessionStreamEvent) {
      if (event.type === "ai_message_delta") {
        setStreamNotice("正在接收 AI 回复片段");
        appendAIStreamDelta(event);
        return;
      }
      if (event.type === "ai_message_done") {
        speakAIReply(event.message_id ?? 0, event.content ?? "");
        void reload();
        return;
      }
      if (event.type === "correction_done" || event.type === "score_updated") {
        void reload();
        return;
      }
      if (event.type === "report_done") {
        setStreamNotice("课后报告已生成，可以前往报告页查看。");
        return;
      }
      if (event.type === "error") {
        setStreamNotice(event.message);
      }
    }

    return connect(sessionId, {
      onEvent: handleStreamEvent,
      onError: (message) => {
        if (!streamErrorShown.current) {
          streamErrorShown.current = true;
          setStreamNotice(message);
        }
      },
    });
  }, [appendAIStreamDelta, connect, reload, sessionId, sessionStatus, setStreamNotice, speakAIReply]);
}
