import { Gauge, Sparkles } from "lucide-react";
import type { TrainingSession } from "../../types";
import { ScoreRing } from "../ui/ScoreRing";
import { CorrectionCard } from "./CorrectionCard";
import { ScoreBar } from "./ScoreBar";

/** RealtimeFeedbackPanelProps 定义对应组件接收的属性。 */
interface RealtimeFeedbackPanelProps {
  session: TrainingSession;
}

/** RealtimeFeedbackPanel 渲染对应的页面或界面组件。 */
export function RealtimeFeedbackPanel({ session }: RealtimeFeedbackPanelProps) {
  const latestCorrection = session.corrections[0];

  return (
    <aside className="coach-scroll h-full min-h-0 overflow-y-auto rounded-panel border border-line bg-white/95 p-3 shadow-panel">
      <div className="mb-3 flex items-center gap-2 text-xs font-black uppercase tracking-[0.08em] text-slate-400">
        <Gauge className="h-4 w-4" />
        实时反馈
      </div>
      <div className="mb-2.5 grid grid-cols-[82px_1fr] items-center gap-3 rounded-[22px] border border-blue-100 bg-gradient-to-br from-blue-50 to-violet-50 p-3">
        <ScoreRing score={session.currentScore} size="sm" />
        <div>
          <div className="flex items-center gap-2 text-sm font-black text-ink">
            <Sparkles className="h-4 w-4 text-brand-purple" />
            当前综合评分
          </div>
          <div className="mt-1 text-2xl font-black tracking-[-0.04em] text-ink">
            {session.currentScore} <span className="text-sm tracking-normal text-muted">/ 100</span>
          </div>
        </div>
      </div>

      <section className="rounded-[22px] border border-line bg-white p-3 shadow-soft">
        <h3 className="m-0 mb-2 text-sm font-black text-ink">分项评分</h3>
        {session.scores.length ? (
          <div className="grid grid-cols-2 gap-x-3 gap-y-2">
            {session.scores.map((score) => (
              <ScoreBar key={score.key} label={score.name} value={score.score} />
            ))}
          </div>
        ) : (
          <p className="m-0 rounded-2xl bg-slate-50 p-3 text-xs font-bold text-muted">发送第一条消息后生成评分。</p>
        )}
      </section>

      <div className="mt-2.5 grid gap-2">
        <div className="rounded-[22px] border border-line bg-white p-2.5 shadow-soft">
          <div className="mb-1.5 flex items-center justify-between gap-3">
            <h4 className="m-0 text-sm font-black text-ink">更自然表达</h4>
            <span className="rounded-full bg-blue-50 px-2.5 py-1 text-[11px] font-black text-blue-700">Upgrade</span>
          </div>
          <p className="m-0 rounded-[18px] bg-slate-50 p-2 text-[10.5px] font-semibold leading-[15px] text-muted">
            {session.naturalExpression}
          </p>
        </div>
        {latestCorrection ? (
          <CorrectionCard correction={latestCorrection} />
        ) : (
          <div className="rounded-[22px] border border-line bg-white p-3 text-xs font-bold leading-5 text-muted shadow-soft">
            暂无纠错。发送文本后，AI 会在这里展示最近一轮表达问题。
          </div>
        )}
      </div>
    </aside>
  );
}
