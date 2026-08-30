import { Tooltip } from '@mantine/core';
import classes from './OnePagerIncompleteIndicator.module.css';

interface OnePagerIncompleteIndicatorProps {
  id: string;
  complete?: boolean;
}

export function OnePagerIncompleteIndicator({ id, complete }: OnePagerIncompleteIndicatorProps) {
  if (complete !== false) {
    return null;
  }

  return (
    <Tooltip label="One-pager incomplete" withArrow>
      <span className={classes.dot} data-testid={`one-pager-incomplete-${id}`} />
    </Tooltip>
  );
}
