/** ScoreRingProps 定义对应组件接收的属性。 */
interface ScoreRingProps {
  score: number;
  size?: "sm" | "md" | "lg";
}

const sizeClasses = {
  sm: "h-20 w-20 text-2xl",
  md: "h-[92px] w-[92px] text-[26px]",
  lg: "h-32 w-32 text-4xl",
};

const innerInset = {
  sm: "inset-2",
  md: "inset-2",
  lg: "inset-3",
};

/** ScoreRing 渲染对应的页面或界面组件。 */
export function ScoreRing({ score, size = "md" }: ScoreRingProps) {
  const safeScore = Math.min(100, Math.max(0, score));

  return (
    <div
      className={`relative grid shrink-0 place-items-center rounded-full ${sizeClasses[size]}`}
      style={{
        background: `conic-gradient(#2563eb 0 ${safeScore}%, #dbeafe ${safeScore}% 100%)`,
      }}
    >
      <div className={`absolute rounded-full bg-white ${innerInset[size]}`} />
      <strong className="relative z-10 font-black tracking-[-0.04em] text-ink">{safeScore}</strong>
    </div>
  );
}
