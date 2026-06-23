import type { ScoreDimension } from "../../types";

/** RadarChartProps 定义对应组件接收的属性。 */
interface RadarChartProps {
  scores: ScoreDimension[];
}

const points = [
  { x: 120, y: 22 },
  { x: 214, y: 90 },
  { x: 180, y: 204 },
  { x: 60, y: 204 },
  { x: 26, y: 90 },
];

const labels = [
  { x: 120, y: 14, anchor: "middle" as const },
  { x: 225, y: 94, anchor: "start" as const },
  { x: 186, y: 222, anchor: "start" as const },
  { x: 54, y: 222, anchor: "end" as const },
  { x: 15, y: 94, anchor: "end" as const },
];

/** scalePoint 按分数缩放雷达图坐标点。 */
function scalePoint(point: { x: number; y: number }, score: number) {
  const ratio = Math.min(1, Math.max(0, score / 100));
  return {
    x: 120 + (point.x - 120) * ratio,
    y: 120 + (point.y - 120) * ratio,
  };
}

/** RadarChart 渲染对应的页面或界面组件。 */
export function RadarChart({ scores }: RadarChartProps) {
  const polygon = scores.map((score, index) => scalePoint(points[index], score.score));
  const polygonPoints = polygon.map((point) => `${point.x.toFixed(1)},${point.y.toFixed(1)}`).join(" ");
  const innerPoints = points.map((point) => scalePoint(point, 66));

  return (
    <div className="mx-auto mb-6 mt-2 h-[230px] w-[230px]">
      <svg viewBox="0 0 240 240" role="img" aria-label="五维能力评分雷达图" className="h-full w-full overflow-visible">
        <polygon points={points.map((point) => `${point.x},${point.y}`).join(" ")} fill="none" stroke="#e2e8f0" strokeWidth="2" />
        <polygon points={innerPoints.map((point) => `${point.x},${point.y}`).join(" ")} fill="none" stroke="#e2e8f0" strokeWidth="2" />
        {points.map((point) => (
          <line key={`${point.x}-${point.y}`} x1="120" y1="120" x2={point.x} y2={point.y} stroke="#e2e8f0" />
        ))}
        <polygon points={polygonPoints} fill="rgba(37,99,235,0.18)" stroke="#2563eb" strokeWidth="3" />
        {scores.map((score, index) => (
          <text
            key={score.key}
            x={labels[index].x}
            y={labels[index].y}
            textAnchor={labels[index].anchor}
            fill="#64748b"
            fontSize="12"
            fontWeight="700"
          >
            {score.name.replace("准确度", "").replace("丰富度", "").replace("自然度", "").replace("场景", "")}
          </text>
        ))}
      </svg>
    </div>
  );
}
