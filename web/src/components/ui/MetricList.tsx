import type { ScoreDimension } from "../../types";
import { ProgressBar } from "./ProgressBar";

/** MetricListProps 定义对应组件接收的属性。 */
interface MetricListProps {
  scores: ScoreDimension[];
}

/** MetricList 渲染对应的页面或界面组件。 */
export function MetricList({ scores }: MetricListProps) {
  return (
    <div className="grid gap-3">
      {scores.map((score) => (
        <div key={score.key}>
          <div className="mb-1.5 flex items-center justify-between text-xs font-extrabold text-muted">
            <span>{score.name}</span>
            <b className="text-ink">{score.score}</b>
          </div>
          <ProgressBar value={score.score} />
        </div>
      ))}
    </div>
  );
}
