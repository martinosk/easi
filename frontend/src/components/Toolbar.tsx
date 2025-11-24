import React from 'react';
import { EdgeTypeSelector } from './EdgeTypeSelector';
import { LayoutDirectionSelector } from './LayoutDirectionSelector';
import { AutoLayoutButton } from './AutoLayoutButton';

interface ToolbarProps {
  onOpenReleaseNotes?: () => void;
}

export const Toolbar: React.FC<ToolbarProps> = ({ onOpenReleaseNotes }) => {
  return (
    <div className="toolbar">
      <div className="toolbar-left">
        <h1 className="toolbar-title">Component Modeler</h1>
      </div>
      <div className="toolbar-right">
        <EdgeTypeSelector />
        <LayoutDirectionSelector />
        <AutoLayoutButton />
        {onOpenReleaseNotes && (
          <button
            type="button"
            className="btn btn-secondary btn-small toolbar-menu-btn"
            onClick={onOpenReleaseNotes}
            title="View release notes"
          >
            What's New
          </button>
        )}
      </div>
    </div>
  );
};
