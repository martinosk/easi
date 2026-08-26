import { UnstyledButton } from '@mantine/core';
import React from 'react';
import classes from './TreeFilterSection.module.css';

interface TreeFilterSectionProps {
  label: string;
  hasSelection: boolean;
  onClear: () => void;
  children: React.ReactNode;
}

export const TreeFilterSection: React.FC<TreeFilterSectionProps> = ({ label, hasSelection, onClear, children }) => (
  <div className={classes.section}>
    <div className={classes.header}>
      <span className={classes.label}>{label}</span>
      {hasSelection && (
        <UnstyledButton
          component="button"
          type="button"
          className={classes.clear}
          onClick={onClear}
          aria-label="Clear filter"
        >
          Clear
        </UnstyledButton>
      )}
    </div>
    {children}
  </div>
);
