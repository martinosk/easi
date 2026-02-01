import React, { useState, useMemo } from 'react';
import type { Component, View } from '../../../../api/types';
import { TreeSection } from '../TreeSection';
import { TreeSearchInput } from '../shared/TreeSearchInput';
import { hasCustomColor } from '../../utils/treeUtils';
import type { EditingState, TreeMultiSelectProps } from '../../types';

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

function filterComponents(components: Component[], search: string): Component[] {
  if (!search.trim()) return components;
  const searchLower = search.toLowerCase();
  return components.filter(
    (c) =>
      c.name.toLowerCase().includes(searchLower) ||
      (c.description && c.description.toLowerCase().includes(searchLower))
  );
}

interface EditingItemProps {
  component: Component;
  editingState: EditingState;
  setEditingState: (state: EditingState | null) => void;
  onRenameSubmit: () => void;
  editInputRef: React.RefObject<HTMLInputElement | null>;
}

const EditingItem: React.FC<EditingItemProps> = ({
  component, editingState, setEditingState, onRenameSubmit, editInputRef,
}) => (
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
        if (e.key === 'Enter') onRenameSubmit();
        else if (e.key === 'Escape') setEditingState(null);
      }}
      autoFocus
    />
  </div>
);

interface ComponentItemProps {
  component: Component;
  isSelected: boolean;
  isInView: boolean;
  showColorIndicator: boolean;
  customColor: string | undefined;
  onClick: (e: React.MouseEvent) => void;
  onContextMenu: (e: React.MouseEvent) => void;
  onDragStart: (e: React.DragEvent) => void;
}

const ComponentItem: React.FC<ComponentItemProps> = ({
  component, isSelected, isInView, showColorIndicator, customColor,
  onClick, onContextMenu, onDragStart,
}) => (
  <button
    className={`tree-item ${isSelected ? 'selected' : ''} ${!isInView ? 'not-in-view' : ''}`}
    onClick={onClick}
    onContextMenu={onContextMenu}
    title={isInView ? component.name : `${component.name} (not in current view)`}
    draggable
    onDragStart={onDragStart}
  >
    <span className="tree-item-icon">📦</span>
    <span className="tree-item-label">{component.name}</span>
    {showColorIndicator && <ColorIndicator customColor={customColor} />}
  </button>
);

function buildComponentViewMap(currentView: View | null): Map<string, { customColor?: string }> {
  const map = new Map<string, { customColor?: string }>();
  for (const vc of currentView?.components ?? []) {
    map.set(vc.componentId, { customColor: vc.customColor });
  }
  return map;
}

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
  multiSelect: TreeMultiSelectProps;
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
  multiSelect,
}) => {
  const [applicationSearch, setApplicationSearch] = useState('');

  const filteredComponents = useMemo(
    () => filterComponents(components, applicationSearch),
    [components, applicationSearch]
  );

  const visibleItems = useMemo(
    () => filteredComponents.map((c) => ({
      id: c.id, name: c.name, type: 'component' as const, links: c._links,
    })),
    [filteredComponents]
  );

  const componentViewMap = useMemo(
    () => buildComponentViewMap(currentView),
    [currentView]
  );

  const handleSelect = (component: Component, event: React.MouseEvent) => {
    const result = multiSelect.handleItemClick(
      { id: component.id, name: component.name, type: 'component', links: component._links },
      'applications',
      visibleItems,
      event
    );
    if (result === 'single') {
      onComponentSelect?.(component.id);
    }
  };

  const handleContextMenu = (e: React.MouseEvent, component: Component) => {
    const handled = multiSelect.handleContextMenu(e, component.id, multiSelect.selectedItems);
    if (!handled) {
      onComponentContextMenu(e, component);
    }
  };

  const handleDragStart = (e: React.DragEvent, component: Component) => {
    const handled = multiSelect.handleDragStart(e, component.id);
    if (!handled && !componentViewMap.has(component.id)) {
      e.dataTransfer.setData('componentId', component.id);
      e.dataTransfer.effectAllowed = 'copy';
    }
  };

  const emptyMessage = components.length === 0 ? 'No applications' : 'No matches';

  const renderComponent = (component: Component) => {
    if (editingState?.componentId === component.id) {
      return (
        <EditingItem
          key={component.id}
          component={component}
          editingState={editingState}
          setEditingState={setEditingState}
          onRenameSubmit={onRenameSubmit}
          editInputRef={editInputRef}
        />
      );
    }

    const viewEntry = componentViewMap.get(component.id);

    return (
      <ComponentItem
        key={component.id}
        component={component}
        isSelected={selectedNodeId === component.id || multiSelect.isMultiSelected(component.id)}
        isInView={!!viewEntry}
        showColorIndicator={hasCustomColor(currentView?.colorScheme, viewEntry?.customColor)}
        customColor={viewEntry?.customColor}
        onClick={(e) => handleSelect(component, e)}
        onContextMenu={(e) => handleContextMenu(e, component)}
        onDragStart={(e) => handleDragStart(e, component)}
      />
    );
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
      <TreeSearchInput
        value={applicationSearch}
        onChange={setApplicationSearch}
        placeholder="Search applications..."
      />
      <div className="tree-items">
        {filteredComponents.length === 0
          ? <div className="tree-item-empty">{emptyMessage}</div>
          : filteredComponents.map(renderComponent)
        }
      </div>
    </TreeSection>
  );
};
