import { ActionIcon, UnstyledButton } from '@mantine/core';
import React from 'react';
import classes from './TreeSection.module.css';

interface TreeSectionProps {
  label: string;
  count: number;
  isExpanded: boolean;
  onToggle: () => void;
  onAdd?: () => void;
  addTitle?: string;
  addTestId?: string;
  children: React.ReactNode;
}

export const TreeSection: React.FC<TreeSectionProps> = ({
  label,
  count,
  isExpanded,
  onToggle,
  onAdd,
  addTitle,
  addTestId,
  children,
}) => {
  return (
    <div className={classes.section}>
      <div className={classes.headerRow}>
        <UnstyledButton component="button" type="button" className={classes.header} onClick={onToggle}>
          <span className={classes.icon}>{isExpanded ? '▼' : '▶'}</span>
          <span className={classes.label}>{label}</span>
          <span className={classes.count}>{count}</span>
        </UnstyledButton>
        {onAdd && (
          <ActionIcon
            variant="filled"
            size="sm"
            className={classes.addButton}
            onClick={onAdd}
            title={addTitle}
            data-testid={addTestId}
            aria-label={addTitle ?? 'Add'}
          >
            +
          </ActionIcon>
        )}
      </div>
      {isExpanded && children}
    </div>
  );
};
