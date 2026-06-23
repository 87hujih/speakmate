import type { Correction } from "../../types";
import { cn } from "../../utils/cn";

/** ErrorAnalysisCardProps 定义对应组件接收的属性。 */
interface ErrorAnalysisCardProps {
  correction: Correction;
}

const tagTone = {
  grammar: "bg-rose-50 text-rose-600",
  expression: "bg-orange-50 text-orange-600",
  vocabulary: "bg-emerald-50 text-emerald-700",
};

/** ErrorAnalysisCard 渲染对应的页面或界面组件。 */
export function ErrorAnalysisCard({ correction }: ErrorAnalysisCardProps) {
  return (
    <div className="rounded-[20px] border border-line bg-white p-4">
      <div className="mb-3 flex items-center justify-between gap-3">
        <strong className="text-sm text-ink">{correction.title}</strong>
        <span className={cn("rounded-full px-2.5 py-1 text-[11px] font-black", tagTone[correction.category])}>
          {correction.category === "grammar" ? "Grammar" : correction.category === "expression" ? "Expression" : "Vocabulary"}
        </span>
      </div>
      <div className="grid grid-cols-2 gap-3">
        <div className="rounded-2xl bg-slate-50 p-3 text-[13px] leading-6 text-muted">
          <strong className="mb-1 block text-ink">原句</strong>
          {correction.original}
        </div>
        <div className="rounded-2xl bg-emerald-50 p-3 text-[13px] leading-6 text-emerald-700">
          <strong className="mb-1 block text-emerald-900">建议</strong>
          {correction.suggestion}
        </div>
      </div>
      {correction.explanation ? (
        <div className="mt-3 rounded-2xl bg-amber-50 p-3 text-[13px] leading-6 text-amber-800">
          <strong className="mb-1 block text-amber-950">分析依据</strong>
          {correction.explanation}
        </div>
      ) : null}
    </div>
  );
}
