import type {
  InputHTMLAttributes,
  ReactNode,
  SelectHTMLAttributes,
  TextareaHTMLAttributes,
} from "react";
import { cn } from "./cn";

const controlClassName =
  "w-full rounded-xl border border-zinc-200 bg-white px-4 py-3 text-zinc-900 shadow-sm outline-none placeholder:text-zinc-500 focus:border-indigo-300 focus:ring-4 focus:ring-indigo-100";

type FieldWrapperProps = {
  label?: string;
  hint?: string;
  htmlFor?: string;
  className?: string;
  children: ReactNode;
};

export function FieldWrapper({
  label,
  hint,
  htmlFor,
  className,
  children,
}: FieldWrapperProps) {
  return (
    <label className={cn("grid gap-2 text-sm", className)} htmlFor={htmlFor}>
      {label && <span className="font-medium text-zinc-800">{label}</span>}
      {children}
      {hint && <span className="text-xs text-zinc-500">{hint}</span>}
    </label>
  );
}

export function Input({
  className,
  ...props
}: InputHTMLAttributes<HTMLInputElement>) {
  return <input className={cn(controlClassName, className)} {...props} />;
}

export function Select({
  className,
  ...props
}: SelectHTMLAttributes<HTMLSelectElement>) {
  return <select className={cn(controlClassName, className)} {...props} />;
}

export function Textarea({
  className,
  ...props
}: TextareaHTMLAttributes<HTMLTextAreaElement>) {
  return (
    <textarea
      className={cn(controlClassName, "min-h-28 resize-y", className)}
      {...props}
    />
  );
}

type CheckboxFieldProps = InputHTMLAttributes<HTMLInputElement> & {
  label: ReactNode;
};

export function CheckboxField({
  className,
  label,
  ...props
}: CheckboxFieldProps) {
  return (
    <label
      className={cn(
        "inline-flex items-center gap-3 rounded-xl border border-zinc-200 bg-zinc-50 px-4 py-3 text-sm text-zinc-700 shadow-sm",
        className,
      )}
    >
      <input
        type="checkbox"
        className="h-4 w-4 rounded border-zinc-300 text-indigo-600 focus:ring-indigo-200"
        {...props}
      />
      <span>{label}</span>
    </label>
  );
}

export function FieldMessage({
  tone = "default",
  children,
}: {
  tone?: "default" | "error" | "success" | "warning";
  children: ReactNode;
}) {
  const toneClassName =
    tone === "error"
      ? "border-rose-200 bg-rose-50 text-rose-700"
      : tone === "success"
        ? "border-emerald-200 bg-emerald-50 text-emerald-700"
        : tone === "warning"
          ? "border-amber-200 bg-amber-50 text-amber-800"
          : "border-zinc-200 bg-white text-zinc-700";

  return (
    <div
      className={cn(
        "rounded-xl border px-4 py-3 text-sm shadow-sm",
        toneClassName,
      )}
    >
      {children}
    </div>
  );
}
