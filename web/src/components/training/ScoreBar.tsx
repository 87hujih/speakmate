import { ProgressBar } from "../ui/ProgressBar";

interface ScoreBarProps {
  label: string;
  value: number;
}

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
