import { LoaderCircle, Plus, TrendingDown, TrendingUp } from "lucide-react";
import { Link } from "react-router-dom";
import { useEffect, useMemo, useState } from "react";
import { loadHistoryState } from "../api/loaders";
import { HistorySessionCard } from "../components/history/HistorySessionCard";
import { PageContainer } from "../components/layout/PageContainer";
import { buttonClasses } from "../components/ui/Button";
import { ProgressBar } from "../components/ui/ProgressBar";
import { SectionHeader } from "../components/ui/SectionHeader";
import type { HistoryRecord } from "../types";

const pageSize = 10;

export function HistoryPage() {
  const [records, setRecords] = useState<HistoryRecord[]>([]);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState("");
  const totalPages = Math.max(1, Math.ceil(total / pageSize));
  const averageScore = useMemo(() => {
    const scored = records.filter((record) => record.score > 0);
    if (scored.length === 0) {
      return 0;
    }

    return Math.round(scored.reduce((sum, record) => sum + record.score, 0) / scored.length);
  }, [records]);

  async function load(targetPage = page) {
    setIsLoading(true);
    setError("");
    try {
      const result = await loadHistoryState(targetPage, pageSize);
      setRecords(result.records);
      setPage(result.page);
      setTotal(result.total);
    } catch (loadError) {
      setError(loadError instanceof Error ? loadError.message : "历史记录加载失败");
    } finally {
      setIsLoading(false);
    }
  }

  useEffect(() => {
    void load(page);
  }, [page]);

  return (
    <PageContainer>
      <SectionHeader
        title="历史训练记录"
        description="持续记录每次训练结果，观察口语能力变化趋势。"
        action={
          <Link to="/" className={buttonClasses("primary")}>
            <Plus className="h-4 w-4" />
            开始新训练
          </Link>
        }
      />

      <div className="grid grid-cols-1 gap-5 xl:grid-cols-[330px_1fr]">
        <aside className="rounded-[30px] border border-line bg-white p-6 shadow-soft">
          <span className="inline-flex items-center gap-2 rounded-full border border-blue-100 bg-blue-50 px-3 py-2 text-[13px] font-black text-blue-700">
            <TrendingUp className="h-4 w-4" />
            Progress
          </span>
          <h2 className="mb-1 mt-5 text-2xl font-black tracking-[-0.03em] text-ink">最近训练</h2>
          <p className="m-0 mb-5 text-sm leading-6 text-muted">
            当前页 {records.length} 条记录，总计 {total} 条训练记录。
          </p>
          <div className="text-[54px] font-black leading-none tracking-[-0.055em] text-ink">{averageScore || "--"}</div>
          <p className="mb-6 mt-2 text-sm font-bold text-muted">当前页平均综合评分</p>
          <div className="mb-5">
            <div className="mb-1.5 flex items-center justify-between text-xs font-extrabold text-muted">
              <span className="inline-flex items-center gap-1.5">
                <TrendingDown className="h-3.5 w-3.5 text-emerald-500" />
                已完成训练占比
              </span>
              <b>{records.length ? Math.round((records.filter((record) => record.status === "finished").length / records.length) * 100) : 0}%</b>
            </div>
            <ProgressBar value={records.length ? Math.round((records.filter((record) => record.status === "finished").length / records.length) * 100) : 0} />
          </div>
          <div>
            <div className="mb-1.5 flex items-center justify-between text-xs font-extrabold text-muted">
              <span>报告生成占比</span>
              <b>{records.length ? Math.round((records.filter((record) => record.reportStatus === "generated").length / records.length) * 100) : 0}%</b>
            </div>
            <ProgressBar value={records.length ? Math.round((records.filter((record) => record.reportStatus === "generated").length / records.length) * 100) : 0} />
          </div>
        </aside>

        <section className="grid content-start gap-3.5">
          {error ? (
            <div className="rounded-panel border border-rose-100 bg-rose-50 p-6 text-center shadow-soft">
              <h3 className="m-0 text-xl font-black text-rose-700">历史记录加载失败</h3>
              <p className="mt-2 text-sm font-bold text-rose-600">{error}</p>
              <button type="button" className={buttonClasses("danger", "mt-5")} onClick={() => load()}>
                重试
              </button>
            </div>
          ) : isLoading ? (
            <div className="grid min-h-[260px] place-items-center rounded-panel border border-line bg-white p-8 shadow-soft">
              <div className="inline-flex items-center gap-2 text-sm font-black text-muted">
                <LoaderCircle className="h-5 w-5 animate-spin text-brand-blue" />
                正在加载历史记录
              </div>
            </div>
          ) : records.length === 0 ? (
            <div className="rounded-panel border border-line bg-white p-8 text-center shadow-soft">
              <h3 className="m-0 text-xl font-black text-ink">暂无历史训练</h3>
              <p className="mt-2 text-sm font-semibold text-muted">完成一次场景训练后，这里会显示记录和报告状态。</p>
              <Link to="/" className={buttonClasses("primary", "mt-5")}>
                选择训练场景
              </Link>
            </div>
          ) : (
            <>
              {records.map((record) => (
                <HistorySessionCard key={record.sessionId} record={record} />
              ))}
              <div className="mt-2 flex items-center justify-between rounded-panel border border-line bg-white p-4 shadow-soft">
                <button
                  type="button"
                  className={buttonClasses("ghost", "h-10 rounded-2xl px-4 disabled:cursor-not-allowed disabled:opacity-60")}
                  disabled={page <= 1 || isLoading}
                  onClick={() => setPage((current) => Math.max(1, current - 1))}
                >
                  上一页
                </button>
                <span className="text-sm font-black text-muted">
                  第 {page} / {totalPages} 页
                </span>
                <button
                  type="button"
                  className={buttonClasses("ghost", "h-10 rounded-2xl px-4 disabled:cursor-not-allowed disabled:opacity-60")}
                  disabled={page >= totalPages || isLoading}
                  onClick={() => setPage((current) => Math.min(totalPages, current + 1))}
                >
                  下一页
                </button>
              </div>
            </>
          )}
        </section>
      </div>
    </PageContainer>
  );
}
