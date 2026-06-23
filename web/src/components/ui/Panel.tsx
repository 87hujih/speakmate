import type { ReactNode } from "react";
import { cn } from "../../utils/cn";

/** PanelProps 定义对应组件接收的属性。 */
interface PanelProps {
  children: ReactNode;
  className?: string;
}

/** Panel 渲染对应的页面或界面组件。 */
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
