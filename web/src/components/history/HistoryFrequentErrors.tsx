import { AlertTriangle } from "lucide-react";
import type { FrequentErrorInsight } from "../../types";

interface HistoryFrequentErrorsProps {
  errors: FrequentErrorInsight[];
}

export function HistoryFrequentErrors({ errors }: HistoryFrequentErrorsProps) {
  const visibleErrors = errors.slice(0, 5);

  return (
    <section className="grid gap-3">
      <div className="flex items-center gap-2 text-sm font-black text-ink">
        <AlertTriangle className="h-4 w-4 text-amber-500" />
        高频错误
      </div>
      {visibleErrors.length === 0 ? (
        <article className="rounded-panel border border-line bg-white p-5 text-sm font-semibold text-muted shadow-soft">
          暂未发现高频错误。
        </article>
      ) : (
        <div className="grid gap-3">
          {visibleErrors.map((error) => (
            <article key={error.key} className="rounded-panel border border-line bg-white p-5 shadow-soft">
              <div className="flex flex-wrap items-start justify-between gap-3">
                <div>
                  <h3 className="m-0 text-base font-black text-ink">{error.title}</h3>
                  {error.suggestion ? <p className="m-0 mt-1 text-sm font-bold text-emerald-700">建议：{error.suggestion}</p> : null}
                </div>
                <span className="rounded-full bg-amber-50 px-3 py-1 text-xs font-black text-amber-700">
                  {error.count} 次
                </span>
              </div>
              <p className="mb-0 mt-3 text-sm leading-6 text-muted">{error.latestEvidence}</p>
              <p className="mb-0 mt-2 text-xs font-bold text-muted">最近出现：{error.lastSeenAt}</p>
            </article>
          ))}
        </div>
      )}
    </section>
  );
}
