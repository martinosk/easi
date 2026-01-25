import React, { useState, useMemo } from 'react';
import type { AcquiredEntity, View } from '../../../../api/types';
import { TreeSection } from '../TreeSection';
import { TreeSearchInput } from '../shared/TreeSearchInput';
import { TreeItemList } from '../shared/TreeItemList';
import { ORIGIN_ENTITY_PREFIXES } from '../../../canvas/utils/nodeFactory';

interface AcquiredEntitiesSectionProps {
  acquiredEntities: AcquiredEntity[];
  currentView: View | null;
  selectedEntityId: string | null;
  isExpanded: boolean;
  onToggle: () => void;
  onAddEntity?: () => void;
  onEntitySelect?: (entityId: string) => void;
  onEntityContextMenu: (e: React.MouseEvent, entity: AcquiredEntity) => void;
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
    (e) =>
      e.name.toLowerCase().includes(searchLower) ||
      (e.notes && e.notes.toLowerCase().includes(searchLower))
  );
}

export const AcquiredEntitiesSection: React.FC<AcquiredEntitiesSectionProps> = ({
  acquiredEntities,
  currentView,
  selectedEntityId,
  isExpanded,
  onToggle,
  onAddEntity,
  onEntitySelect,
  onEntityContextMenu,
}) => {
  const [search, setSearch] = useState('');

  const entityIdsOnCanvas = useMemo(() => {
    const viewOriginEntityIds = new Set(
      (currentView?.originEntities ?? []).map((oe) => oe.originEntityId)
    );
    const onCanvas = new Set<string>();
    for (const entity of acquiredEntities) {
      const nodeId = `${ORIGIN_ENTITY_PREFIXES.acquired}${entity.id}`;
      if (viewOriginEntityIds.has(nodeId)) {
        onCanvas.add(entity.id);
      }
    }
    return onCanvas;
  }, [acquiredEntities, currentView?.originEntities]);

  const filteredEntities = useMemo(
    () => filterEntities(acquiredEntities, search),
    [acquiredEntities, search]
  );

  const hasNoEntities = acquiredEntities.length === 0;
  const emptyMessage = hasNoEntities ? 'No acquired entities' : 'No matches';

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
      <TreeSearchInput
        value={search}
        onChange={setSearch}
        placeholder="Search acquired entities..."
      />
      <div className="tree-items">
        <TreeItemList
          items={filteredEntities}
          emptyMessage={emptyMessage}
          icon="🏢"
          dragDataKey="acquiredEntityId"
          isSelected={(entity) => selectedEntityId === entity.id}
          isInView={(entity) => !currentView || entityIdsOnCanvas.has(entity.id)}
          getTitle={(entity, isInView) =>
            isInView ? entity.name : `${entity.name} (not on canvas)`
          }
          renderLabel={(entity) => (
            <>
              {entity.name}
              {formatAcquisitionYear(entity.acquisitionDate)}
            </>
          )}
          onSelect={(entity) => onEntitySelect?.(entity.id)}
          onContextMenu={onEntityContextMenu}
        />
      </div>
    </TreeSection>
  );
};
