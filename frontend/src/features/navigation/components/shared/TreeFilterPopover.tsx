import { ActionIcon, Button, Group, Indicator, Popover } from '@mantine/core';
import React from 'react';
import classes from './TreeFilterPopover.module.css';

interface TreeFilterPopoverProps {
  ariaLabel: string;
  activeCount: number;
  onClearAll?: () => void;
  children: React.ReactNode;
}

const FILTER_ICON = (
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
    <polygon points="22 3 2 3 10 12.46 10 19 14 21 14 12.46 22 3" />
  </svg>
);

export const TreeFilterPopover: React.FC<TreeFilterPopoverProps> = ({
  ariaLabel,
  activeCount,
  onClearAll,
  children,
}) => {
  const hasActiveFilters = activeCount > 0;

  return (
    <Popover withinPortal position="bottom-start" offset={4} shadow="md" classNames={{ dropdown: classes.dropdown }}>
      <Indicator
        label={hasActiveFilters ? activeCount : undefined}
        disabled={!hasActiveFilters}
        size={16}
        offset={2}
        color="blue"
      >
        <Popover.Target>
          <ActionIcon
            variant={hasActiveFilters ? 'light' : 'subtle'}
            color={hasActiveFilters ? 'blue' : 'gray'}
            size="sm"
            aria-label={ariaLabel}
          >
            {FILTER_ICON}
          </ActionIcon>
        </Popover.Target>
      </Indicator>

      <Popover.Dropdown p={0}>
        <Group justify="space-between" className={classes.header}>
          <span className={classes.title}>Filters</span>
          {hasActiveFilters && onClearAll && (
            <Button variant="subtle" size="compact-xs" onClick={onClearAll}>
              Clear all
            </Button>
          )}
        </Group>
        {children}
      </Popover.Dropdown>
    </Popover>
  );
};
