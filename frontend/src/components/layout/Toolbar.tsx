import React from 'react';
import { EdgeTypeSelector, ColorSchemeSelector } from '../../features/views';

export const Toolbar: React.FC = () => {
  return (
    <div className="toolbar">
      <div className="toolbar-left">
        <EdgeTypeSelector />
        <ColorSchemeSelector />
      </div>
    </div>
  );
};
