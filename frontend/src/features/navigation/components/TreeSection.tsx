import React from 'react';

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
    <div className="tree-category">
      <div className="category-header-wrapper">
        <button className="category-header" onClick={onToggle}>
          <span className="category-icon">{isExpanded ? '▼' : '▶'}</span>
          <span className="category-label">{label}</span>
          <span className="category-count">{count}</span>
        </button>
        {onAdd && (
          <button
            className="add-view-btn"
            onClick={onAdd}
            title={addTitle}
            data-testid={addTestId}
          >
            +
          </button>
        )}
      </div>
      {isExpanded && children}
    </div>
  );
};
