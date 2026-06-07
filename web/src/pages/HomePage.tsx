import { ArrowRight, Bot, CheckCircle2, FileText, LoaderCircle, Mic, Sparkles } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { createTrainingSession, loadScenarioChoices } from "../api/loaders";
import { PageContainer } from "../components/layout/PageContainer";
import { ScenarioCard } from "../components/scenario/ScenarioCard";
import { Button, buttonClasses } from "../components/ui/Button";
import { ProgressBar } from "../components/ui/ProgressBar";
import { SectionHeader } from "../components/ui/SectionHeader";
import type { Scenario } from "../types";

function HeroPreview() {
  return (
    <div className="rounded-hero border border-line/80 bg-white/80 p-4 shadow-panel backdrop-blur-xl md:p-5">
      <div className="overflow-hidden rounded-[26px] border border-line bg-white">
        <div className="flex h-11 items-center gap-2 border-b border-line bg-slate-50 px-4">
          <span className="h-2.5 w-2.5 rounded-full bg-rose-400" />
          <span className="h-2.5 w-2.5 rounded-full bg-amber-300" />
          <span className="h-2.5 w-2.5 rounded-full bg-emerald-400" />
          <span className="ml-auto text-xs font-black uppercase tracking-[0.1em] text-slate-400">Live Training</span>
        </div>
        <div className="p-5">
          <div className="mb-3 max-w-[86%] rounded-[18px] rounded-tl-lg bg-blue-50 px-4 py-3 text-sm leading-6 text-blue-900">
            Could you briefly introduce yourself?
          </div>
          <div className="mb-3 ml-auto max-w-[86%] rounded-[18px] rounded-tr-lg bg-violet-50 px-4 py-3 text-sm leading-6 text-purple-950">
            I am study computer science and I have did a robot project.
          </div>
          <div className="mb-5 max-w-[86%] rounded-[18px] rounded-tl-lg bg-blue-50 px-4 py-3 text-sm leading-6 text-blue-900">
            That sounds interesting. Could you tell me more about your role?
          </div>
          <div className="rounded-[22px] border border-line bg-gradient-to-br from-white to-slate-50 p-4 shadow-soft">
            <div className="mb-3 flex items-center gap-2">
              <Sparkles className="h-4 w-4 text-brand-purple" />
              <strong className="text-sm text-ink">AI Coach 实时反馈</strong>
            </div>
            <div className="mb-2 flex items-center justify-between text-xs font-extrabold text-muted">
              <span>语法准确度</span>
              <b>72</b>
            </div>
            <ProgressBar value={72} />
            <div className="mb-2 mt-3 flex items-center justify-between text-xs font-extrabold text-muted">
              <span>表达自然度</span>
              <b>80</b>
            </div>
            <ProgressBar value={80} />
            <p className="mt-4 text-[13px] leading-6 text-muted">
              am study -&gt; am studying
              <br />
              have did -&gt; have done
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}

export function HomePage() {
  const navigate = useNavigate();
  const [scenarios, setScenarios] = useState<Scenario[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState("");
  const [startingScenarioId, setStartingScenarioId] = useState<number | null>(null);
  const interviewScenario = useMemo(() => scenarios.find((scenario) => scenario.code === "interview") ?? scenarios[0], [scenarios]);

  useEffect(() => {
    let ignore = false;

    async function load() {
      setIsLoading(true);
      setError("");
      try {
        const result = await loadScenarioChoices();
        if (!ignore) {
          setScenarios(result);
        }
      } catch (loadError) {
        if (!ignore) {
          setError(loadError instanceof Error ? loadError.message : "场景加载失败");
        }
      } finally {
        if (!ignore) {
          setIsLoading(false);
        }
      }
    }

    void load();

    return () => {
      ignore = true;
    };
  }, []);

  async function handleStart(scenario: Scenario) {
    setStartingScenarioId(scenario.id);
    setError("");

    try {
      const session = await createTrainingSession(scenario.id);
      navigate(`/training/${session.session_id}`);
    } catch (startError) {
      setError(startError instanceof Error ? startError.message : "创建训练失败");
    } finally {
      setStartingScenarioId(null);
    }
  }

  return (
    <PageContainer className="pt-8 md:pt-10">
      <section className="grid grid-cols-1 items-center gap-6 xl:grid-cols-[1.08fr_0.92fr] xl:gap-8">
        <div className="relative min-h-[410px] overflow-hidden rounded-hero bg-gradient-to-br from-brand-blue to-brand-purple p-7 text-white shadow-glow md:p-12">
          <div className="absolute -right-28 -top-32 h-96 w-96 rounded-full bg-white/15" />
          <div className="absolute -bottom-28 -left-24 h-64 w-64 rounded-full bg-white/10" />
          <div className="relative">
            <span className="inline-flex items-center gap-2 rounded-full border border-white/25 bg-white/15 px-3 py-2 text-[13px] font-black">
              <Sparkles className="h-4 w-4" />
              SpeakMate 场景化英语训练
            </span>
            <h1 className="mb-4 mt-6 max-w-[640px] text-[40px] font-black leading-[1.05] tracking-[-0.04em] md:text-[54px]">
              在真实场景中练英语，让 AI 帮你说得更自然
            </h1>
            <p className="m-0 max-w-[570px] text-[16px] leading-8 text-white/80 md:text-[17px]">
              SpeakMate 支持面试、点餐、会议等场景化文本对话，并在不打断交流的前提下给出表达纠错、能力评分与课后复盘。
            </p>
            <div className="mt-8 flex flex-wrap gap-3.5">
              <Button
                variant="ghost"
                className="border-white bg-white text-brand-blue shadow-none hover:bg-white/95 disabled:cursor-not-allowed disabled:opacity-70"
                disabled={!interviewScenario || startingScenarioId !== null}
                onClick={() => interviewScenario && handleStart(interviewScenario)}
              >
                {startingScenarioId === interviewScenario?.id ? <LoaderCircle className="h-4 w-4 animate-spin" /> : <Mic className="h-4 w-4" />}
                开始英语面试训练
              </Button>
              <Link to="/history" className={buttonClasses("secondary")}>
                <FileText className="h-4 w-4" />
                查看历史记录
              </Link>
            </div>
          </div>
        </div>
        <HeroPreview />
      </section>

      <SectionHeader title="选择训练场景" description="每个场景都有 AI 角色、任务目标和评分重点。" />
      {error ? (
        <div className="mb-5 rounded-[22px] border border-rose-100 bg-rose-50 p-4 text-sm font-bold text-rose-700">
          {error}
        </div>
      ) : null}
      {isLoading ? (
        <section className="grid min-h-[220px] place-items-center rounded-panel border border-line bg-white/85 shadow-soft">
          <div className="inline-flex items-center gap-2 text-sm font-black text-muted">
            <LoaderCircle className="h-5 w-5 animate-spin text-brand-blue" />
            正在加载训练场景
          </div>
        </section>
      ) : scenarios.length === 0 ? (
        <section className="rounded-panel border border-line bg-white/85 p-8 text-center shadow-soft">
          <h3 className="m-0 text-xl font-black text-ink">暂无可训练场景</h3>
          <p className="mt-2 text-sm font-semibold text-muted">当前没有可用训练场景，请稍后重试。</p>
        </section>
      ) : (
        <section className="grid grid-cols-1 gap-5 md:grid-cols-2 xl:grid-cols-3">
          {scenarios.map((scenario) => (
            <ScenarioCard
              key={scenario.id}
              scenario={scenario}
              isStarting={startingScenarioId === scenario.id}
              onStart={handleStart}
            />
          ))}
        </section>
      )}

      <section className="mt-10 grid grid-cols-1 gap-5 md:grid-cols-3">
        {[
          { icon: Bot, title: "AI 主动追问", text: "根据当前阶段继续追问，让练习保持真实任务感。" },
          { icon: CheckCircle2, title: "低打断纠错", text: "对话中给轻量提示，结束后集中生成结构化报告。" },
          { icon: ArrowRight, title: "训练记录闭环", text: "练习过程、即时反馈和课后报告保持同步，方便持续复盘表现变化。" },
        ].map((item) => {
          const Icon = item.icon;
          return (
            <div key={item.title} className="rounded-panel border border-line bg-white/80 p-5 shadow-soft">
              <Icon className="mb-4 h-6 w-6 text-brand-blue" />
              <h3 className="m-0 text-base font-black text-ink">{item.title}</h3>
              <p className="mt-2 text-sm leading-6 text-muted">{item.text}</p>
            </div>
          );
        })}
      </section>
    </PageContainer>
  );
}
