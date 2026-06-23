import { useCallback, useEffect, useState } from "react";
import type { NavigateFunction } from "react-router-dom";
import { ApiError } from "../api/client";
import { demoErrorMessage } from "../api/errors";
import { finishTrainingSession, loadTrainingSessionState, sendTrainingAudio, sendTrainingText } from "../api/loaders";
import type { SessionStreamEvent } from "../api/sessionStream";
import type { ChatMessage, TrainingSession } from "../types";

/** parseRouteSessionId 将路由参数解析为合法 Session ID。 */
export function parseRouteSessionId(value: string | undefined) {
  const numeric = Number(value);

  return Number.isInteger(numeric) && numeric > 0 ? numeric : null;
}

const streamingAIMessageId = -1;

/** appendAIStreamDelta 将 AI 流式片段合并到训练页临时消息。 */
export function appendAIStreamDelta(
  session: TrainingSession,
  event: Extract<SessionStreamEvent, { type: "ai_message_delta" }>,
): TrainingSession {
  if (!event.content) {
    return session;
  }

  const targetId = event.message_id && event.message_id > 0 ? event.message_id : streamingAIMessageId;
  const existingIndex = session.messages.findIndex((message) => message.id === targetId || message.id === streamingAIMessageId);
  const messages = [...session.messages];
  if (existingIndex >= 0) {
    const existing = messages[existingIndex];
    messages[existingIndex] = {
      ...existing,
      id: targetId,
      content: `${existing.content}${event.content}`,
      isTyping: true,
    };

    return { ...session, messages };
  }

  const streamingMessage: ChatMessage = {
    id: targetId,
    role: "ai",
    speaker: session.scenario.aiRole || "AI 教练",
    content: event.content,
    stage: session.currentStage,
    createdAt: new Date().toISOString(),
    isTyping: true,
  };

  return {
    ...session,
    messages: [...messages, streamingMessage],
  };
}

/** voiceUploadErrorMessage 将语音上传错误转换为中文提示。 */
function voiceUploadErrorMessage(error: unknown) {
  if (error instanceof ApiError) {
    if (error.code === 7004) {
      return "当前音频格式不支持，请换用支持 ogg、mp4 或 wav 录音的浏览器。";
    }
    if (error.code === 7005) {
      return "语音识别失败，请检查浏览器录音格式或后端 ASR 配置后重试。";
    }
    if (error.code === 7006) {
      return "语音识别没有返回有效文本，请重新录制。";
    }
  }

  return error instanceof Error ? error.message : "音频上传失败，请稍后重试。";
}

interface VoiceReply {
  messageId: number;
  content: string;
}

export type VoiceTextSendResult =
  | { type: "sent"; reply: VoiceReply }
  | { type: "invalid"; message: string }
  | { type: "ended"; message: string }
  | { type: "failed"; message: string }
  | { type: "ignored" };

export type VoiceAudioUploadResult =
  | { type: "sent"; transcript: string }
  | { type: "ended"; message: string }
  | { type: "failed"; message: string }
  | { type: "ignored" };

/** useTrainingSession 维护训练 Session 的加载、文本发送、语音提交和结束流程。 */
export function useTrainingSession(sessionId: string | undefined, navigate: NavigateFunction) {
  const numericSessionId = parseRouteSessionId(sessionId);
  const [session, setSession] = useState<TrainingSession | null>(null);
  const [draft, setDraft] = useState("");
  const [isLoading, setIsLoading] = useState(true);
  const [isSending, setIsSending] = useState(false);
  const [isFinishing, setIsFinishing] = useState(false);
  const [loadError, setLoadError] = useState("");
  const [sendError, setSendError] = useState("");

  const reload = useCallback(
    async (nextGoal?: string) => {
      if (!numericSessionId) {
        setLoadError("训练 ID 不合法");
        setIsLoading(false);
        return;
      }

      setIsLoading(true);
      setLoadError("");
      try {
        const result = await loadTrainingSessionState(numericSessionId, undefined, nextGoal);
        setSession(result.session);
      } catch (error) {
        setLoadError(error instanceof Error ? error.message : "训练加载失败");
      } finally {
        setIsLoading(false);
      }
    },
    [numericSessionId],
  );

  useEffect(() => {
    void reload();
  }, [reload]);

  const appendStreamDelta = useCallback((event: Extract<SessionStreamEvent, { type: "ai_message_delta" }>) => {
    setSession((current) => (current ? appendAIStreamDelta(current, event) : current));
  }, []);

  const clearSendError = useCallback(() => {
    setSendError("");
  }, []);

  const beginVoiceSocketSend = useCallback(() => {
    setIsSending(true);
    setSendError("");
  }, []);

  const endVoiceSocketSend = useCallback(() => {
    setIsSending(false);
  }, []);

  const sendDraft = useCallback(async () => {
    if (!numericSessionId || isSending) {
      return;
    }

    const content = draft.trim();
    if (!content) {
      setSendError("请输入有效文本后再发送。");
      return;
    }

    setIsSending(true);
    setSendError("");
    try {
      const result = await sendTrainingText(numericSessionId, content);
      setSession(result.session);
      setDraft("");
    } catch (error) {
      if (error instanceof ApiError && error.code === 2004) {
        setSendError("本次训练已结束，不能继续发送消息。");
        void reload();
      } else {
        setSendError(demoErrorMessage(error, "消息发送失败，请稍后重试。"));
      }
    } finally {
      setIsSending(false);
    }
  }, [draft, isSending, numericSessionId, reload]);

  const sendVoiceTranscript = useCallback(
    async (transcript: string): Promise<VoiceTextSendResult> => {
      if (!numericSessionId || isSending) {
        return { type: "ignored" };
      }

      const content = transcript.trim();
      if (!content) {
        return { type: "invalid", message: "实时听写没有生成有效文本，请重新尝试。" };
      }

      setIsSending(true);
      setSendError("");
      try {
        const result = await sendTrainingText(numericSessionId, content);
        setSession(result.session);

        return {
          type: "sent",
          reply: {
            messageId: result.result.ai_message.id,
            content: result.result.ai_message.content,
          },
        };
      } catch (error) {
        if (error instanceof ApiError && error.code === 2004) {
          void reload();
          return { type: "ended", message: "本次训练已结束，不能继续发送语音。" };
        }

        return { type: "failed", message: demoErrorMessage(error, "实时语音发送失败，请稍后重试。") };
      } finally {
        setIsSending(false);
      }
    },
    [isSending, numericSessionId, reload],
  );

  const uploadVoiceAudio = useCallback(
    async (file: File): Promise<VoiceAudioUploadResult> => {
      if (!numericSessionId) {
        return { type: "ignored" };
      }

      setIsSending(true);
      setSendError("");
      try {
        const result = await sendTrainingAudio(numericSessionId, file);
        setSession(result.session);

        return { type: "sent", transcript: result.result.transcript };
      } catch (error) {
        if (error instanceof ApiError && error.code === 2004) {
          void reload();
          return { type: "ended", message: "本次训练已结束，不能继续发送语音。" };
        }

        return { type: "failed", message: voiceUploadErrorMessage(error) };
      } finally {
        setIsSending(false);
      }
    },
    [numericSessionId, reload],
  );

  const finish = useCallback(async () => {
    if (!numericSessionId) {
      return;
    }
    if (session?.status === "finished") {
      navigate(`/report/${numericSessionId}`);
      return;
    }

    setIsFinishing(true);
    setSendError("");
    try {
      await finishTrainingSession(numericSessionId);
      navigate(`/report/${numericSessionId}`);
    } catch (error) {
      if (error instanceof ApiError && error.code === 2004) {
        navigate(`/report/${numericSessionId}`);
      } else {
        setSendError(demoErrorMessage(error, "结束训练失败"));
      }
    } finally {
      setIsFinishing(false);
    }
  }, [navigate, numericSessionId, session?.status]);

  return {
    numericSessionId,
    session,
    draft,
    setDraft,
    isLoading,
    isSending,
    isFinishing,
    loadError,
    sendError,
    reload,
    appendStreamDelta,
    clearSendError,
    beginVoiceSocketSend,
    endVoiceSocketSend,
    sendDraft,
    sendVoiceTranscript,
    uploadVoiceAudio,
    finish,
  };
}
