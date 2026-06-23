import { ArrowRight, BriefcaseBusiness, LoaderCircle, MessageCircle, RotateCcw, Utensils, UsersRound, type LucideIcon } from "lucide-react";
import { Link } from "react-router-dom";
import type { HistoryRecord, ScenarioCode } from "../../types";
import { buttonClasses } from "../ui/Button";

const scenarioIconMap: Record<string, LucideIcon> = {
  interview: BriefcaseBusiness,
  restaurant: Utensils,
  meeting: UsersRound,
};

/** HistorySessionCardProps 定义对应组件接收的属性。 */
interface HistorySessionCardProps {
  record: HistoryRecord;
  onRepeat?: (record: HistoryRecord) => void;
  isRepeating?: boolean;
  isPracticeStarting?: boolean;
}

/** HistorySessionCard 渲染对应的页面或界面组件。 */
export function HistorySessionCard({ record, onRepeat, isRepeating = false, isPracticeStarting = false }: HistorySessionCardProps) {
  const Icon = scenarioIconMap[record.scenario.code as ScenarioCode] ?? MessageCircle;
  const actionTo = record.status === "running" ? `/training/${record.sessionId}` : `/report/${record.sessionId}`;
  const actionLabel = record.status === "running" ? "继续训练" : record.reportStatus === "generated" ? "查看报告" : "生成报告";

  return (
    <article className="flex flex-col justify-between gap-5 rounded-panel border border-line bg-white p-5 shadow-soft md:flex-row md:items-center">
      <div className="flex min-w-0 items-center gap-4">
        <div className="grid h-[52px] w-[52px] place-items-center rounded-[18px] bg-blue-50 text-brand-blue">
          <Icon className="h-6 w-6" />
        </div>
        <div className="min-w-0">
          <div className="flex flex-wrap items-center gap-2">
            <h3 className="m-0 text-lg font-black text-ink">{record.scenario.name}</h3>
            <span className="rounded-full bg-slate-100 px-2.5 py-1 text-[11px] font-black text-muted">
              {record.status === "running" ? "进行中" : "已结束"}
            </span>
            <span className="rounded-full bg-blue-50 px-2.5 py-1 text-[11px] font-black text-blue-700">
              {record.reportStatus === "generated" ? "报告已生成" : "报告未生成"}
            </span>
          </div>
          <p className="mt-1 text-sm text-muted">
            {record.trainedAt} · {record.durationLabel} · {record.turnCount} 轮 · 主要问题：{record.majorProblem}
          </p>
        </div>
      </div>
      <div className="flex shrink-0 flex-wrap items-center gap-3 md:justify-end">
        <span className="text-[28px] font-black tracking-[-0.04em] text-brand-blue">{record.score || "未评分"}</span>
        <Link to={actionTo} className={buttonClasses("ghost", "h-10 rounded-2xl px-4")}>
          {actionLabel}
          <ArrowRight className="h-4 w-4" />
        </Link>
        {record.status === "finished" ? (
          <button
            type="button"
            className={buttonClasses("soft", "h-10 rounded-2xl px-4 disabled:cursor-not-allowed disabled:opacity-70")}
            disabled={isRepeating || isPracticeStarting || !onRepeat}
            onClick={() => onRepeat?.(record)}
          >
            {isRepeating ? <LoaderCircle className="h-4 w-4 animate-spin" /> : <RotateCcw className="h-4 w-4" />}
            再练一次同场景
          </button>
        ) : null}
      </div>
    </article>
  );
}
