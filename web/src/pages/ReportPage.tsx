import { LoaderCircle } from "lucide-react";
import { Link, useParams } from "react-router-dom";
import { useEffect, useState } from "react";
import { generateReportState, loadReportState } from "../api/loaders";
import { ErrorAnalysisCard } from "../components/report/ErrorAnalysisCard";
import { PracticePlanCard } from "../components/report/PracticePlanCard";
import { ReportSummaryCard } from "../components/report/ReportSummaryCard";
import { ScoreOverview } from "../components/report/ScoreOverview";
import { PageContainer } from "../components/layout/PageContainer";
import { Button, buttonClasses } from "../components/ui/Button";
import { Panel } from "../components/ui/Panel";
import type { TrainingReport } from "../types";

function parseRouteSessionId(value: string | undefined) {
  const numeric = Number(value);

  return Number.isInteger(numeric) && numeric > 0 ? numeric : null;
}

export function ReportPage() {
  const { sessionId } = useParams();
  const numericSessionId = parseRouteSessionId(sessionId);
  const [report, setReport] = useState<TrainingReport | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isGenerating, setIsGenerating] = useState(false);
  const [isMissing, setIsMissing] = useState(false);
  const [error, setError] = useState("");

  async function load() {
    if (!numericSessionId) {
      setError("训练 ID 不合法");
      setIsLoading(false);
      return;
    }

    setIsLoading(true);
    setError("");
    setIsMissing(false);
    try {
      const result = await loadReportState(numericSessionId);
      if (result.status === "ready") {
        setReport(result.report);
      } else {
        setReport(null);
        setIsMissing(true);
      }
    } catch (loadError) {
      setError(loadError instanceof Error ? loadError.message : "报告加载失败");
    } finally {
      setIsLoading(false);
    }
  }

  useEffect(() => {
    void load();
  }, [numericSessionId]);

  async function handleGenerate() {
    if (!numericSessionId || isGenerating) {
      return;
    }

    setIsGenerating(true);
    setError("");
    try {
      const generated = await generateReportState(numericSessionId);
      setReport(generated);
      setIsMissing(false);
    } catch (generateError) {
      setError(generateError instanceof Error ? generateError.message : "报告生成失败");
    } finally {
      setIsGenerating(false);
    }
  }

  if (isLoading) {
    return (
      <PageContainer size="wide">
        <section className="grid min-h-[360px] place-items-center rounded-panel border border-line bg-white p-8 text-center shadow-panel">
          <div>
            <LoaderCircle className="mx-auto h-7 w-7 animate-spin text-brand-blue" />
            <h2 className="m-0 mt-4 text-xl font-black text-ink">正在查询课后报告</h2>
            <p className="mt-2 text-sm font-semibold text-muted">如果报告尚未生成，页面会提供生成入口。</p>
          </div>
        </section>
      </PageContainer>
    );
  }

  if (isMissing || (!report && !error)) {
    return (
      <PageContainer size="wide">
        <Panel className="p-8 text-center">
          <h1 className="m-0 text-2xl font-black text-ink">报告尚未生成</h1>
          <p className="mx-auto mt-3 max-w-[560px] text-sm font-semibold leading-6 text-muted">
            生成报告前需要先结束训练，并至少完成一轮文本对话和反馈生成。
          </p>
          {error ? <p className="mt-4 rounded-2xl bg-rose-50 px-4 py-3 text-sm font-bold text-rose-700">{error}</p> : null}
          <div className="mt-6 flex flex-wrap justify-center gap-3">
            <Button disabled={isGenerating} onClick={handleGenerate} className="disabled:cursor-not-allowed disabled:opacity-70">
              {isGenerating ? <LoaderCircle className="h-4 w-4 animate-spin" /> : null}
              生成课后报告
            </Button>
            <Link to={`/training/${numericSessionId ?? ""}`} className={buttonClasses("ghost")}>
              返回训练详情
            </Link>
          </div>
        </Panel>
      </PageContainer>
    );
  }

  if (error || !report) {
    return (
      <PageContainer size="wide">
        <Panel className="p-8 text-center">
          <h1 className="m-0 text-2xl font-black text-rose-700">报告加载失败</h1>
          <p className="mt-3 text-sm font-bold text-rose-600">{error || "报告不存在或服务已重启。"}</p>
          <div className="mt-6 flex flex-wrap justify-center gap-3">
            <button type="button" className={buttonClasses("danger")} onClick={() => load()}>
              重试
            </button>
            <Link to="/history" className={buttonClasses("ghost")}>
              查看历史记录
            </Link>
          </div>
        </Panel>
      </PageContainer>
    );
  }

  return (
    <PageContainer size="wide">
      <ReportSummaryCard report={report} />
      <div className="grid grid-cols-1 items-start gap-5 xl:grid-cols-[430px_1fr]">
        <ScoreOverview scores={report.scores} />

        <section className="grid gap-4">
          <Panel className="p-6">
            <h3 className="m-0 mb-4 text-xl font-black tracking-[-0.02em] text-ink">高频错误总结</h3>
            {report.frequentErrors.length ? (
              <div className="grid gap-3">
                {report.frequentErrors.map((correction) => (
                  <ErrorAnalysisCard key={`${correction.title}-${correction.original}`} correction={correction} />
                ))}
              </div>
            ) : (
              <p className="m-0 rounded-2xl bg-slate-50 p-4 text-sm font-semibold text-muted">暂无高频错误。</p>
            )}
          </Panel>

          <Panel className="p-6">
            <h3 className="m-0 mb-4 text-xl font-black tracking-[-0.02em] text-ink">表达升级</h3>
            {report.betterExpressions.length ? (
              <div className="grid gap-3">
                {report.betterExpressions.map((expression) => (
                  <div key={expression.after} className="grid grid-cols-1 gap-3 rounded-[20px] border border-line bg-white p-4 md:grid-cols-2">
                    <div className="rounded-2xl bg-slate-50 p-3 text-[13px] leading-6 text-muted">
                      <strong className="mb-1 block text-ink">你的表达</strong>
                      {expression.before}
                    </div>
                    <div className="rounded-2xl bg-blue-50 p-3 text-[13px] leading-6 text-blue-800">
                      <strong className="mb-1 block text-blue-950">更自然表达</strong>
                      {expression.after}
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <p className="m-0 rounded-2xl bg-slate-50 p-4 text-sm font-semibold text-muted">暂无表达升级建议。</p>
            )}
          </Panel>

          <Panel className="p-6">
            <div className="mb-4 flex flex-wrap items-center justify-between gap-4">
              <h3 className="m-0 text-xl font-black tracking-[-0.02em] text-ink">下次练习计划</h3>
              <Link to={`/training/${report.sessionId}`} className={buttonClasses("soft", "h-10 rounded-2xl px-4")}>
                回到训练详情
              </Link>
            </div>
            {report.nextPracticePlan.length ? (
              <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
                {report.nextPracticePlan.map((item, index) => (
                  <PracticePlanCard key={item.title} item={item} index={index} />
                ))}
              </div>
            ) : (
              <p className="m-0 rounded-2xl bg-slate-50 p-4 text-sm font-semibold text-muted">暂无练习计划。</p>
            )}
          </Panel>
        </section>
      </div>
    </PageContainer>
  );
}
