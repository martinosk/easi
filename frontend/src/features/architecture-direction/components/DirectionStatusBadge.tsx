import { Badge } from '@mantine/core';
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
  draft: 'gray',
  proposed: 'blue',
  agreed: 'green',
  rejected: 'red',
};

export function DirectionStatusBadge({ status }: DirectionStatusBadgeProps) {
  return (
    <Badge color={STATUS_COLORS[status]} variant="filled" data-testid="direction-status-badge">
      {STATUS_LABELS[status]}
    </Badge>
  );
}
