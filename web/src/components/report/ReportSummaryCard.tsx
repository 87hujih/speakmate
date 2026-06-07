import { CalendarClock, FileText, MessageSquareText, Target, TriangleAlert, type LucideIcon } from "lucide-react";
import type { TrainingReport } from "../../types";

interface ReportSummaryCardProps {
  report: TrainingReport;
}

function StatCard({ label, value, icon: Icon }: { label: string; value: string; icon: LucideIcon }) {
  return (
    <div className="rounded-3xl border border-line bg-white p-5 shadow-soft">
      <div className="flex items-center justify-between">
        <span className="text-sm font-extrabold text-muted">{label}</span>
        <Icon className="h-5 w-5 text-brand-blue" />
      </div>
      <strong className="mt-2 block text-[28px] font-black tracking-[-0.03em] text-ink">{value}</strong>
    </div>
  );
}

export function ReportSummaryCard({ report }: ReportSummaryCardProps) {
  return (
    <>
      <section className="flex items-center justify-between gap-5 rounded-[32px] bg-gradient-to-br from-slate-950 via-blue-900 to-violet-800 p-8 text-white shadow-[0_30px_70px_rgba(15,23,42,0.24)]">
        <div>
          <span className="inline-flex items-center gap-2 rounded-full border border-white/25 bg-white/15 px-3 py-2 text-[13px] font-black">
            <FileText className="h-4 w-4" />
            Post-class Report
          </span>
          <h1 className="m-0 mt-4 text-[34px] font-black tracking-[-0.035em]">{report.scenario.name}训练报告</h1>
          <p className="mt-2 text-sm leading-6 text-white/70">{report.summary}</p>
        </div>
        <div className="min-w-[150px] rounded-[26px] border border-white/25 bg-white/15 px-7 py-5 text-center">
          <strong className="block text-5xl font-black leading-none tracking-[-0.04em]">{report.totalScore}</strong>
          <span className="mt-2 block text-sm font-bold text-white/75">{report.grade}</span>
        </div>
      </section>

      <div className="my-5 grid grid-cols-4 gap-4">
        <StatCard label="训练时长" value={report.durationLabel} icon={CalendarClock} />
        <StatCard label="对话轮次" value={`${report.turnCount}`} icon={MessageSquareText} />
        <StatCard label="发现问题" value={`${report.issueCount}`} icon={TriangleAlert} />
        <StatCard label="场景完成度" value={`${report.completionRate}%`} icon={Target} />
      </div>
    </>
  );
}
