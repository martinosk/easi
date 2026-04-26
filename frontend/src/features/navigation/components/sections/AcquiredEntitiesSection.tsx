import React, { useMemo, useState } from 'react';
import type { AcquiredEntity, View } from '../../../../api/types';
import type { TreeMultiSelectProps } from '../../types';
import { TreeItemList } from '../shared/TreeItemList';
import { TreeSearchInput } from '../shared/TreeSearchInput';
import { TreeSection } from '../TreeSection';

interface AcquiredEntitiesSectionProps {
  acquiredEntities: AcquiredEntity[];
  currentView: View | null;
  originEntitiesInView?: Set<string>;
  selectedEntityId: string | null;
  isExpanded: boolean;
  onToggle: () => void;
  onAddEntity?: () => void;
  onEntitySelect?: (entityId: string) => void;
  onEntityContextMenu: (e: React.MouseEvent, entity: AcquiredEntity) => void;
  multiSelect: TreeMultiSelectProps;
}

function defaultOriginEntitiesInView(currentView: View | null): Set<string> {
  return new Set((currentView?.originEntities ?? []).map((oe) => oe.originEntityId));
}

function formatAcquisitionYear(acquisitionDate: string | undefined): string {
  if (!acquisitionDate) return '';
  try {
    const year = new Date(acquisitionDate).getFullYear();
    return ` (${year})`;
  } catch {
    return '';
  }
}

function filterEntities(entities: AcquiredEntity[], search: string): AcquiredEntity[] {
  if (!search.trim()) return entities;
  const searchLower = search.toLowerCase();
  return entities.filter(
    (e) => e.name.toLowerCase().includes(searchLower) || (e.notes && e.notes.toLowerCase().includes(searchLower)),
  );
}

export const AcquiredEntitiesSection: React.FC<AcquiredEntitiesSectionProps> = ({
  acquiredEntities,
  currentView,
  originEntitiesInView,
  selectedEntityId,
  isExpanded,
  onToggle,
  onAddEntity,
  onEntitySelect,
  onEntityContextMenu,
  multiSelect,
}) => {
  const [search, setSearch] = useState('');
  const entityIdsOnCanvas = useMemo(
    () => originEntitiesInView ?? defaultOriginEntitiesInView(currentView),
    [originEntitiesInView, currentView],
  );

  const filteredEntities = useMemo(() => filterEntities(acquiredEntities, search), [acquiredEntities, search]);

  const visibleItems = useMemo(
    () =>
      filteredEntities.map((e) => ({
        id: e.id,
        name: e.name,
        type: 'acquired' as const,
        links: e._links,
      })),
    [filteredEntities],
  );

  const emptyMessage = acquiredEntities.length === 0 ? 'No acquired entities' : 'No matches';

  const handleSelect = (entity: AcquiredEntity, event: React.MouseEvent) => {
    const result = multiSelect.handleItemClick(
      { id: entity.id, name: entity.name, type: 'acquired', links: entity._links },
      'acquired',
      visibleItems,
      event,
    );
    if (result === 'single') {
      onEntitySelect?.(entity.id);
    }
  };

  const handleContextMenu = (e: React.MouseEvent, entity: AcquiredEntity) => {
    const handled = multiSelect.handleContextMenu(e, entity.id, multiSelect.selectedItems);
    if (!handled) {
      onEntityContextMenu(e, entity);
    }
  };

  const handleDragStart = (e: React.DragEvent, entity: AcquiredEntity) => {
    const handled = multiSelect.handleDragStart(e, entity.id);
    if (!handled) {
      e.dataTransfer.setData('acquiredEntityId', entity.id);
      e.dataTransfer.effectAllowed = 'copy';
    }
  };

  return (
    <TreeSection
      label="Acquired Entities"
      count={acquiredEntities.length}
      isExpanded={isExpanded}
      onToggle={onToggle}
      onAdd={onAddEntity}
      addTitle="Create new acquired entity"
      addTestId="create-acquired-entity-button"
    >
      <TreeSearchInput value={search} onChange={setSearch} placeholder="Search acquired entities..." />
      <div className="tree-items">
        <TreeItemList
          items={filteredEntities}
          emptyMessage={emptyMessage}
          icon="🏢"
          dragDataKey="acquiredEntityId"
          isSelected={(entity) => selectedEntityId === entity.id || multiSelect.isMultiSelected(entity.id)}
          isInView={(entity) => !currentView || entityIdsOnCanvas.has(entity.id)}
          getTitle={(entity, isInView) => (isInView ? entity.name : `${entity.name} (not on canvas)`)}
          renderLabel={(entity) => (
            <>
              {entity.name}
              {formatAcquisitionYear(entity.acquisitionDate)}
            </>
          )}
          onSelect={handleSelect}
          onContextMenu={handleContextMenu}
          onDragStart={handleDragStart}
        />
      </div>
    </TreeSection>
  );
};
