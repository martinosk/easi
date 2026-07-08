import { ActionIcon, Highlight } from '@mantine/core';
import React from 'react';
import type { CapabilityTreeNode } from '../../hooks/useCapabilityTree';
import classes from './CapabilityTree.module.css';

const HIGHLIGHT_STYLES: React.CSSProperties = {
  fontWeight: 700,
  backgroundColor: 'transparent',
  color: 'inherit',
  padding: 0,
};

export interface CapabilityTreeRowOptions {
  draggable?: boolean;
  onDragStart?: (e: React.DragEvent) => void;
  onDragEnd?: () => void;
  onClick?: (e: React.MouseEvent) => void;
  onContextMenu?: (e: React.MouseEvent) => void;
  selected?: boolean;
  dimmed?: boolean;
  title?: string;
  testId?: string;
}

interface ChevronProps {
  hasChildren: boolean;
  isExpanded: boolean;
  onToggle: () => void;
}

const Chevron: React.FC<ChevronProps> = ({ hasChildren, isExpanded, onToggle }) => {
  if (!hasChildren) {
    return <span className={classes.chevronPlaceholder} />;
  }
  return (
    <ActionIcon
      variant="subtle"
      color="gray"
      size="xs"
      onClick={(e) => {
        e.stopPropagation();
        onToggle();
      }}
      aria-label={isExpanded ? 'Collapse' : 'Expand'}
    >
      <svg
        viewBox="0 0 24 24"
        fill="none"
        width="12"
        height="12"
        className={`${classes.chevron} ${isExpanded ? classes.chevronOpen : ''}`}
      >
        <path d="M9 18l6-6-6-6" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
      </svg>
    </ActionIcon>
  );
};

function rowClassName(options: CapabilityTreeRowOptions): string {
  return [classes.row, options.selected && classes.rowSelected, options.dimmed && classes.rowDimmed]
    .filter(Boolean)
    .join(' ');
}

interface CapabilityTreeRowProps {
  node: CapabilityTreeNode;
  isExpanded: boolean;
  onToggle: () => void;
  searchQuery: string;
  options: CapabilityTreeRowOptions;
  rightContent?: React.ReactNode;
}

const activateOnEnterOrSpace = (e: React.KeyboardEvent<HTMLDivElement>) => {
  if (e.key === 'Enter' || e.key === ' ') {
    e.preventDefault();
    e.currentTarget.click();
  }
};

export const CapabilityTreeRow: React.FC<CapabilityTreeRowProps> = ({
  node,
  isExpanded,
  onToggle,
  searchQuery,
  options,
  rightContent,
}) => {
  const level = node.capability.level || 'L1';
  const hasChildren = node.children.length > 0;

  return (
    <div
      role="treeitem"
      aria-selected={options.selected ?? false}
      aria-expanded={hasChildren ? isExpanded : undefined}
      tabIndex={options.onClick ? 0 : undefined}
      onKeyDown={options.onClick ? activateOnEnterOrSpace : undefined}
      className={rowClassName(options)}
      draggable={options.draggable}
      data-draggable={options.draggable ? 'true' : undefined}
      onDragStart={options.onDragStart}
      onDragEnd={options.onDragEnd}
      onClick={options.onClick}
      onContextMenu={options.onContextMenu}
      title={options.title}
      data-testid={options.testId}
    >
      <Chevron hasChildren={node.children.length > 0} isExpanded={isExpanded} onToggle={onToggle} />
      <span className={`${classes.level} ${classes[`level${level}`]}`}>{level}</span>
      <Highlight component="span" className={classes.name} highlight={searchQuery} highlightStyles={HIGHLIGHT_STYLES}>
        {node.capability.name}
      </Highlight>
      {rightContent}
    </div>
  );
};
