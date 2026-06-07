import type { ReactNode } from "react";
import { cn } from "../../utils/cn";

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

export function PageContainer({ children, className, size = "default" }: PageContainerProps) {
  return <div className={cn(sizeClasses[size], className)}>{children}</div>;
}
