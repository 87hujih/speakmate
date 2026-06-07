import { useEffect, useRef, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { ApiError } from "../api/client";
import { finishTrainingSession, loadTrainingSessionState, sendTrainingText } from "../api/loaders";
import { connectSessionStream, type SessionStreamEvent } from "../api/sessionStream";
import { PageContainer } from "../components/layout/PageContainer";
import { ConversationPanel } from "../components/training/ConversationPanel";
import { RealtimeFeedbackPanel } from "../components/training/RealtimeFeedbackPanel";
import { TaskPanel } from "../components/training/TaskPanel";
import { TrainingHeader } from "../components/training/TrainingHeader";
import { buttonClasses } from "../components/ui/Button";
import type { TrainingSession } from "../types";

function parseRouteSessionId(value: string | undefined) {
  const numeric = Number(value);

  return Number.isInteger(numeric) && numeric > 0 ? numeric : null;
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
  const streamErrorShown = useRef(false);

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
    if (!numericSessionId || session?.status !== "running") {
      return undefined;
    }

    function handleStreamEvent(event: SessionStreamEvent) {
      if (event.type === "ai_message_delta") {
        setStreamNotice("正在接收 AI 回复片段");
        return;
      }
      if (event.type === "ai_message_done" || event.type === "correction_done" || event.type === "score_updated") {
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
        setSendError(error instanceof Error ? error.message : "消息发送失败，请稍后重试。");
      }
    } finally {
      setIsSending(false);
    }
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
        setSendError(error instanceof Error ? error.message : "结束训练失败");
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
          onDraftChange={setDraft}
          onSend={handleSend}
        />
        <RealtimeFeedbackPanel session={session} />
      </div>
    </PageContainer>
  );
}
