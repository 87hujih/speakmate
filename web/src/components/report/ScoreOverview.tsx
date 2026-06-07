import type { ScoreDimension } from "../../types";
import { MetricList } from "../ui/MetricList";
import { Panel } from "../ui/Panel";
import { RadarChart } from "./RadarChart";

interface ScoreOverviewProps {
  scores: ScoreDimension[];
}

export function ScoreOverview({ scores }: ScoreOverviewProps) {
  return (
    <Panel className="p-6">
      <h3 className="m-0 mb-5 text-xl font-black tracking-[-0.02em] text-ink">五维能力评分</h3>
      <RadarChart scores={scores} />
      <MetricList scores={scores} />
    </Panel>
  );
}
