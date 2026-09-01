import { ColorSwatch, Select, TextInput, UnstyledButton } from '@mantine/core';
import { IconBox } from '@tabler/icons-react';
import React, { useMemo, useState } from 'react';
import type { Component, HostingClassification, OwnershipState, View } from '../../../../api/types';
import { TreeSearchInput } from '../../../../components/shared';
import { HOSTING_CLASSIFICATION_LABELS } from '../../../components/components/ComponentHostingSection';
import { OWNERSHIP_STATE_LABELS } from '../../../components/components/ComponentOwnershipSection';
import { StatisticsSummary } from '../../../components/components/StatisticsSummary';
import { OnePagerIncompleteIndicator } from '../../../one-pagers/components/OnePagerIncompleteIndicator';
import { useOnePagerCompleteness } from '../../../one-pagers/hooks/useOnePagerCompleteness';
import type { EditingState, TreeMultiSelectProps } from '../../types';
import { hasCustomColor } from '../../utils/treeUtils';
import classes from '../shared/TreeItem.module.css';
import { TreeSection } from '../TreeSection';

interface ColorIndicatorProps {
  customColor: string | undefined;
}

const ColorIndicator: React.FC<ColorIndicatorProps> = ({ customColor }) => (
  <ColorSwatch data-testid="custom-color-indicator" color={customColor ?? ''} size="xs" radius="xs" ml="sm" />
);

function filterComponents(
  components: Component[],
  search: string,
  ownershipState: string | null,
  hosting: string | null,
): Component[] {
  const byOwnership = ownershipState ? components.filter((c) => c.ownershipState === ownershipState) : components;
  const byHosting = hosting ? byOwnership.filter((c) => c.hosting === hosting) : byOwnership;
  if (!search.trim()) return byHosting;
  const searchLower = search.toLowerCase();
  return byHosting.filter(
    (c) => c.name.toLowerCase().includes(searchLower) || c.description?.toLowerCase().includes(searchLower),
  );
}

const OWNERSHIP_FILTER_OPTIONS = (Object.keys(OWNERSHIP_STATE_LABELS) as OwnershipState[]).map((state) => ({
  value: state,
  label: OWNERSHIP_STATE_LABELS[state],
}));

const HOSTING_FILTER_OPTIONS = (Object.keys(HOSTING_CLASSIFICATION_LABELS) as HostingClassification[]).map(
  (hosting) => ({
    value: hosting,
    label: HOSTING_CLASSIFICATION_LABELS[hosting],
  }),
);

interface EditingItemProps {
  component: Component;
  editingState: EditingState;
  setEditingState: (state: EditingState | null) => void;
  onRenameSubmit: () => void;
  editInputRef: React.RefObject<HTMLInputElement | null>;
}

const EditingItem: React.FC<EditingItemProps> = ({
  component,
  editingState,
  setEditingState,
  onRenameSubmit,
  editInputRef,
}) => (
  <div key={component.id} className={classes.edit}>
    <span className={classes.icon}>
      <IconBox size={16} stroke={1.75} />
    </span>
    <TextInput
      ref={editInputRef}
      className={classes.input}
      value={editingState.name}
      onChange={(e) => setEditingState({ ...editingState, name: e.currentTarget.value })}
      onBlur={onRenameSubmit}
      onKeyDown={(e) => {
        if (e.key === 'Enter') onRenameSubmit();
        else if (e.key === 'Escape') setEditingState(null);
      }}
      size="xs"
      data-autofocus
    />
  </div>
);

interface ComponentItemProps {
  component: Component;
  onePagerComplete?: boolean;
  isSelected: boolean;
  isInView: boolean;
  showColorIndicator: boolean;
  customColor: string | undefined;
  onClick: (e: React.MouseEvent) => void;
  onContextMenu: (e: React.MouseEvent) => void;
  onDragStart: (e: React.DragEvent) => void;
}

const ComponentItem: React.FC<ComponentItemProps> = ({
  component,
  onePagerComplete,
  isSelected,
  isInView,
  showColorIndicator,
  customColor,
  onClick,
  onContextMenu,
  onDragStart,
}) => (
  <UnstyledButton
    component="button"
    type="button"
    className={classes.item}
    data-testid="tree-item"
    data-selected={isSelected || undefined}
    data-in-view={isInView}
    onClick={onClick}
    onContextMenu={onContextMenu}
    title={isInView ? component.name : `${component.name} (not in current view)`}
    draggable
    onDragStart={onDragStart}
  >
    <span className={classes.icon}>
      <IconBox size={16} stroke={1.75} />
    </span>
    <span className={classes.label}>{component.name}</span>
    <OnePagerIncompleteIndicator id={component.id} complete={onePagerComplete} />
    {showColorIndicator && <ColorIndicator customColor={customColor} />}
  </UnstyledButton>
);

function buildComponentColorMap(currentView: View | null): Map<string, { customColor?: string }> {
  const map = new Map<string, { customColor?: string }>();
  for (const vc of currentView?.components ?? []) {
    map.set(vc.componentId, { customColor: vc.customColor });
  }
  return map;
}

interface ApplicationsSectionProps {
  components: Component[];
  currentView: View | null;
  componentsInView?: Set<string>;
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

function defaultComponentsInView(currentView: View | null): Set<string> {
  return new Set((currentView?.components ?? []).map((c) => c.componentId));
}

export const ApplicationsSection: React.FC<ApplicationsSectionProps> = ({
  components,
  currentView,
  componentsInView,
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
  const [ownershipFilter, setOwnershipFilter] = useState<string | null>(null);
  const [hostingFilter, setHostingFilter] = useState<string | null>(null);
  const { data: onePagerCompleteness } = useOnePagerCompleteness('application');

  const filteredComponents = useMemo(
    () => filterComponents(components, applicationSearch, ownershipFilter, hostingFilter),
    [components, applicationSearch, ownershipFilter, hostingFilter],
  );

  const visibleItems = useMemo(
    () =>
      filteredComponents.map((c) => ({
        id: c.id,
        name: c.name,
        type: 'component' as const,
        links: c._links,
      })),
    [filteredComponents],
  );

  const componentColorMap = useMemo(() => buildComponentColorMap(currentView), [currentView]);
  const effectiveComponentsInView = useMemo(
    () => componentsInView ?? defaultComponentsInView(currentView),
    [componentsInView, currentView],
  );

  const handleSelect = (component: Component, event: React.MouseEvent) => {
    const result = multiSelect.handleItemClick(
      { id: component.id, name: component.name, type: 'component', links: component._links },
      'applications',
      visibleItems,
      event,
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
    if (!handled && !effectiveComponentsInView.has(component.id)) {
      e.dataTransfer.setData('componentId', component.id);
      e.dataTransfer.effectAllowed = 'copy';
    }
  };

  const emptyMessage = components.length === 0 && !ownershipFilter && !hostingFilter ? 'No applications' : 'No matches';

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

    const colorEntry = componentColorMap.get(component.id);

    return (
      <ComponentItem
        key={component.id}
        component={component}
        onePagerComplete={onePagerCompleteness?.get(component.id)}
        isSelected={selectedNodeId === component.id || multiSelect.isMultiSelected(component.id)}
        isInView={effectiveComponentsInView.has(component.id)}
        showColorIndicator={hasCustomColor(currentView?.colorScheme, colorEntry?.customColor)}
        customColor={colorEntry?.customColor}
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
      <StatisticsSummary />
      <TreeSearchInput value={applicationSearch} onChange={setApplicationSearch} placeholder="Search applications..." />
      <Select
        size="xs"
        px="sm"
        pb="xs"
        clearable
        placeholder="Filter by ownership"
        data={OWNERSHIP_FILTER_OPTIONS}
        value={ownershipFilter}
        onChange={setOwnershipFilter}
        data-testid="ownership-filter"
      />
      <Select
        size="xs"
        px="sm"
        pb="xs"
        clearable
        placeholder="Filter by hosting"
        data={HOSTING_FILTER_OPTIONS}
        value={hostingFilter}
        onChange={setHostingFilter}
        data-testid="hosting-filter"
      />
      <div className={classes.list}>
        {filteredComponents.length === 0 ? (
          <div className={classes.empty}>{emptyMessage}</div>
        ) : (
          filteredComponents.map(renderComponent)
        )}
      </div>
    </TreeSection>
  );
};
