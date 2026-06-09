import { TrendingDown, TrendingUp } from "lucide-react";
import type { ScenarioTrend } from "../../types";

interface HistoryScenarioTrendsProps {
  trends: ScenarioTrend[];
}

function signedDelta(delta: number | null) {
  if (delta === null) {
    return "--";
  }

  return delta > 0 ? `+${delta}` : String(delta);
}

export function HistoryScenarioTrends({ trends }: HistoryScenarioTrendsProps) {
  return (
    <section className="grid gap-3">
      <div className="flex items-center gap-2 text-sm font-black text-ink">
        <TrendingUp className="h-4 w-4 text-brand-blue" />
        场景表现
      </div>
      {trends.length === 0 ? (
        <article className="rounded-panel border border-line bg-white p-5 text-sm font-semibold text-muted shadow-soft">
          暂无场景评分数据。
        </article>
      ) : (
        <div className="grid gap-3 md:grid-cols-2">
          {trends.map((trend) => {
            const isUp = (trend.scoreDelta ?? 0) >= 0;
            const DeltaIcon = isUp ? TrendingUp : TrendingDown;
            return (
              <article key={trend.scenario.id} className="rounded-panel border border-line bg-white p-5 shadow-soft">
                <div className="flex items-start justify-between gap-3">
                  <div>
                    <h3 className="m-0 text-base font-black text-ink">{trend.scenario.name}</h3>
                    <p className="m-0 mt-1 text-xs font-bold text-muted">{trend.lastTrainedAt}</p>
                  </div>
                  <span className="text-2xl font-black tracking-[-0.04em] text-brand-blue">
                    {trend.averageScore ?? "--"}
                  </span>
                </div>
                <div className="mt-4 grid grid-cols-3 gap-2 text-xs font-black text-muted">
                  <span>首分 {trend.firstScore ?? "--"}</span>
                  <span>最近 {trend.latestScore ?? "--"}</span>
                  <span className={isUp ? "text-emerald-600" : "text-rose-600"}>
                    <DeltaIcon className="mr-1 inline h-3.5 w-3.5" />
                    {signedDelta(trend.scoreDelta)}
                  </span>
                </div>
              </article>
            );
          })}
        </div>
      )}
    </section>
  );
}
