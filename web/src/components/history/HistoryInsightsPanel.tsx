import { LoaderCircle, RefreshCw } from "lucide-react";
import type { HistoryInsights, NextPracticeRecommendation } from "../../types";
import { buttonClasses } from "../ui/Button";
import { HistoryFrequentErrors } from "./HistoryFrequentErrors";
import { HistoryRecommendationCard } from "./HistoryRecommendationCard";
import { HistoryScenarioTrends } from "./HistoryScenarioTrends";
import { HistoryScoreTrend } from "./HistoryScoreTrend";

interface HistoryInsightsPanelProps {
  insights: HistoryInsights | null;
  insightsDays: 7 | 30;
  isInsightsLoading: boolean;
  insightsError: string;
  isStartingRecommendation?: boolean;
  isPracticeStarting?: boolean;
  onDaysChange: (days: 7 | 30) => void;
  retryInsights: () => void;
  onRecommendationStart: (recommendation: NextPracticeRecommendation) => void;
}

export function HistoryInsightsPanel({
  insights,
  insightsDays,
  isInsightsLoading,
  insightsError,
  isStartingRecommendation = false,
  isPracticeStarting = false,
  onDaysChange,
  retryInsights,
  onRecommendationStart,
}: HistoryInsightsPanelProps) {
  return (
    <section className="grid gap-3.5">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h2 className="m-0 text-xl font-black tracking-[-0.03em] text-ink">学习洞察</h2>
          <p className="m-0 mt-1 text-sm font-semibold text-muted">近 {insightsDays} 天训练复盘</p>
        </div>
        <div className="inline-flex rounded-2xl border border-line bg-white p-1 shadow-soft">
          {[7, 30].map((days) => (
            <button
              key={days}
              type="button"
              aria-pressed={days === insightsDays}
              className={
                days === insightsDays
                  ? "h-9 rounded-xl bg-brand-blue px-4 text-sm font-black text-white"
                  : "h-9 rounded-xl px-4 text-sm font-black text-muted hover:text-ink"
              }
              onClick={() => onDaysChange(days as 7 | 30)}
            >
              {days} 天
            </button>
          ))}
        </div>
      </div>

      {insightsError ? (
        <article className="rounded-panel border border-rose-100 bg-rose-50 p-5 shadow-soft">
          <h3 className="m-0 text-base font-black text-rose-700">洞察加载失败</h3>
          <p className="mb-4 mt-2 text-sm font-bold text-rose-600">{insightsError}</p>
          <button type="button" className={buttonClasses("danger", "h-10 rounded-2xl px-4")} onClick={retryInsights}>
            <RefreshCw className="h-4 w-4" />
            重试
          </button>
        </article>
      ) : isInsightsLoading ? (
        <article className="grid min-h-[180px] place-items-center rounded-panel border border-line bg-white p-8 shadow-soft">
          <div className="inline-flex items-center gap-2 text-sm font-black text-muted">
            <LoaderCircle className="h-5 w-5 animate-spin text-brand-blue" />
            正在加载学习洞察
          </div>
        </article>
      ) : !insights || insights.summary.totalSessions === 0 ? (
        <article className="rounded-panel border border-line bg-white p-5 text-sm font-semibold text-muted shadow-soft">
          暂无可分析的训练数据。
        </article>
      ) : (
        <>
          <HistoryRecommendationCard
            recommendation={insights.nextRecommendation}
            isStarting={isStartingRecommendation}
            isPracticeStarting={isPracticeStarting}
            onStart={onRecommendationStart}
          />
          <HistoryScoreTrend points={insights.scoreTrend} />
          <HistoryScenarioTrends trends={insights.scenarioTrends} />
          <HistoryFrequentErrors errors={insights.frequentErrors} />
        </>
      )}
    </section>
  );
}
