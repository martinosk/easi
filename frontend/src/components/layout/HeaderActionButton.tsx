import { Tooltip, UnstyledButton } from '@mantine/core';
import classes from './AppNavigation.module.css';

interface HeaderActionButtonProps {
  icon: React.ReactNode;
  label: string;
  onClick: () => void;
  testId: string;
  active?: boolean;
}

export function HeaderActionButton({ icon, label, onClick, testId, active }: HeaderActionButtonProps) {
  const className = active ? `${classes.actionButton} ${classes.actionButtonActive}` : classes.actionButton;
  return (
    <Tooltip label={label} openDelay={300} withinPortal>
      <UnstyledButton
        component="button"
        type="button"
        className={className}
        onClick={onClick}
        aria-label={label}
        aria-pressed={active === undefined ? undefined : active}
        data-testid={testId}
      >
        {icon}
        <span className={classes.actionLabel}>{label}</span>
      </UnstyledButton>
    </Tooltip>
  );
}
