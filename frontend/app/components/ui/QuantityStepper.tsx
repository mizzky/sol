import Button from "./Button";

type QuantityStepperProps = {
  value: number;
  min?: number;
  onChange: (next: number) => void;
  disabled?: boolean;
  label?: string;
};

export default function QuantityStepper({
  value,
  min = 1,
  onChange,
  disabled = false,
  label = "数量",
}: QuantityStepperProps) {
  const decreaseDisabled = disabled || value <= min;

  return (
    <div className="inline-flex items-center gap-3 rounded-2xl border border-zinc-200 bg-white px-3 py-2 shadow-sm">
      <div className="min-w-14 text-center">
        <div className="text-[11px] uppercase tracking-[0.24em] text-zinc-500">
          {label}
        </div>
        <div className="text-lg font-semibold text-zinc-900">{value}</div>
      </div>
      <div className="grid gap-2">
        <Button
          className="h-9 w-9 rounded-lg"
          disabled={disabled}
          onClick={() => onChange(value + 1)}
          size="icon"
          variant="secondary"
        >
          +
        </Button>
        <Button
          className="h-9 w-9 rounded-lg"
          disabled={decreaseDisabled}
          onClick={() => onChange(Math.max(min, value - 1))}
          size="icon"
          variant="secondary"
        >
          -
        </Button>
      </div>
    </div>
  );
}
