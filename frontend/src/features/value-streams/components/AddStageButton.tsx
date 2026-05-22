import { Button } from '@mantine/core';

interface AddStageButtonProps {
  onClick: () => void;
}

const PLUS_ICON = (
  <svg viewBox="0 0 24 24" fill="none" width="16" height="16" aria-hidden="true">
    <path d="M12 5V19M5 12H19" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
  </svg>
);

export function AddStageButton({ onClick }: AddStageButtonProps) {
  return (
    <Button
      variant="default"
      leftSection={PLUS_ICON}
      className="stage-add-btn"
      onClick={onClick}
      data-testid="add-stage-btn"
    >
      Add Stage
    </Button>
  );
}
