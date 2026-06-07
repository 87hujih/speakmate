import type { ButtonHTMLAttributes, ReactNode } from "react";
import { Link, type LinkProps } from "react-router-dom";
import { cn } from "../../utils/cn";

type ButtonVariant = "primary" | "secondary" | "ghost" | "danger" | "soft";

const variantClasses: Record<ButtonVariant, string> = {
  primary:
    "bg-gradient-to-br from-brand-blue to-brand-purple text-white shadow-[0_14px_28px_rgba(37,99,235,0.23)] hover:-translate-y-0.5",
  secondary:
    "border border-white/25 bg-white/15 text-white hover:-translate-y-0.5",
  ghost: "border border-line bg-white text-ink hover:-translate-y-0.5 hover:shadow-soft",
  danger: "border border-rose-100 bg-rose-50 text-rose-600 hover:-translate-y-0.5",
  soft: "border border-blue-100 bg-blue-50 text-blue-700 hover:-translate-y-0.5",
};

export function buttonClasses(variant: ButtonVariant = "primary", className?: string) {
  return cn(
    "inline-flex h-12 items-center justify-center gap-2 rounded-2xl px-5 text-sm font-extrabold transition duration-200 active:translate-y-px",
    variantClasses[variant],
    className,
  );
}

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant;
}

export function Button({ className, variant = "primary", type = "button", ...props }: ButtonProps) {
  return <button className={buttonClasses(variant, className)} type={type} {...props} />;
}

interface ButtonLinkProps extends LinkProps {
  variant?: ButtonVariant;
  children: ReactNode;
}

export function ButtonLink({ className, variant = "primary", children, ...props }: ButtonLinkProps) {
  return (
    <Link className={buttonClasses(variant, className)} {...props}>
      {children}
    </Link>
  );
}
