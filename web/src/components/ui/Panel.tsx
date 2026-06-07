import type { ReactNode } from "react";
import { cn } from "../../utils/cn";

interface PanelProps {
  children: ReactNode;
  className?: string;
}

export function Panel({ children, className }: PanelProps) {
  return (
    <section
      className={cn(
        "overflow-hidden rounded-panel border border-line bg-white/90 shadow-soft backdrop-blur-xl",
        className,
      )}
    >
      {children}
    </section>
  );
}
