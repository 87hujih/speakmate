import { ArrowRight, LoaderCircle, PlayCircle } from "lucide-react";
import type { NextPracticeRecommendation } from "../../types";
import { buttonClasses } from "../ui/Button";

interface HistoryRecommendationCardProps {
  recommendation: NextPracticeRecommendation | null;
  isStarting?: boolean;
  onStart: (recommendation: NextPracticeRecommendation) => void;
}

export function HistoryRecommendationCard({ recommendation, isStarting = false, onStart }: HistoryRecommendationCardProps) {
  if (!recommendation) {
    return (
      <article className="rounded-panel border border-line bg-white p-5 shadow-soft">
        <div className="flex items-center gap-2 text-sm font-black text-muted">
          <PlayCircle className="h-4 w-4 text-brand-blue" />
          下一步建议
        </div>
        <p className="mb-0 mt-3 text-sm font-semibold leading-6 text-muted">暂无新的复练建议。</p>
      </article>
    );
  }

  const actionLabel = recommendation.type === "continue_session" ? "继续训练" : "开始复练";

  return (
    <article className="rounded-panel border border-blue-100 bg-blue-50 p-5 shadow-soft">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div>
          <div className="flex items-center gap-2 text-sm font-black text-blue-700">
            <PlayCircle className="h-4 w-4" />
            下一步建议
          </div>
          <h3 className="mb-2 mt-3 text-lg font-black text-ink">
            {recommendation.scenario?.name ?? "继续当前训练"}
          </h3>
          <p className="m-0 text-sm font-semibold leading-6 text-blue-900/75">{recommendation.reason}</p>
          {recommendation.focus ? <p className="mt-2 text-xs font-black text-blue-700">重点：{recommendation.focus}</p> : null}
        </div>
        <button
          type="button"
          className={buttonClasses("primary", "h-10 rounded-2xl px-4 disabled:cursor-not-allowed disabled:opacity-70")}
          disabled={isStarting}
          onClick={() => onStart(recommendation)}
        >
          {isStarting ? <LoaderCircle className="h-4 w-4 animate-spin" /> : <ArrowRight className="h-4 w-4" />}
          {actionLabel}
        </button>
      </div>
    </article>
  );
}
