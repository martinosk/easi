import React, { useState, useMemo } from 'react';
import type { Component, View } from '../../../../api/types';
import { TreeSection } from '../TreeSection';
import { hasCustomColor } from '../../utils/treeUtils';
import type { EditingState } from '../../types';

interface ColorIndicatorProps {
  customColor: string | undefined;
}

const ColorIndicator: React.FC<ColorIndicatorProps> = ({ customColor }) => (
  <div
    data-testid="custom-color-indicator"
    style={{
      width: '10px',
      height: '10px',
      borderRadius: '2px',
      backgroundColor: customColor,
      display: 'inline-block',
      marginLeft: '8px',
      border: '1px solid rgba(0,0,0,0.1)',
    }}
  />
);

interface ApplicationsSectionProps {
  components: Component[];
  currentView: View | null;
  selectedNodeId: string | null;
  isExpanded: boolean;
  onToggle: () => void;
  onAddComponent?: () => void;
  onComponentSelect?: (componentId: string) => void;
  onComponentContextMenu: (e: React.MouseEvent, component: Component) => void;
  editingState: EditingState | null;
  setEditingState: (state: EditingState | null) => void;
  onRenameSubmit: () => void;
  editInputRef: React.RefObject<HTMLInputElement | null>;
}

export const ApplicationsSection: React.FC<ApplicationsSectionProps> = ({
  components,
  currentView,
  selectedNodeId,
  isExpanded,
  onToggle,
  onAddComponent,
  onComponentSelect,
  onComponentContextMenu,
  editingState,
  setEditingState,
  onRenameSubmit,
  editInputRef,
}) => {
  const [applicationSearch, setApplicationSearch] = useState('');

  const filteredComponents = useMemo(() => {
    if (!applicationSearch.trim()) {
      return components;
    }
    const searchLower = applicationSearch.toLowerCase();
    return components.filter(
      (c) =>
        c.name.toLowerCase().includes(searchLower) ||
        (c.description && c.description.toLowerCase().includes(searchLower))
    );
  }, [components, applicationSearch]);

  const handleComponentClick = (componentId: string) => {
    if (onComponentSelect) {
      onComponentSelect(componentId);
    }
  };

  return (
    <TreeSection
      label="Applications"
      count={components.length}
      isExpanded={isExpanded}
      onToggle={onToggle}
      onAdd={onAddComponent}
      addTitle="Create new application"
      addTestId="create-component-button"
    >
      <div className="tree-search">
        <input
          type="text"
          className="tree-search-input"
          placeholder="Search applications..."
          value={applicationSearch}
          onChange={(e) => setApplicationSearch(e.target.value)}
        />
        {applicationSearch && (
          <button
            className="tree-search-clear"
            onClick={() => setApplicationSearch('')}
            aria-label="Clear search"
          >
            ×
          </button>
        )}
      </div>
      <div className="tree-items">
        {filteredComponents.length === 0 ? (
          <div className="tree-item-empty">
            {components.length === 0 ? 'No applications' : 'No matches'}
          </div>
        ) : (
          filteredComponents.map((component) => {
            const isInCurrentView = currentView?.components.some(
              vc => vc.componentId === component.id
            );
            const isSelected = selectedNodeId === component.id;
            const isEditing = editingState?.componentId === component.id;

            if (isEditing) {
              return (
                <div key={component.id} className="tree-item-edit">
                  <span className="tree-item-icon">📦</span>
                  <input
                    ref={editInputRef}
                    type="text"
                    className="tree-item-input"
                    value={editingState.name}
                    onChange={(e) => setEditingState({ ...editingState, name: e.target.value })}
                    onBlur={onRenameSubmit}
                    onKeyDown={(e) => {
                      if (e.key === 'Enter') {
                        onRenameSubmit();
                      } else if (e.key === 'Escape') {
                        setEditingState(null);
                      }
                    }}
                    autoFocus
                  />
                </div>
              );
            }

            const viewComponent = currentView?.components.find(vc => vc.componentId === component.id);
            const customColor = viewComponent?.customColor;
            const showColorIndicator = hasCustomColor(currentView?.colorScheme, customColor);

            return (
              <button
                key={component.id}
                className={`tree-item ${isSelected ? 'selected' : ''} ${!isInCurrentView ? 'not-in-view' : ''}`}
                onClick={() => handleComponentClick(component.id)}
                onContextMenu={(e) => onComponentContextMenu(e, component)}
                title={isInCurrentView ? component.name : `${component.name} (not in current view)`}
                draggable={!isInCurrentView}
                onDragStart={(e) => {
                  if (!isInCurrentView) {
                    e.dataTransfer.setData('componentId', component.id);
                    e.dataTransfer.effectAllowed = 'copy';
                  }
                }}
              >
                <span className="tree-item-icon">📦</span>
                <span className="tree-item-label">{component.name}</span>
                {showColorIndicator && <ColorIndicator customColor={customColor} />}
              </button>
            );
          })
        )}
      </div>
    </TreeSection>
  );
};
