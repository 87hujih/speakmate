import { ArrowRight, BriefcaseBusiness, LoaderCircle, MessageCircle, Utensils, UsersRound, type LucideIcon } from "lucide-react";
import type { Scenario, ScenarioCode } from "../../types";
import { Button, ButtonLink } from "../ui/Button";
import { Badge } from "../ui/Badge";

const scenarioIconMap: Record<string, LucideIcon> = {
  interview: BriefcaseBusiness,
  restaurant: Utensils,
  meeting: UsersRound,
};

/** ScenarioCardProps 定义对应组件接收的属性。 */
interface ScenarioCardProps {
  scenario: Scenario;
  isStarting?: boolean;
  onStart?: (scenario: Scenario) => void;
}

/** ScenarioCard 渲染对应的页面或界面组件。 */
export function ScenarioCard({ scenario, isStarting = false, onStart }: ScenarioCardProps) {
  const Icon = scenarioIconMap[scenario.code as ScenarioCode] ?? MessageCircle;
  const startContent = (
    <>
      {isStarting ? <LoaderCircle className="h-4 w-4 animate-spin" /> : <ArrowRight className="h-4 w-4" />}
      {isStarting ? "创建训练中" : "开始训练"}
    </>
  );

  return (
    <article className="group relative flex h-full flex-col overflow-hidden rounded-[28px] border border-line bg-white/90 p-6 shadow-soft transition duration-200 hover:-translate-y-1 hover:shadow-panel">
      <div className="absolute inset-x-0 top-0 h-1.5 bg-gradient-to-r from-brand-blue to-brand-purple" />
      <div className="mb-5 grid h-14 w-14 place-items-center rounded-[20px] bg-blue-50 text-brand-blue">
        <Icon className="h-7 w-7" strokeWidth={1.8} />
      </div>
      <h3 className="m-0 text-[22px] font-black tracking-[-0.02em] text-ink">{scenario.name}</h3>
      <p className="mt-1 text-xs font-black uppercase tracking-[0.08em] text-muted">{scenario.englishName}</p>
      <div className="my-4 flex flex-wrap gap-2">
        <Badge tone="blue">AI：{scenario.aiRole}</Badge>
        <Badge tone={scenario.difficulty === "easy" ? "green" : scenario.difficulty === "medium" ? "violet" : "amber"}>
          难度：{scenario.difficultyLabel}
        </Badge>
      </div>
      <ul className="mb-6 grid flex-1 content-start gap-2 text-sm leading-6 text-muted">
        {scenario.goals.map((goal) => (
          <li key={goal} className="flex items-start gap-2">
            <span className="mt-0.5 grid h-5 w-5 shrink-0 place-items-center rounded-full bg-emerald-50 text-xs font-black text-emerald-600">
              ✓
            </span>
            {goal}
          </li>
        ))}
      </ul>
      {onStart ? (
        <Button className="mt-auto w-full justify-between disabled:cursor-not-allowed disabled:opacity-70" disabled={isStarting} onClick={() => onStart(scenario)}>
          {startContent}
        </Button>
      ) : (
        <ButtonLink to={`/training/${scenario.sessionId}`} className="mt-auto w-full justify-between">
          {startContent}
        </ButtonLink>
      )}
    </article>
  );
}
