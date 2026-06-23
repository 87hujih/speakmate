import type { ReactNode } from "react";
import { cn } from "../../utils/cn";

/** PageContainerProps 定义对应组件接收的属性。 */
interface PageContainerProps {
  children: ReactNode;
  className?: string;
  size?: "default" | "wide" | "full";
}

const sizeClasses = {
  default: "mx-auto max-w-[1200px] px-4 pb-20 pt-8 md:px-7",
  wide: "mx-auto max-w-[1180px] px-4 pb-20 pt-8 md:px-7",
  full: "p-3 md:p-[18px]",
};

/** PageContainer 渲染对应的页面或界面组件。 */
export function PageContainer({ children, className, size = "default" }: PageContainerProps) {
  return <div className={cn(sizeClasses[size], className)}>{children}</div>;
}
