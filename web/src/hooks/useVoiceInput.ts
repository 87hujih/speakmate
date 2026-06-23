import { useCallback, useEffect, useReducer, useRef } from "react";
import { createAudioWebSocketUrl, parseAudioWebSocketEvent, type AudioWebSocketEvent } from "../api/audioWebSocket";
import type { TrainingSession, VoiceStatus } from "../types";
import { extensionForAudioMimeType, selectSupportedAudioMimeType } from "../utils/audioMime";
import {
  createRealtimeSpeechSession,
  isRealtimeSpeechSupported,
  shouldFallbackToRecordedAudio,
  type RealtimeSpeechSession,
} from "../utils/realtimeSpeech";
import type { VoiceAudioUploadResult, VoiceTextSendResult } from "./useTrainingSession";

export interface VoiceInputState {
  status: VoiceStatus;
  transcript: string;
  error: string;
}

export type VoiceInputAction =
  | { type: "recording_started" }
  | { type: "partial_transcript"; transcript: string }
  | { type: "recognizing"; transcript?: string }
  | { type: "thinking"; transcript?: string }
  | { type: "idle" }
  | { type: "clear_error" }
  | { type: "clear_transcript" }
  | { type: "error"; message: string };

export const initialVoiceInputState: VoiceInputState = {
  status: "idle",
  transcript: "",
  error: "",
};

/** voiceInputStateReducer 描述语音输入控件的核心状态机。 */
export function voiceInputStateReducer(state: VoiceInputState, action: VoiceInputAction): VoiceInputState {
  if (action.type === "recording_started") {
    return { status: "recording", transcript: "", error: "" };
  }
  if (action.type === "partial_transcript") {
    return { ...state, transcript: action.transcript };
  }
  if (action.type === "recognizing") {
    return { ...state, status: "recognizing", transcript: action.transcript ?? state.transcript };
  }
  if (action.type === "thinking") {
    return { ...state, status: "thinking", transcript: action.transcript ?? state.transcript };
  }
  if (action.type === "idle") {
    return { ...state, status: "idle" };
  }
  if (action.type === "clear_error") {
    return { ...state, error: "" };
  }
  if (action.type === "clear_transcript") {
    return { ...state, transcript: "" };
  }
  if (action.type === "error") {
    return { ...state, status: "idle", error: action.message };
  }

  return state;
}

interface UseVoiceInputOptions {
  sessionId: number | null;
  sessionStatus?: TrainingSession["status"];
  isSending: boolean;
  isFinishing: boolean;
  isPlaybackSpeaking: boolean;
  setStreamNotice: (message: string) => void;
  reload: (nextGoal?: string) => Promise<void>;
  clearSendError: () => void;
  beginVoiceSocketSend: () => void;
  endVoiceSocketSend: () => void;
  sendVoiceTranscript: (transcript: string) => Promise<VoiceTextSendResult>;
  uploadVoiceAudio: (file: File) => Promise<VoiceAudioUploadResult>;
  requestSpeakNextAI: () => void;
  cancelPendingAIReply: () => void;
  speakAIReply: (messageId: number, content: string) => void;
}

/** realtimeSpeechErrorMessage 将浏览器实时听写错误转换为中文提示。 */
function realtimeSpeechErrorMessage(code: string) {
  if (code === "not-allowed" || code === "service-not-allowed") {
    return "浏览器实时听写没有麦克风权限，请授权后重试。";
  }
  if (code === "no-speech") {
    return "没有检测到有效英文语音，请重新尝试。";
  }
  if (code === "network") {
    return "浏览器实时听写网络不可用，请稍后重试或使用录音上传。";
  }
  if (code === "realtime_speech_unsupported") {
    return "当前浏览器不支持实时听写，已改用录音上传。";
  }

  return "实时听写失败，请重新尝试或换用录音上传。";
}

/** voiceSocketErrorMessage 将实时语音 WebSocket 错误转换为中文提示。 */
function voiceSocketErrorMessage(event: Extract<AudioWebSocketEvent, { type: "error" }>) {
  if (event.code === "audio_file_type_unsupported") {
    return "当前录音格式不支持，请换用支持 ogg、mp4 或 wav 录音的浏览器。";
  }
  if (event.code === "asr_client_failed") {
    return "语音识别失败，请检查浏览器录音格式或后端 ASR 配置后重试。";
  }
  if (event.code === "audio_transcript_required") {
    return "语音识别没有返回有效文本，请重新录制。";
  }

  return event.message;
}

/** useVoiceInput 管理实时听写、录音上传和音频 WebSocket 输入。 */
export function useVoiceInput({
  sessionId,
  sessionStatus,
  isSending,
  isFinishing,
  isPlaybackSpeaking,
  setStreamNotice,
  reload,
  clearSendError,
  beginVoiceSocketSend,
  endVoiceSocketSend,
  sendVoiceTranscript,
  uploadVoiceAudio,
  requestSpeakNextAI,
  cancelPendingAIReply,
  speakAIReply,
}: UseVoiceInputOptions) {
  const [state, dispatch] = useReducer(voiceInputStateReducer, initialVoiceInputState);
  const mediaRecorderRef = useRef<MediaRecorder | null>(null);
  const mediaStreamRef = useRef<MediaStream | null>(null);
  const voiceSocketRef = useRef<WebSocket | null>(null);
  const voiceSocketReadyRef = useRef(false);
  const voiceSocketChunkSentRef = useRef(false);
  const realtimeSpeechRef = useRef<RealtimeSpeechSession | null>(null);
  const realtimeFinalReceivedRef = useRef(false);
  const realtimeStopRequestedRef = useRef(false);
  const realtimeFallbackAttemptedRef = useRef(false);
  const realtimeFallbackOnEndRef = useRef(false);
  const audioChunksRef = useRef<Blob[]>([]);

  const stopMediaStream = useCallback(() => {
    mediaStreamRef.current?.getTracks().forEach((track) => track.stop());
    mediaStreamRef.current = null;
  }, []);

  const closeVoiceSocket = useCallback(() => {
    const socket = voiceSocketRef.current;
    voiceSocketRef.current = null;
    voiceSocketReadyRef.current = false;
    voiceSocketChunkSentRef.current = false;
    if (socket && socket.readyState === WebSocket.OPEN) {
      socket.close();
    }
  }, []);

  const closeRealtimeSpeech = useCallback(() => {
    realtimeSpeechRef.current?.abort();
    realtimeSpeechRef.current = null;
    realtimeFinalReceivedRef.current = false;
  }, []);

  useEffect(() => {
    return () => {
      const recorder = mediaRecorderRef.current;
      if (recorder && recorder.state !== "inactive") {
        recorder.onstop = null;
        recorder.stop();
      }
      closeRealtimeSpeech();
      closeVoiceSocket();
      stopMediaStream();
    };
  }, [closeRealtimeSpeech, closeVoiceSocket, stopMediaStream]);

  const supportedAudioMimeType = useCallback(() => {
    if (typeof MediaRecorder === "undefined" || typeof MediaRecorder.isTypeSupported !== "function") {
      return "";
    }

    return selectSupportedAudioMimeType((candidate) => MediaRecorder.isTypeSupported(candidate));
  }, []);

  const uploadRecordedAudio = useCallback(
    async (blob: Blob, mimeType: string) => {
      if (!sessionId) {
        return;
      }
      if (blob.size === 0) {
        dispatch({ type: "error", message: "录音为空，请重新录制。" });
        return;
      }

      const fileType = mimeType || blob.type || "audio/webm";
      const file = new File([blob], `answer-${Date.now()}.${extensionForAudioMimeType(fileType)}`, { type: fileType });

      clearSendError();
      dispatch({ type: "clear_error" });
      dispatch({ type: "clear_transcript" });
      const result = await uploadVoiceAudio(file);
      if (result.type === "sent") {
        dispatch({ type: "partial_transcript", transcript: result.transcript });
        dispatch({ type: "idle" });
        return;
      }
      if (result.type === "ended" || result.type === "failed") {
        dispatch({ type: "error", message: result.message });
        return;
      }

      dispatch({ type: "idle" });
    },
    [clearSendError, sessionId, uploadVoiceAudio],
  );

  const handleAudioWebSocketEvent = useCallback(
    (event: AudioWebSocketEvent) => {
      if (event.type === "start") {
        setStreamNotice("实时语音通道已连接。");
        return;
      }
      if (event.type === "partial_transcript") {
        dispatch({ type: "partial_transcript", transcript: event.transcript });
        return;
      }
      if (event.type === "final_transcript") {
        dispatch({ type: "thinking", transcript: event.transcript });
        void reload(event.next_goal);
        return;
      }
      if (event.type === "correction" || event.type === "score_updated") {
        void reload();
        return;
      }
      if (event.type === "end") {
        endVoiceSocketSend();
        dispatch({ type: "idle" });
        closeVoiceSocket();
        return;
      }
      if (event.type === "error") {
        endVoiceSocketSend();
        dispatch({ type: "error", message: voiceSocketErrorMessage(event) });
      }
    },
    [closeVoiceSocket, endVoiceSocketSend, reload, setStreamNotice],
  );

  const openVoiceSocket = useCallback(
    (nextSessionId: number, mimeType: string) => {
      if (typeof WebSocket === "undefined") {
        return null;
      }

      const socket = new WebSocket(createAudioWebSocketUrl(nextSessionId));
      voiceSocketRef.current = socket;
      voiceSocketReadyRef.current = false;
      voiceSocketChunkSentRef.current = false;
      socket.onopen = () => {
        voiceSocketReadyRef.current = true;
        socket.send(
          JSON.stringify({
            type: "start",
            payload: { content_type: mimeType || "audio/webm" },
          }),
        );
      };
      socket.onmessage = (event) => {
        handleAudioWebSocketEvent(parseAudioWebSocketEvent(String(event.data)));
      };
      socket.onerror = () => {
        voiceSocketReadyRef.current = false;
        setStreamNotice("实时语音连接失败，结束后将使用整段上传。");
      };
      socket.onclose = () => {
        voiceSocketRef.current = null;
        voiceSocketReadyRef.current = false;
      };

      return socket;
    },
    [handleAudioWebSocketEvent, setStreamNotice],
  );

  const sendAudioChunkOverSocket = useCallback((blob: Blob) => {
    const socket = voiceSocketRef.current;
    if (!socket || !voiceSocketReadyRef.current || socket.readyState !== WebSocket.OPEN || blob.size === 0) {
      return;
    }

    socket.send(blob);
    voiceSocketChunkSentRef.current = true;
  }, []);

  const finishRealtimeAudioOrUpload = useCallback(
    (blob: Blob, mimeType: string) => {
      const socket = voiceSocketRef.current;
      if (socket && voiceSocketReadyRef.current && voiceSocketChunkSentRef.current && socket.readyState === WebSocket.OPEN) {
        beginVoiceSocketSend();
        dispatch({ type: "clear_error" });
        socket.send(JSON.stringify({ type: "end" }));
        return;
      }

      closeVoiceSocket();
      void uploadRecordedAudio(blob, mimeType);
    },
    [beginVoiceSocketSend, closeVoiceSocket, uploadRecordedAudio],
  );

  const startRecording = useCallback(async () => {
    if (!sessionId) {
      return;
    }
    if (!navigator.mediaDevices?.getUserMedia || typeof MediaRecorder === "undefined") {
      dispatch({ type: "error", message: "当前浏览器不支持录音上传。" });
      return;
    }

    try {
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
      const mimeType = supportedAudioMimeType();
      const recorder = new MediaRecorder(stream, mimeType ? { mimeType } : undefined);
      audioChunksRef.current = [];
      mediaStreamRef.current = stream;
      mediaRecorderRef.current = recorder;
      openVoiceSocket(sessionId, mimeType || "audio/webm");
      recorder.ondataavailable = (event) => {
        if (event.data.size > 0) {
          audioChunksRef.current.push(event.data);
          sendAudioChunkOverSocket(event.data);
        }
      };
      recorder.onstop = () => {
        const type = recorder.mimeType || mimeType || "audio/webm";
        const blob = new Blob(audioChunksRef.current, { type });
        stopMediaStream();
        finishRealtimeAudioOrUpload(blob, type);
      };
      recorder.start(700);
      dispatch({ type: "recording_started" });
    } catch (error) {
      stopMediaStream();
      dispatch({ type: "error", message: error instanceof Error ? error.message : "无法启动录音，请检查麦克风权限。" });
    }
  }, [finishRealtimeAudioOrUpload, openVoiceSocket, sendAudioChunkOverSocket, sessionId, stopMediaStream, supportedAudioMimeType]);

  const stopRecording = useCallback(() => {
    const recorder = mediaRecorderRef.current;
    if (!recorder || recorder.state === "inactive") {
      dispatch({ type: "idle" });
      return;
    }

    dispatch({ type: "recognizing" });
    recorder.stop();
  }, []);

  const sendRealtimeTranscript = useCallback(
    async (transcript: string) => {
      const content = transcript.trim();
      if (!content) {
        dispatch({ type: "error", message: "实时听写没有生成有效文本，请重新尝试。" });
        return;
      }

      dispatch({ type: "thinking" });
      requestSpeakNextAI();
      const result = await sendVoiceTranscript(content);
      if (result.type === "sent") {
        speakAIReply(result.reply.messageId, result.reply.content);
        return;
      }

      cancelPendingAIReply();
      if (result.type === "invalid" || result.type === "ended" || result.type === "failed") {
        dispatch({ type: "error", message: result.message });
        return;
      }

      dispatch({ type: "idle" });
    },
    [cancelPendingAIReply, requestSpeakNextAI, sendVoiceTranscript, speakAIReply],
  );

  const startRealtimeSpeech = useCallback(() => {
    if (!sessionId || !isRealtimeSpeechSupported()) {
      return false;
    }

    realtimeFinalReceivedRef.current = false;
    realtimeStopRequestedRef.current = false;
    realtimeFallbackAttemptedRef.current = false;
    realtimeFallbackOnEndRef.current = false;
    clearSendError();
    setStreamNotice("实时听写已启动，结束后会发送最终转写。");
    dispatch({ type: "recording_started" });
    const realtimeSession = createRealtimeSpeechSession({
      onPartial: (transcript) => dispatch({ type: "partial_transcript", transcript }),
      onFinal: (transcript) => {
        if (realtimeFinalReceivedRef.current) {
          return;
        }
        realtimeFinalReceivedRef.current = true;
        dispatch({ type: "recognizing", transcript });
        void sendRealtimeTranscript(transcript);
      },
      onError: (code) => {
        const shouldUseRecordedAudioFallback = shouldFallbackToRecordedAudio(code, {
          finalReceived: realtimeFinalReceivedRef.current,
          stopRequested: realtimeStopRequestedRef.current,
          fallbackAttempted: realtimeFallbackAttemptedRef.current,
        });
        if (shouldUseRecordedAudioFallback) {
          realtimeFallbackAttemptedRef.current = true;
          realtimeFallbackOnEndRef.current = true;
          dispatch({ type: "clear_error" });
          setStreamNotice("浏览器实时听写启动失败，已自动切换为录音上传。");
          closeRealtimeSpeech();
          return;
        }

        realtimeFallbackOnEndRef.current = false;
        closeRealtimeSpeech();
        dispatch({ type: "error", message: realtimeSpeechErrorMessage(code) });
      },
      onEnd: () => {
        realtimeSpeechRef.current = null;
        if (realtimeFallbackOnEndRef.current) {
          realtimeFallbackOnEndRef.current = false;
          void startRecording();
          return;
        }
        if (!realtimeFinalReceivedRef.current) {
          dispatch({ type: "idle" });
        }
      },
    });
    realtimeSpeechRef.current = realtimeSession;
    try {
      realtimeSession.start();
      return true;
    } catch (error) {
      closeRealtimeSpeech();
      dispatch({ type: "error", message: error instanceof Error ? error.message : "实时听写启动失败，已改用录音上传。" });
      return false;
    }
  }, [clearSendError, closeRealtimeSpeech, sendRealtimeTranscript, sessionId, setStreamNotice, startRecording]);

  const stopRealtimeSpeech = useCallback(() => {
    const realtimeSession = realtimeSpeechRef.current;
    if (!realtimeSession) {
      return false;
    }

    realtimeStopRequestedRef.current = true;
    realtimeFallbackOnEndRef.current = false;
    dispatch({ type: "recognizing" });
    realtimeSession.stop();
    return true;
  }, []);

  const handleVoiceToggle = useCallback(() => {
    if (!sessionId || sessionStatus === "finished" || isFinishing || isPlaybackSpeaking) {
      return;
    }
    if (state.status === "recording") {
      if (stopRealtimeSpeech()) {
        return;
      }
      stopRecording();
      return;
    }
    if (state.status !== "idle" || isSending) {
      return;
    }

    if (startRealtimeSpeech()) {
      return;
    }

    void startRecording();
  }, [
    isFinishing,
    isPlaybackSpeaking,
    isSending,
    sessionId,
    sessionStatus,
    startRealtimeSpeech,
    startRecording,
    state.status,
    stopRealtimeSpeech,
    stopRecording,
  ]);

  const finishPlayback = useCallback(() => {
    dispatch({ type: "idle" });
  }, []);

  return {
    voiceStatus: state.status,
    voiceTranscript: state.transcript,
    voiceError: state.error,
    onVoiceToggle: handleVoiceToggle,
    finishPlayback,
  };
}
