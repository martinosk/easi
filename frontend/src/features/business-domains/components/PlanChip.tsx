import classes from './PlanChip.module.css';

export type PlanChipVariant = 'plain' | 'legacy' | 'future' | 'standard';

export interface PlanChipProps {
  label: string;
  variant?: PlanChipVariant;
}

export function PlanChip({ label, variant = 'plain' }: PlanChipProps) {
  return (
    <span
      className={[classes.chip, classes[variant]].join(' ')}
      title={label}
      data-testid="plan-chip"
      data-variant={variant}
    >
      {label}
    </span>
  );
}
