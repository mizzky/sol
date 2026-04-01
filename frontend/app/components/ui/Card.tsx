import type { HTMLAttributes, ReactNode } from "react";
import { cn } from "./cn";

type CardProps = HTMLAttributes<HTMLDivElement> & {
  children: ReactNode;
};

export default function Card({ className, children, ...props }: CardProps) {
  return (
    <div
      className={cn(
        "rounded-2xl border border-zinc-200 bg-white p-6 shadow-sm transition-shadow hover:shadow-md",
        className,
      )}
      {...props}
    >
      {children}
    </div>
  );
}
