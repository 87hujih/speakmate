import { useEffect, useRef, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { createAudioWebSocketUrl, parseAudioWebSocketEvent, type AudioWebSocketEvent } from "../api/audioWebSocket";
import { ApiError } from "../api/client";
import { demoErrorMessage } from "../api/errors";
import { finishTrainingSession, loadTrainingSessionState, sendTrainingAudio, sendTrainingText } from "../api/loaders";
import { connectSessionStream, type SessionStreamEvent } from "../api/sessionStream";
import { PageContainer } from "../components/layout/PageContainer";
import { ConversationPanel } from "../components/training/ConversationPanel";
import { RealtimeFeedbackPanel } from "../components/training/RealtimeFeedbackPanel";
import { TaskPanel } from "../components/training/TaskPanel";
import { TrainingHeader } from "../components/training/TrainingHeader";
import { buttonClasses } from "../components/ui/Button";
import type { ChatMessage, TrainingSession, VoiceStatus } from "../types";
import { extensionForAudioMimeType, selectSupportedAudioMimeType } from "../utils/audioMime";
import { createRealtimeSpeechSession, isRealtimeSpeechSupported, type RealtimeSpeechSession } from "../utils/realtimeSpeech";
import { createTextToSpeechPlayer, type TextToSpeechPlayer } from "../utils/tts";

function parseRouteSessionId(value: string | undefined) {
  const numeric = Number(value);

  return Number.isInteger(numeric) && numeric > 0 ? numeric : null;
}

const streamingAIMessageId = -1;

function appendAIStreamDelta(session: TrainingSession, event: Extract<SessionStreamEvent, { type: "ai_message_delta" }>): TrainingSession {
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

export function TrainingPage() {
  const { sessionId } = useParams();
  const navigate = useNavigate();
  const numericSessionId = parseRouteSessionId(sessionId);
  const [session, setSession] = useState<TrainingSession | null>(null);
  const [draft, setDraft] = useState("");
  const [isLoading, setIsLoading] = useState(true);
  const [isSending, setIsSending] = useState(false);
  const [isFinishing, setIsFinishing] = useState(false);
  const [loadError, setLoadError] = useState("");
  const [sendError, setSendError] = useState("");
  const [streamNotice, setStreamNotice] = useState("");
  const [voiceStatus, setVoiceStatus] = useState<VoiceStatus>("idle");
  const [voiceTranscript, setVoiceTranscript] = useState("");
  const [voiceError, setVoiceError] = useState("");
  const streamErrorShown = useRef(false);
  const mediaRecorderRef = useRef<MediaRecorder | null>(null);
  const mediaStreamRef = useRef<MediaStream | null>(null);
  const voiceSocketRef = useRef<WebSocket | null>(null);
  const voiceSocketReadyRef = useRef(false);
  const voiceSocketChunkSentRef = useRef(false);
  const realtimeSpeechRef = useRef<RealtimeSpeechSession | null>(null);
  const realtimeFinalReceivedRef = useRef(false);
  const shouldSpeakNextAIRef = useRef(false);
  const spokenAIMessageIdsRef = useRef<Set<number>>(new Set());
  const ttsPlayerRef = useRef<TextToSpeechPlayer | null>(null);
  const audioChunksRef = useRef<Blob[]>([]);

  async function reload(nextGoal?: string) {
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
  }

  useEffect(() => {
    void reload();
  }, [numericSessionId]);

  useEffect(() => {
    return () => {
      const recorder = mediaRecorderRef.current;
      if (recorder && recorder.state !== "inactive") {
        recorder.onstop = null;
        recorder.stop();
      }
      closeRealtimeSpeech();
      ttsPlayerRef.current?.cancel();
      closeVoiceSocket();
      stopMediaStream();
    };
  }, []);

  useEffect(() => {
    if (!numericSessionId || session?.status !== "running") {
      return undefined;
    }

    function handleStreamEvent(event: SessionStreamEvent) {
      if (event.type === "ai_message_delta") {
        setStreamNotice("正在接收 AI 回复片段");
        setSession((current) => (current ? appendAIStreamDelta(current, event) : current));
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

    return connectSessionStream(numericSessionId, {
      onEvent: handleStreamEvent,
      onError: (message) => {
        if (!streamErrorShown.current) {
          streamErrorShown.current = true;
          setStreamNotice(message);
        }
      },
    });
  }, [numericSessionId, session?.status]);

  async function handleSend() {
    if (!numericSessionId || isSending || voiceStatus !== "idle") {
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
  }

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

  function textToSpeechPlayer() {
    if (!ttsPlayerRef.current) {
      ttsPlayerRef.current = createTextToSpeechPlayer({
        onUnavailable: () => {
          setStreamNotice("当前浏览器不支持 AI 语音播放，已保留文本回复。");
          setVoiceStatus("idle");
        },
        onEnd: () => {
          setVoiceStatus((current) => (current === "speaking" ? "idle" : current));
        },
      });
    }

    return ttsPlayerRef.current;
  }

  function speakAIReply(messageId: number, content: string) {
    const reply = content.trim();
    if (!shouldSpeakNextAIRef.current || !reply) {
      return;
    }
    if (messageId > 0 && spokenAIMessageIdsRef.current.has(messageId)) {
      return;
    }
    if (messageId > 0) {
      spokenAIMessageIdsRef.current.add(messageId);
    }
    shouldSpeakNextAIRef.current = false;
    setVoiceStatus("speaking");
    textToSpeechPlayer().speak(reply);
  }

  async function sendRealtimeTranscript(transcript: string) {
    if (!numericSessionId || isSending) {
      return;
    }

    const content = transcript.trim();
    if (!content) {
      setVoiceStatus("idle");
      setVoiceError("实时听写没有生成有效文本，请重新尝试。");
      return;
    }

    setIsSending(true);
    setSendError("");
    setVoiceError("");
    setVoiceStatus("thinking");
    shouldSpeakNextAIRef.current = true;
    try {
      const result = await sendTrainingText(numericSessionId, content);
      setSession(result.session);
      speakAIReply(result.result.ai_message.id, result.result.ai_message.content);
    } catch (error) {
      shouldSpeakNextAIRef.current = false;
      if (error instanceof ApiError && error.code === 2004) {
        setVoiceError("本次训练已结束，不能继续发送语音。");
        void reload();
      } else {
        setVoiceError(demoErrorMessage(error, "实时语音发送失败，请稍后重试。"));
      }
      setVoiceStatus("idle");
    } finally {
      setIsSending(false);
    }
  }

  function closeRealtimeSpeech() {
    realtimeSpeechRef.current?.abort();
    realtimeSpeechRef.current = null;
    realtimeFinalReceivedRef.current = false;
  }

  function startRealtimeSpeech() {
    if (!numericSessionId || !isRealtimeSpeechSupported()) {
      return false;
    }

    realtimeFinalReceivedRef.current = false;
    setVoiceTranscript("");
    setVoiceError("");
    setSendError("");
    setStreamNotice("实时听写已启动，结束后会发送最终转写。");
    const session = createRealtimeSpeechSession({
      onPartial: setVoiceTranscript,
      onFinal: (transcript) => {
        if (realtimeFinalReceivedRef.current) {
          return;
        }
        realtimeFinalReceivedRef.current = true;
        setVoiceTranscript(transcript);
        setVoiceStatus("recognizing");
        void sendRealtimeTranscript(transcript);
      },
      onError: (code) => {
        closeRealtimeSpeech();
        setVoiceStatus("idle");
        setVoiceError(realtimeSpeechErrorMessage(code));
      },
      onEnd: () => {
        realtimeSpeechRef.current = null;
        if (!realtimeFinalReceivedRef.current) {
          setVoiceStatus("idle");
        }
      },
    });
    realtimeSpeechRef.current = session;
    setVoiceStatus("recording");
    try {
      session.start();
      return true;
    } catch (error) {
      closeRealtimeSpeech();
      setVoiceStatus("idle");
      setVoiceError(error instanceof Error ? error.message : "实时听写启动失败，已改用录音上传。");
      return false;
    }
  }

  function stopRealtimeSpeech() {
    const session = realtimeSpeechRef.current;
    if (!session) {
      return false;
    }

    setVoiceStatus("recognizing");
    session.stop();
    return true;
  }

  function stopMediaStream() {
    mediaStreamRef.current?.getTracks().forEach((track) => track.stop());
    mediaStreamRef.current = null;
  }

  function closeVoiceSocket() {
    const socket = voiceSocketRef.current;
    voiceSocketRef.current = null;
    voiceSocketReadyRef.current = false;
    voiceSocketChunkSentRef.current = false;
    if (socket && socket.readyState === WebSocket.OPEN) {
      socket.close();
    }
  }

  function supportedAudioMimeType() {
    if (typeof MediaRecorder === "undefined" || typeof MediaRecorder.isTypeSupported !== "function") {
      return "";
    }

    return selectSupportedAudioMimeType((candidate) => MediaRecorder.isTypeSupported(candidate));
  }

  async function uploadRecordedAudio(blob: Blob, mimeType: string) {
    if (!numericSessionId) {
      return;
    }
    if (blob.size === 0) {
      setVoiceStatus("idle");
      setVoiceError("录音为空，请重新录制。");
      return;
    }

    const fileType = mimeType || blob.type || "audio/webm";
    const file = new File([blob], `answer-${Date.now()}.${extensionForAudioMimeType(fileType)}`, { type: fileType });

    setIsSending(true);
    setSendError("");
    setVoiceError("");
    setVoiceTranscript("");
    try {
      const result = await sendTrainingAudio(numericSessionId, file);
      setSession(result.session);
      setVoiceTranscript(result.result.transcript);
    } catch (error) {
      if (error instanceof ApiError && error.code === 2004) {
        setVoiceError("本次训练已结束，不能继续发送语音。");
        void reload();
      } else {
        setVoiceError(voiceUploadErrorMessage(error));
      }
    } finally {
      setIsSending(false);
      setVoiceStatus("idle");
    }
  }

  function handleAudioWebSocketEvent(event: AudioWebSocketEvent) {
    if (event.type === "start") {
      setStreamNotice("实时语音通道已连接。");
      return;
    }
    if (event.type === "partial_transcript") {
      setVoiceTranscript(event.transcript);
      return;
    }
    if (event.type === "final_transcript") {
      setVoiceStatus("thinking");
      setVoiceTranscript(event.transcript);
      void reload(event.next_goal);
      return;
    }
    if (event.type === "correction" || event.type === "score_updated") {
      void reload();
      return;
    }
    if (event.type === "end") {
      setIsSending(false);
      setVoiceStatus("idle");
      closeVoiceSocket();
      return;
    }
    if (event.type === "error") {
      setIsSending(false);
      setVoiceStatus("idle");
      setVoiceError(voiceSocketErrorMessage(event));
    }
  }

  function openVoiceSocket(sessionId: number, mimeType: string) {
    if (typeof WebSocket === "undefined") {
      return null;
    }

    const socket = new WebSocket(createAudioWebSocketUrl(sessionId));
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
  }

  function sendAudioChunkOverSocket(blob: Blob) {
    const socket = voiceSocketRef.current;
    if (!socket || !voiceSocketReadyRef.current || socket.readyState !== WebSocket.OPEN || blob.size === 0) {
      return;
    }

    socket.send(blob);
    voiceSocketChunkSentRef.current = true;
  }

  function finishRealtimeAudioOrUpload(blob: Blob, mimeType: string) {
    const socket = voiceSocketRef.current;
    if (socket && voiceSocketReadyRef.current && voiceSocketChunkSentRef.current && socket.readyState === WebSocket.OPEN) {
      setIsSending(true);
      setSendError("");
      setVoiceError("");
      socket.send(JSON.stringify({ type: "end" }));
      return;
    }

    closeVoiceSocket();
    void uploadRecordedAudio(blob, mimeType);
  }

  async function startRecording() {
    if (!numericSessionId) {
      return;
    }
    if (!navigator.mediaDevices?.getUserMedia || typeof MediaRecorder === "undefined") {
      setVoiceError("当前浏览器不支持录音上传。");
      return;
    }

    try {
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
      const mimeType = supportedAudioMimeType();
      const recorder = new MediaRecorder(stream, mimeType ? { mimeType } : undefined);
      audioChunksRef.current = [];
      mediaStreamRef.current = stream;
      mediaRecorderRef.current = recorder;
      openVoiceSocket(numericSessionId, mimeType || "audio/webm");
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
      setVoiceTranscript("");
      setVoiceError("");
      setVoiceStatus("recording");
    } catch (error) {
      stopMediaStream();
      setVoiceStatus("idle");
      setVoiceError(error instanceof Error ? error.message : "无法启动录音，请检查麦克风权限。");
    }
  }

  function stopRecording() {
    const recorder = mediaRecorderRef.current;
    if (!recorder || recorder.state === "inactive") {
      setVoiceStatus("idle");
      return;
    }

    setVoiceStatus("recognizing");
    recorder.stop();
  }

  function handleVoiceToggle() {
    if (!numericSessionId || session?.status === "finished" || isFinishing) {
      return;
    }
    if (voiceStatus === "recording") {
      if (stopRealtimeSpeech()) {
        return;
      }
      stopRecording();
      return;
    }
    if (voiceStatus !== "idle" || isSending) {
      return;
    }

    if (startRealtimeSpeech()) {
      return;
    }

    void startRecording();
  }

  async function handleFinish() {
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
  }

  if (isLoading && !session) {
    return (
      <PageContainer size="full" className="grid min-h-[calc(100vh-72px)] place-items-center">
        <div className="rounded-panel border border-line bg-white p-8 text-center shadow-panel">
          <h2 className="m-0 text-xl font-black text-ink">正在加载训练</h2>
          <p className="mt-2 text-sm font-semibold text-muted">请稍候，正在同步 Session、消息和反馈。</p>
        </div>
      </PageContainer>
    );
  }

  if (loadError || !session) {
    return (
      <PageContainer>
        <section className="rounded-panel border border-rose-100 bg-rose-50 p-8 text-center shadow-soft">
          <h2 className="m-0 text-2xl font-black text-rose-700">训练加载失败</h2>
          <p className="mt-2 text-sm font-bold text-rose-600">{loadError || "训练不存在或服务已重启。"}</p>
          <div className="mt-5 flex justify-center gap-3">
            <button type="button" className={buttonClasses("danger")} onClick={() => reload()}>
              重试
            </button>
            <Link to="/" className={buttonClasses("ghost")}>
              返回场景选择
            </Link>
          </div>
        </section>
      </PageContainer>
    );
  }

  return (
    <PageContainer size="full" className="flex min-h-[calc(100vh-72px)] flex-col gap-3 overflow-hidden">
      <TrainingHeader session={session} isFinishing={isFinishing} onFinish={handleFinish} />
      <div className="grid min-h-0 flex-1 grid-cols-1 gap-3 xl:grid-cols-[260px_minmax(0,1fr)_360px]">
        <TaskPanel
          scenario={session.scenario}
          tasks={session.tasks}
          progress={session.progress}
          focusTags={session.focusTags}
        />
        <ConversationPanel
          currentStage={session.currentStage}
          messages={session.messages}
          turnCount={session.turnCount}
          draft={draft}
          isSending={isSending}
          isDisabled={session.status === "finished"}
          error={sendError}
          streamNotice={streamNotice}
          voiceStatus={voiceStatus}
          voiceTranscript={voiceTranscript}
          voiceError={voiceError}
          isVoiceDisabled={session.status === "finished" || isFinishing || (isSending && voiceStatus !== "recording")}
          onDraftChange={setDraft}
          onSend={handleSend}
          onVoiceToggle={handleVoiceToggle}
        />
        <RealtimeFeedbackPanel session={session} />
      </div>
    </PageContainer>
  );
}
