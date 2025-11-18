import React from 'react';
import { EdgeTypeSelector } from './EdgeTypeSelector';
import { LayoutDirectionSelector } from './LayoutDirectionSelector';
import { AutoLayoutButton } from './AutoLayoutButton';

export const Toolbar: React.FC = () => {
  return (
    <div className="toolbar">
      <div className="toolbar-left">
        <h1 className="toolbar-title">Component Modeler</h1>
      </div>
      <div className="toolbar-right">
        <EdgeTypeSelector />
        <LayoutDirectionSelector />
        <AutoLayoutButton />
      </div>
    </div>
  );
};
