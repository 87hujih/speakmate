import { CheckCircle2, Circle, LoaderCircle, Target, UserRound } from "lucide-react";
import type { Scenario, TrainingTask } from "../../types";
import { cn } from "../../utils/cn";
import { ProgressBar } from "../ui/ProgressBar";

/** TaskPanelProps 定义对应组件接收的属性。 */
interface TaskPanelProps {
  scenario: Scenario;
  tasks: TrainingTask[];
  progress: number;
  focusTags: string[];
}

/** TaskPanel 渲染对应的页面或界面组件。 */
export function TaskPanel({ scenario, tasks, progress, focusTags }: TaskPanelProps) {
  return (
    <aside className="h-full min-h-0 overflow-hidden rounded-panel border border-line bg-white/95 p-3.5 shadow-panel">
      <div className="mb-3 rounded-[22px] border border-blue-100 bg-blue-50/80 p-3.5">
        <div className="mb-2.5 flex items-center gap-3">
          <span className="grid h-10 w-10 place-items-center rounded-[16px] bg-white text-brand-blue shadow-soft">
            <Target className="h-5 w-5" />
          </span>
          <div>
            <div className="text-[11px] font-black uppercase tracking-[0.08em] text-blue-500">Scenario</div>
            <h1 className="m-0 text-[19px] font-black tracking-[-0.02em] text-ink">{scenario.name}训练</h1>
          </div>
        </div>
        <div className="grid gap-2.5 text-[13px] leading-5">
          <div>
            <span className="block font-black text-slate-400">AI 角色</span>
            <strong className="text-ink">{scenario.aiRole}</strong>
          </div>
          <div>
            <span className="block font-black text-slate-400">用户目标</span>
            <p className="m-0 font-semibold text-muted">{scenario.userGoal}</p>
          </div>
        </div>
      </div>

      <div className="mb-3 flex items-center gap-2 text-xs font-black uppercase tracking-[0.08em] text-slate-400">
        <UserRound className="h-4 w-4" />
        任务进度
      </div>
      <div className="mb-4 grid gap-1.5">
        {tasks.map((task) => {
          const Icon = task.status === "done" ? CheckCircle2 : task.status === "active" ? LoaderCircle : Circle;

          return (
          <div
            key={task.label}
            className={cn(
              "rounded-[18px] border px-3 py-2.5",
              task.status === "done" && "border-emerald-100 bg-emerald-50 text-emerald-700",
              task.status === "active" && "border-blue-200 bg-white text-blue-700 shadow-soft",
              task.status === "pending" && "border-transparent bg-slate-50",
            )}
          >
            <div className="flex items-center justify-between gap-2">
              <span className="inline-flex min-w-0 items-center gap-2 text-sm font-black">
                <Icon className={cn("h-4 w-4 shrink-0", task.status === "active" && "animate-spin")} />
                <span className="truncate">{task.label}</span>
              </span>
              <span className="shrink-0 text-[11px] font-black text-current/70">
                {task.status === "done" ? "已完成" : task.status === "active" ? "进行中" : "未完成"}
              </span>
            </div>
          </div>
          );
        })}
      </div>

      <div className="rounded-[22px] border border-line bg-white p-3.5 shadow-soft">
        <div className="mb-2 flex items-center justify-between text-xs font-extrabold text-muted">
          <span>场景完成度</span>
          <b className="text-ink">{progress}%</b>
        </div>
        <ProgressBar value={progress} className="h-3" />
      </div>

      <div className="mt-4">
        <div className="mb-3 text-xs font-black uppercase tracking-[0.08em] text-slate-400">训练重点</div>
        <div className="flex flex-wrap gap-2">
          {focusTags.map((tag) => (
            <span key={tag} className="rounded-full border border-line bg-slate-50 px-3 py-2 text-xs font-extrabold text-slate-600">
              {tag}
            </span>
          ))}
        </div>
      </div>
    </aside>
  );
}
