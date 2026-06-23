import { cn } from "../../utils/cn";

/** ProgressBarProps 定义对应组件接收的属性。 */
interface ProgressBarProps {
  value: number;
  className?: string;
}

/** ProgressBar 渲染对应的页面或界面组件。 */
export function ProgressBar({ value, className }: ProgressBarProps) {
  const safeValue = Math.min(100, Math.max(0, value));

  return (
    <div className={cn("h-2 overflow-hidden rounded-full bg-slate-200", className)}>
      <div
        className="h-full rounded-full bg-gradient-to-r from-brand-blue to-brand-purple"
        style={{ width: `${safeValue}%` }}
      />
    </div>
  );
}
