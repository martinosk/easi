import type { DirectionStatus } from '../types';

interface DirectionStatusBadgeProps {
  status: DirectionStatus;
}

const STATUS_LABELS: Record<DirectionStatus, string> = {
  draft: 'Draft',
  proposed: 'Proposed',
  agreed: 'Agreed',
  rejected: 'Rejected',
};

const STATUS_COLORS: Record<DirectionStatus, string> = {
  draft: '#6B7280',
  proposed: '#2563EB',
  agreed: '#15803D',
  rejected: '#9CA3AF',
};

export function DirectionStatusBadge({ status }: DirectionStatusBadgeProps) {
  return (
    <span
      data-testid="direction-status-badge"
      style={{
        backgroundColor: STATUS_COLORS[status],
        color: 'white',
        padding: '2px 10px',
        borderRadius: 12,
        fontSize: 12,
        fontWeight: 600,
        textTransform: 'uppercase',
        letterSpacing: '0.04em',
      }}
    >
      {STATUS_LABELS[status]}
    </span>
  );
}
