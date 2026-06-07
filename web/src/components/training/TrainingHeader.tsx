import { CircleStop, Clock3, LoaderCircle, Radio, ShieldCheck } from "lucide-react";
import type { TrainingSession } from "../../types";
import { Badge } from "../ui/Badge";
import { Button } from "../ui/Button";

interface TrainingHeaderProps {
  session: TrainingSession;
  isFinishing?: boolean;
  onFinish: () => void;
}

export function TrainingHeader({ session, isFinishing = false, onFinish }: TrainingHeaderProps) {
  return (
    <div className="z-20 flex min-h-16 shrink-0 flex-wrap items-center justify-between gap-3 rounded-[24px] border border-line bg-white/95 px-4 py-3 shadow-panel backdrop-blur-xl md:px-5">
      <div className="flex flex-wrap items-center gap-3 md:gap-4">
        <Badge tone="violet">{session.scenario.name}训练</Badge>
        <strong className="text-[15px] text-ink">AI {session.scenario.aiRole} · 口语陪练中</strong>
        <span className="inline-flex items-center gap-2 rounded-full bg-slate-100 px-3 py-1.5 text-xs font-black text-muted">
          <ShieldCheck className="h-4 w-4 text-emerald-500" />
          当前阶段：{session.currentStage}
        </span>
      </div>
      <div className="flex flex-wrap items-center gap-3 md:gap-4">
        <span className="inline-flex items-center gap-2 text-sm font-extrabold text-muted">
          <Radio className="h-4 w-4 text-emerald-500" />
          {session.liveStatus}
        </span>
        <span className="inline-flex items-center gap-2 text-sm font-extrabold text-muted">
          <Clock3 className="h-4 w-4" />
          {session.durationLabel}
        </span>
        <Button
          variant="danger"
          className="h-10 rounded-[16px] px-4 disabled:cursor-not-allowed disabled:opacity-70"
          disabled={isFinishing}
          onClick={onFinish}
        >
          {isFinishing ? <LoaderCircle className="h-4 w-4 animate-spin" /> : <CircleStop className="h-4 w-4" />}
          {session.status === "finished" ? "查看报告" : "结束并生成报告"}
        </Button>
      </div>
    </div>
  );
}
