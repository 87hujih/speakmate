import type { Correction } from "../../types";
import { cn } from "../../utils/cn";

interface CorrectionCardProps {
  correction: Correction;
}

const tagTone = {
  grammar: "bg-rose-50 text-rose-600",
  expression: "bg-orange-50 text-orange-600",
  vocabulary: "bg-emerald-50 text-emerald-700",
};

export function CorrectionCard({ correction }: CorrectionCardProps) {
  return (
    <div className="rounded-[22px] border border-line bg-white p-3 shadow-soft">
      <div className="mb-2 flex items-center justify-between gap-3">
        <div>
          <div className="mb-1 text-[11px] font-black uppercase tracking-[0.08em] text-slate-400">最新纠错</div>
          <h4 className="m-0 text-sm font-black text-ink">{correction.title}</h4>
        </div>
        <span className={cn("rounded-full px-2.5 py-1 text-[11px] font-black", tagTone[correction.category])}>
          {correction.category === "grammar" ? "Grammar" : correction.category === "expression" ? "Expression" : "Vocabulary"}
        </span>
      </div>
      {correction.issues?.length ? (
        <div className="mb-2">
          <strong className="mb-1 block text-xs font-black text-ink">问题</strong>
          <div className="flex flex-wrap gap-1.5">
            {correction.issues.map((issue) => (
              <span key={issue} className="rounded-full bg-slate-50 px-2 py-0.5 text-[11px] font-bold text-muted">
                {issue}
              </span>
            ))}
          </div>
        </div>
      ) : null}
      <div className="grid grid-cols-2 gap-2 text-[11px] leading-4">
        <div className="rounded-[18px] bg-rose-50 px-2.5 py-1.5 text-rose-700">
          <strong className="mb-1 block text-[11px] uppercase tracking-[0.08em] text-rose-500">原句</strong>
          {correction.original}
        </div>
        <div className="rounded-[18px] bg-emerald-50 px-2.5 py-1.5 text-emerald-700">
          <strong className="mb-1 block text-[11px] uppercase tracking-[0.08em] text-emerald-600">建议</strong>
          {correction.suggestion}
        </div>
      </div>
    </div>
  );
}
