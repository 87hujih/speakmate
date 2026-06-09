import { LoaderCircle, Plus, TrendingDown, TrendingUp } from "lucide-react";
import { Link, useNavigate } from "react-router-dom";
import { useEffect, useRef, useState } from "react";
import { createTrainingSession, loadHistoryInsights, loadHistoryState } from "../api/loaders";
import { HistoryInsightsPanel } from "../components/history/HistoryInsightsPanel";
import { HistorySessionCard } from "../components/history/HistorySessionCard";
import { PageContainer } from "../components/layout/PageContainer";
import { buttonClasses } from "../components/ui/Button";
import { ProgressBar } from "../components/ui/ProgressBar";
import { SectionHeader } from "../components/ui/SectionHeader";
import type { HistoryInsights, HistoryRecord, NextPracticeRecommendation } from "../types";

const pageSize = 10;

function percent(value: number, total: number) {
  if (total <= 0) {
    return 0;
  }

  return Math.round((value / total) * 100);
}

function scoreDeltaText(delta: number | null | undefined) {
  if (delta === null || delta === undefined) {
    return "暂无对比";
  }

  return delta > 0 ? `+${delta}` : String(delta);
}

export function HistoryPage() {
  const navigate = useNavigate();
  const [records, setRecords] = useState<HistoryRecord[]>([]);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState("");
  const [insightsDays, setInsightsDays] = useState<7 | 30>(30);
  const [insights, setInsights] = useState<HistoryInsights | null>(null);
  const [loadedInsightsDays, setLoadedInsightsDays] = useState<7 | 30 | null>(null);
  const [isInsightsLoading, setIsInsightsLoading] = useState(true);
  const [insightsError, setInsightsError] = useState("");
  const [startError, setStartError] = useState("");
  const [startingAction, setStartingAction] = useState<{ type: "repeat" | "recommendation"; id: string } | null>(null);
  const historyRequestID = useRef(0);
  const insightsRequestID = useRef(0);
  const totalPages = Math.max(1, Math.ceil(total / pageSize));
  const displayedInsights = loadedInsightsDays === insightsDays ? insights : null;
  const summary = displayedInsights?.summary;
  const finishedPercent = percent(summary?.finishedSessions ?? 0, summary?.totalSessions ?? 0);
  const reportPercent = percent(summary?.generatedReports ?? 0, summary?.totalSessions ?? 0);
  const ScoreDeltaIcon = (summary?.scoreDelta ?? 0) >= 0 ? TrendingUp : TrendingDown;

  async function loadHistoryList(targetPage = page) {
    const requestID = historyRequestID.current + 1;
    historyRequestID.current = requestID;
    setIsLoading(true);
    setError("");
    try {
      const result = await loadHistoryState(targetPage, pageSize);
      if (historyRequestID.current !== requestID) {
        return;
      }
      setRecords(result.records);
      setPage(result.page);
      setTotal(result.total);
    } catch (loadError) {
      if (historyRequestID.current !== requestID) {
        return;
      }
      setError(loadError instanceof Error ? loadError.message : "历史记录加载失败");
    } finally {
      if (historyRequestID.current === requestID) {
        setIsLoading(false);
      }
    }
  }

  async function loadInsights(days = insightsDays) {
    const requestID = insightsRequestID.current + 1;
    insightsRequestID.current = requestID;
    setIsInsightsLoading(true);
    setInsightsError("");
    setLoadedInsightsDays(null);
    try {
      const result = await loadHistoryInsights(days);
      if (insightsRequestID.current !== requestID) {
        return;
      }
      setInsights(result);
      setLoadedInsightsDays(days);
    } catch (loadError) {
      if (insightsRequestID.current !== requestID) {
        return;
      }
      setInsights(null);
      setLoadedInsightsDays(null);
      setInsightsError(loadError instanceof Error ? loadError.message : "学习洞察加载失败");
    } finally {
      if (insightsRequestID.current === requestID) {
        setIsInsightsLoading(false);
      }
    }
  }

  function retryInsights() {
    void loadInsights(insightsDays);
  }

  async function startPracticeForScenario(scenarioId: number, action: { type: "repeat" | "recommendation"; id: string }) {
    if (startingAction) {
      return;
    }
    setStartingAction(action);
    setStartError("");
    try {
      const session = await createTrainingSession(scenarioId);
      navigate(`/training/${session.session_id}`);
    } catch (startError) {
      setStartError(startError instanceof Error ? startError.message : "创建复练失败");
    } finally {
      setStartingAction(null);
    }
  }

  async function handleRepeat(record: HistoryRecord) {
    await startPracticeForScenario(record.scenario.id, { type: "repeat", id: record.sessionId });
  }

  async function handleRecommendationStart(recommendation: NextPracticeRecommendation) {
    if (startingAction) {
      return;
    }
    setStartingAction({ type: "recommendation", id: recommendation.sessionId });
    setStartError("");
    try {
      if (recommendation.type === "continue_session") {
        navigate(`/training/${recommendation.sessionId}`);
        return;
      }
      if (recommendation.scenario) {
        const session = await createTrainingSession(recommendation.scenario.id);
        navigate(`/training/${session.session_id}`);
        return;
      }

      navigate(`/training/${recommendation.sessionId}`);
    } catch (startError) {
      setStartError(startError instanceof Error ? startError.message : "创建复练失败");
    } finally {
      setStartingAction(null);
    }
  }

  useEffect(() => {
    void loadHistoryList(page);
  }, [page]);

  useEffect(() => {
    void loadInsights(insightsDays);
  }, [insightsDays]);

  return (
    <PageContainer>
      <SectionHeader
        title="历史训练记录"
        description="持续记录每次训练结果，观察口语能力变化趋势。"
        action={
          <Link to="/" className={buttonClasses("primary")}>
            <Plus className="h-4 w-4" />
            开始新训练
          </Link>
        }
      />

      {startError ? (
        <div className="mb-5 rounded-[22px] border border-rose-100 bg-rose-50 p-4 text-sm font-bold text-rose-700" aria-live="polite">
          {startError}
        </div>
      ) : null}

      <div className="grid grid-cols-1 gap-5 xl:grid-cols-[330px_1fr]">
        <aside className="rounded-[30px] border border-line bg-white p-6 shadow-soft">
          <span className="inline-flex items-center gap-2 rounded-full border border-blue-100 bg-blue-50 px-3 py-2 text-[13px] font-black text-blue-700">
            <TrendingUp className="h-4 w-4" />
            Progress
          </span>
          <h2 className="mb-1 mt-5 text-2xl font-black tracking-[-0.03em] text-ink">近 {insightsDays} 天</h2>
          <p className="m-0 mb-5 text-sm leading-6 text-muted">
            已完成 {summary?.finishedSessions ?? 0} 次训练，生成 {summary?.generatedReports ?? 0} 份报告。
          </p>
          <div className="text-[54px] font-black leading-none tracking-[-0.055em] text-ink">
            {summary?.averageScore ?? "--"}
          </div>
          <p className="mb-4 mt-2 text-sm font-bold text-muted">窗口平均综合评分</p>
          <div className="mb-6 inline-flex items-center gap-2 rounded-full bg-slate-100 px-3 py-2 text-xs font-black text-muted">
            <ScoreDeltaIcon className="h-3.5 w-3.5 text-emerald-500" />
            环比 {scoreDeltaText(summary?.scoreDelta)}
          </div>
          <div className="mb-5">
            <div className="mb-1.5 flex items-center justify-between text-xs font-extrabold text-muted">
              <span>已完成训练占比</span>
              <b>{finishedPercent}%</b>
            </div>
            <ProgressBar value={finishedPercent} />
          </div>
          <div className="mb-5">
            <div className="mb-1.5 flex items-center justify-between text-xs font-extrabold text-muted">
              <span>报告生成占比</span>
              <b>{reportPercent}%</b>
            </div>
            <ProgressBar value={reportPercent} />
          </div>
          <div className="grid grid-cols-2 gap-3 border-t border-line pt-4">
            <div>
              <p className="m-0 text-xs font-black text-muted">总训练</p>
              <strong className="text-2xl font-black text-ink">{summary?.totalSessions ?? 0}</strong>
            </div>
            <div>
              <p className="m-0 text-xs font-black text-muted">已评分</p>
              <strong className="text-2xl font-black text-ink">{summary?.scoredSessions ?? 0}</strong>
            </div>
          </div>
        </aside>

        <section className="grid content-start gap-5">
          <HistoryInsightsPanel
            insights={displayedInsights}
            insightsDays={insightsDays}
            isInsightsLoading={isInsightsLoading}
            insightsError={insightsError}
            isStartingRecommendation={startingAction?.type === "recommendation"}
            isPracticeStarting={startingAction !== null}
            onDaysChange={setInsightsDays}
            retryInsights={retryInsights}
            onRecommendationStart={handleRecommendationStart}
          />

          <div className="grid content-start gap-3.5">
            {error ? (
              <div className="rounded-panel border border-rose-100 bg-rose-50 p-6 text-center shadow-soft">
                <h3 className="m-0 text-xl font-black text-rose-700">历史记录加载失败</h3>
                <p className="mt-2 text-sm font-bold text-rose-600">{error}</p>
                <button type="button" className={buttonClasses("danger", "mt-5")} onClick={() => loadHistoryList()}>
                  重试
                </button>
              </div>
            ) : isLoading ? (
              <div className="grid min-h-[260px] place-items-center rounded-panel border border-line bg-white p-8 shadow-soft">
                <div className="inline-flex items-center gap-2 text-sm font-black text-muted">
                  <LoaderCircle className="h-5 w-5 animate-spin text-brand-blue" />
                  正在加载历史记录
                </div>
              </div>
            ) : records.length === 0 ? (
              <div className="rounded-panel border border-line bg-white p-8 text-center shadow-soft">
                <h3 className="m-0 text-xl font-black text-ink">暂无历史训练</h3>
                <p className="mt-2 text-sm font-semibold text-muted">完成一次场景训练后，这里会显示记录和报告状态。</p>
                <Link to="/" className={buttonClasses("primary", "mt-5")}>
                  选择训练场景
                </Link>
              </div>
            ) : (
              <>
                {records.map((record) => (
                  <HistorySessionCard
                    key={record.sessionId}
                    record={record}
                    onRepeat={handleRepeat}
                    isRepeating={startingAction?.type === "repeat" && startingAction.id === record.sessionId}
                    isPracticeStarting={startingAction !== null}
                  />
                ))}
                <div className="mt-2 flex items-center justify-between rounded-panel border border-line bg-white p-4 shadow-soft">
                  <button
                    type="button"
                    className={buttonClasses("ghost", "h-10 rounded-2xl px-4 disabled:cursor-not-allowed disabled:opacity-60")}
                    disabled={page <= 1 || isLoading}
                    onClick={() => setPage((current) => Math.max(1, current - 1))}
                  >
                    上一页
                  </button>
                  <span className="text-sm font-black text-muted">
                    第 {page} / {totalPages} 页
                  </span>
                  <button
                    type="button"
                    className={buttonClasses("ghost", "h-10 rounded-2xl px-4 disabled:cursor-not-allowed disabled:opacity-60")}
                    disabled={page >= totalPages || isLoading}
                    onClick={() => setPage((current) => Math.min(totalPages, current + 1))}
                  >
                    下一页
                  </button>
                </div>
              </>
            )}
          </div>
        </section>
      </div>
    </PageContainer>
  );
}
