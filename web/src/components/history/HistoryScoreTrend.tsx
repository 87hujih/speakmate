import { BarChart3 } from "lucide-react";
import type { HistoryScoreTrendPoint } from "../../types";

interface HistoryScoreTrendProps {
  points: HistoryScoreTrendPoint[];
}

export function HistoryScoreTrend({ points }: HistoryScoreTrendProps) {
  const maxScore = Math.max(100, ...points.map((point) => point.averageScore));

  return (
    <article className="rounded-panel border border-line bg-white p-5 shadow-soft">
      <div className="mb-4 flex items-center justify-between gap-3">
        <div className="flex items-center gap-2 text-sm font-black text-ink">
          <BarChart3 className="h-4 w-4 text-brand-blue" />
          评分趋势
        </div>
        <span className="text-xs font-black text-muted">{points.length} 天</span>
      </div>
      {points.length === 0 ? (
        <p className="m-0 text-sm font-semibold text-muted">暂无评分趋势。</p>
      ) : (
        <div className="flex h-32 items-end gap-2">
          {points.map((point) => {
            const height = Math.max(10, Math.round((point.averageScore / maxScore) * 100));
            return (
              <div key={point.date} className="flex min-w-0 flex-1 flex-col items-center gap-2">
                <div className="flex h-24 w-full items-end rounded-xl bg-slate-100 px-1.5">
                  <div
                    className="w-full rounded-t-lg bg-gradient-to-t from-brand-blue to-brand-purple"
                    style={{ height: `${height}%` }}
                    title={`${point.date} · ${point.averageScore}`}
                  />
                </div>
                <span className="max-w-full truncate text-[11px] font-black text-muted">{point.date.slice(5)}</span>
              </div>
            );
          })}
        </div>
      )}
    </article>
  );
}
