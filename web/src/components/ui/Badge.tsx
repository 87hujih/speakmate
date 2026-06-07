import type { ReactNode } from "react";
import { cn } from "../../utils/cn";

interface BadgeProps {
  children: ReactNode;
  tone?: "blue" | "green" | "amber" | "red" | "violet" | "slate";
  className?: string;
}

const toneClasses = {
  blue: "bg-blue-50 text-blue-700 border-blue-100",
  green: "bg-emerald-50 text-emerald-700 border-emerald-100",
  amber: "bg-amber-50 text-amber-700 border-amber-100",
  red: "bg-rose-50 text-rose-600 border-rose-100",
  violet: "bg-violet-50 text-violet-700 border-violet-100",
  slate: "bg-slate-100 text-slate-600 border-slate-200",
};

export function Badge({ children, tone = "slate", className }: BadgeProps) {
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1.5 rounded-full border px-3 py-1.5 text-xs font-black",
        toneClasses[tone],
        className,
      )}
    >
      {children}
    </span>
  );
}
