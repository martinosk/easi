import { Tooltip } from '@mantine/core';
import type { ReactNode } from 'react';
import classes from './HelpTooltip.module.css';

export interface HelpTooltipProps {
  content: ReactNode;
  label?: string;
  iconOnly?: boolean;
  position?: 'top' | 'right' | 'bottom' | 'left';
}

export function HelpTooltip({ content, label, iconOnly = false, position = 'top' }: HelpTooltipProps) {
  return (
    <span className={classes.wrapper}>
      {label && !iconOnly && <span className={classes.label}>{label}</span>}
      <Tooltip label={content} position={position} withArrow multiline classNames={{ tooltip: classes.tooltip }}>
        <span className={classes.icon} role="img" aria-label="Help">
          <svg
            width="14"
            height="14"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <circle cx="12" cy="12" r="10" />
            <path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3" />
            <line x1="12" y1="17" x2="12.01" y2="17" />
          </svg>
        </span>
      </Tooltip>
    </span>
  );
}
