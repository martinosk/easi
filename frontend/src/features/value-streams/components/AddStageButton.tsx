interface AddStageButtonProps {
  onClick: () => void;
}

export function AddStageButton({ onClick }: AddStageButtonProps) {
  return (
    <button
      type="button"
      className="stage-add-btn"
      onClick={onClick}
      data-testid="add-stage-btn"
    >
      <svg viewBox="0 0 24 24" fill="none" width="20" height="20">
        <path d="M12 5V19M5 12H19" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
      </svg>
      <span>Add Stage</span>
    </button>
  );
}
