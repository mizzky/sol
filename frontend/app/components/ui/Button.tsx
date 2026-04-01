import type { ButtonHTMLAttributes, ReactNode } from "react";
import { cn } from "./cn";

type ButtonVariant = "primary" | "secondary" | "outline" | "danger" | "ghost";
type ButtonSize = "sm" | "md" | "icon";

type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: ButtonVariant;
  size?: ButtonSize;
  leftIcon?: ReactNode;
};

const variantClasses: Record<ButtonVariant, string> = {
  primary:
    "bg-indigo-600 text-white shadow-sm hover:bg-indigo-500 hover:shadow-md focus-visible:ring-indigo-300",
  secondary:
    "bg-white text-zinc-900 ring-1 ring-zinc-200 shadow-sm hover:bg-zinc-100 hover:shadow-md focus-visible:ring-zinc-300",
  outline:
    "bg-white text-indigo-700 ring-1 ring-indigo-200 shadow-sm hover:bg-indigo-50 hover:shadow-md focus-visible:ring-indigo-300",
  danger:
    "bg-white text-rose-700 ring-1 ring-rose-200 shadow-sm hover:bg-rose-50 hover:shadow-md focus-visible:ring-rose-300",
  ghost:
    "bg-transparent text-zinc-700 hover:bg-zinc-100 focus-visible:ring-zinc-300",
};

const sizeClasses: Record<ButtonSize, string> = {
  sm: "min-h-9 px-3 py-2 text-sm",
  md: "min-h-11 px-4 py-2.5 text-sm font-medium",
  icon: "h-10 w-10 items-center justify-center p-0",
};

export default function Button({
  className,
  variant = "primary",
  size = "md",
  leftIcon,
  type = "button",
  children,
  ...props
}: ButtonProps) {
  return (
    <button
      type={type}
      className={cn(
        "inline-flex items-center justify-center gap-2 rounded-xl border border-transparent transition focus-visible:outline-none focus-visible:ring-4 disabled:opacity-50 disabled:hover:shadow-sm",
        variantClasses[variant],
        sizeClasses[size],
        className,
      )}
      {...props}
    >
      {leftIcon}
      {children}
    </button>
  );
}
