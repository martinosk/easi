import { SegmentedControl } from '@mantine/core';
import { BOARD_LENSES, type BoardLens, LENS_LABELS } from '../lens/boardLens';

export interface LensSwitcherProps {
  lens: BoardLens;
  onLensChange: (lens: BoardLens) => void;
}

export function LensSwitcher({ lens, onLensChange }: LensSwitcherProps) {
  return (
    <SegmentedControl
      value={lens}
      onChange={(value) => onLensChange(value as BoardLens)}
      data={BOARD_LENSES.map((value) => ({ value, label: LENS_LABELS[value] }))}
      size="sm"
      data-testid="lens-switcher"
    />
  );
}
