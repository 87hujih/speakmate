import type { ReactNode } from "react";

interface SectionHeaderProps {
  title: string;
  description?: string;
  action?: ReactNode;
}

export function SectionHeader({ title, description, action }: SectionHeaderProps) {
  return (
    <div className="mb-5 mt-12 flex items-end justify-between gap-6">
      <div>
        <h2 className="m-0 text-3xl font-black tracking-[-0.035em] text-ink">{title}</h2>
        {description ? <p className="mt-2 text-sm leading-6 text-muted">{description}</p> : null}
      </div>
      {action}
    </div>
  );
}
