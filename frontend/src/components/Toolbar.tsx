import React from 'react';

interface ToolbarProps {
  onAddComponent: () => void;
  onFitView: () => void;
}

export const Toolbar: React.FC<ToolbarProps> = ({ onAddComponent, onFitView }) => {
  return (
    <div className="toolbar">
      <div className="toolbar-left">
        <h1 className="toolbar-title">Component Modeler</h1>
      </div>
      <div className="toolbar-right">
        <button className="btn btn-secondary" onClick={onFitView}>
          Fit View
        </button>
        <button className="btn btn-primary" onClick={onAddComponent}>
          + Add Component
        </button>
      </div>
    </div>
  );
};
