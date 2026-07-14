import { SegmentedControl, Text } from '@mantine/core';
import { BOARD_LENSES, type BoardLens, LENS_DESCRIPTIONS, LENS_LABELS } from '../lens/boardLens';
import classes from './LensSwitcher.module.css';

export interface LensSwitcherProps {
  lens: BoardLens;
  onLensChange: (lens: BoardLens) => void;
}

export function LensSwitcher({ lens, onLensChange }: LensSwitcherProps) {
  return (
    <div className={classes.wrapper}>
      <SegmentedControl
        value={lens}
        onChange={(value) => onLensChange(value as BoardLens)}
        data={BOARD_LENSES.map((value) => ({ value, label: LENS_LABELS[value] }))}
        size="sm"
        data-testid="lens-switcher"
      />
      <Text className={classes.description} data-testid="lens-description">
        {LENS_DESCRIPTIONS[lens]}
      </Text>
    </div>
  );
}
