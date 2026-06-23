import { ProgressBar } from "../ui/ProgressBar";

/** ScoreBarProps 定义对应组件接收的属性。 */
interface ScoreBarProps {
  label: string;
  value: number;
}

/** ScoreBar 渲染对应的页面或界面组件。 */
export function ScoreBar({ label, value }: ScoreBarProps) {
  return (
    <div>
      <div className="mb-1 flex items-center justify-between text-xs font-extrabold text-muted">
        <span>{label}</span>
        <b className="text-ink">{value}</b>
      </div>
      <ProgressBar value={value} className="h-1.5" />
    </div>
  );
}
