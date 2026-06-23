import { useEffect, useRef, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { PageContainer } from "../components/layout/PageContainer";
import { ConversationPanel } from "../components/training/ConversationPanel";
import { RealtimeFeedbackPanel } from "../components/training/RealtimeFeedbackPanel";
import { TaskPanel } from "../components/training/TaskPanel";
import { TrainingHeader } from "../components/training/TrainingHeader";
import { buttonClasses } from "../components/ui/Button";
import { useSessionStream } from "../hooks/useSessionStream";
import { useTTSPlayback } from "../hooks/useTTSPlayback";
import { useTrainingSession } from "../hooks/useTrainingSession";
import { useVoiceInput } from "../hooks/useVoiceInput";

/** TrainingPage 渲染对应的页面或界面组件。 */
export function TrainingPage() {
  const { sessionId } = useParams();
  const navigate = useNavigate();
  const [streamNotice, setStreamNotice] = useState("");
  const finishPlaybackRef = useRef<() => void>(() => undefined);

  const training = useTrainingSession(sessionId, navigate);
  const tts = useTTSPlayback({
    setStreamNotice,
    onPlaybackEnd: () => finishPlaybackRef.current(),
  });
  const voice = useVoiceInput({
    sessionId: training.numericSessionId,
    sessionStatus: training.session?.status,
    isSending: training.isSending,
    isFinishing: training.isFinishing,
    isPlaybackSpeaking: tts.isSpeaking,
    setStreamNotice,
    reload: training.reload,
    clearSendError: training.clearSendError,
    beginVoiceSocketSend: training.beginVoiceSocketSend,
    endVoiceSocketSend: training.endVoiceSocketSend,
    sendVoiceTranscript: training.sendVoiceTranscript,
    uploadVoiceAudio: training.uploadVoiceAudio,
    requestSpeakNextAI: tts.requestSpeakNextAI,
    cancelPendingAIReply: tts.cancelPendingAIReply,
    speakAIReply: tts.speakAIReply,
  });

  useEffect(() => {
    finishPlaybackRef.current = voice.finishPlayback;
  }, [voice.finishPlayback]);

  useSessionStream({
    sessionId: training.numericSessionId,
    sessionStatus: training.session?.status,
    appendAIStreamDelta: training.appendStreamDelta,
    reload: training.reload,
    speakAIReply: tts.speakAIReply,
    setStreamNotice,
  });

  const voiceStatus = tts.isSpeaking ? "speaking" : voice.voiceStatus;

  function handleSend() {
    if (voiceStatus !== "idle") {
      return;
    }

    void training.sendDraft();
  }

  if (training.isLoading && !training.session) {
    return (
      <PageContainer size="full" className="grid min-h-[calc(100vh-72px)] place-items-center">
        <div className="rounded-panel border border-line bg-white p-8 text-center shadow-panel">
          <h2 className="m-0 text-xl font-black text-ink">正在加载训练</h2>
          <p className="mt-2 text-sm font-semibold text-muted">请稍候，正在同步 Session、消息和反馈。</p>
        </div>
      </PageContainer>
    );
  }

  if (training.loadError || !training.session) {
    return (
      <PageContainer>
        <section className="rounded-panel border border-rose-100 bg-rose-50 p-8 text-center shadow-soft">
          <h2 className="m-0 text-2xl font-black text-rose-700">训练加载失败</h2>
          <p className="mt-2 text-sm font-bold text-rose-600">{training.loadError || "训练不存在或服务已重启。"}</p>
          <div className="mt-5 flex justify-center gap-3">
            <button type="button" className={buttonClasses("danger")} onClick={() => void training.reload()}>
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
    <PageContainer size="full" className="flex min-h-[calc(100vh-72px)] flex-col gap-3 xl:h-[calc(100vh-72px)] xl:min-h-0 xl:overflow-hidden">
      <TrainingHeader session={training.session} isFinishing={training.isFinishing} onFinish={training.finish} />
      <div className="grid min-h-0 flex-1 grid-cols-1 gap-3 xl:grid-cols-[260px_minmax(0,1fr)_360px] xl:overflow-hidden">
        <TaskPanel
          scenario={training.session.scenario}
          tasks={training.session.tasks}
          progress={training.session.progress}
          focusTags={training.session.focusTags}
        />
        <ConversationPanel
          currentStage={training.session.currentStage}
          messages={training.session.messages}
          turnCount={training.session.turnCount}
          draft={training.draft}
          isSending={training.isSending}
          isDisabled={training.session.status === "finished"}
          error={training.sendError}
          streamNotice={streamNotice}
          voiceStatus={voiceStatus}
          voiceTranscript={voice.voiceTranscript}
          voiceError={voice.voiceError}
          isVoiceDisabled={training.session.status === "finished" || training.isFinishing || (training.isSending && voiceStatus !== "recording")}
          onDraftChange={training.setDraft}
          onSend={handleSend}
          onVoiceToggle={voice.onVoiceToggle}
        />
        <RealtimeFeedbackPanel session={training.session} />
      </div>
    </PageContainer>
  );
}
