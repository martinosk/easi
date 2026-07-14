import { Badge } from '@mantine/core';
import type { OnePagerQualityCompleteness } from '../types';

interface CompletenessBadgeProps {
  completeness: OnePagerQualityCompleteness;
}

const COMPLETENESS_LABELS: Record<OnePagerQualityCompleteness, string> = {
  complete: 'Complete',
  incomplete: 'Incomplete',
  'not-applicable': 'Not Applicable',
};

const COMPLETENESS_COLORS: Record<OnePagerQualityCompleteness, string> = {
  complete: 'teal',
  incomplete: 'red',
  'not-applicable': 'gray',
};

export function CompletenessBadge({ completeness }: CompletenessBadgeProps) {
  return (
    <Badge color={COMPLETENESS_COLORS[completeness]} variant="light" radius="sm" data-testid="completeness-badge">
      {COMPLETENESS_LABELS[completeness]}
    </Badge>
  );
}
