import { SegmentedControl } from '@mantine/core';
import type { BoardViewMode } from '../hooks/useMapViewState';

export interface ViewModeToggleProps {
  value: BoardViewMode;
  onChange: (mode: BoardViewMode) => void;
}

export function ViewModeToggle({ value, onChange }: ViewModeToggleProps) {
  return (
    <SegmentedControl
      value={value}
      onChange={(next) => onChange(next as BoardViewMode)}
      data={[
        { value: 'board', label: 'Board' },
        { value: 'map', label: 'Map' },
      ]}
      size="sm"
      data-testid="view-mode-toggle"
    />
  );
}
