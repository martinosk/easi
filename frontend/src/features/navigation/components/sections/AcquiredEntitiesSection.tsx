import React, { useState, useMemo } from 'react';
import type { AcquiredEntity } from '../../../../api/types';
import { TreeSection } from '../TreeSection';

interface AcquiredEntitiesSectionProps {
  acquiredEntities: AcquiredEntity[];
  selectedEntityId: string | null;
  isExpanded: boolean;
  onToggle: () => void;
  onAddEntity?: () => void;
  onEntitySelect?: (entityId: string) => void;
  onEntityContextMenu: (e: React.MouseEvent, entity: AcquiredEntity) => void;
}

const formatAcquisitionYear = (acquisitionDate: string | undefined): string => {
  if (!acquisitionDate) return '';
  try {
    const year = new Date(acquisitionDate).getFullYear();
    return ` (${year})`;
  } catch {
    return '';
  }
};

export const AcquiredEntitiesSection: React.FC<AcquiredEntitiesSectionProps> = ({
  acquiredEntities,
  selectedEntityId,
  isExpanded,
  onToggle,
  onAddEntity,
  onEntitySelect,
  onEntityContextMenu,
}) => {
  const [search, setSearch] = useState('');

  const filteredEntities = useMemo(() => {
    if (!search.trim()) {
      return acquiredEntities;
    }
    const searchLower = search.toLowerCase();
    return acquiredEntities.filter(
      (e) =>
        e.name.toLowerCase().includes(searchLower) ||
        (e.notes && e.notes.toLowerCase().includes(searchLower))
    );
  }, [acquiredEntities, search]);

  const handleEntityClick = (entityId: string) => {
    if (onEntitySelect) {
      onEntitySelect(entityId);
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
      <div className="tree-search">
        <input
          type="text"
          className="tree-search-input"
          placeholder="Search acquired entities..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
        />
        {search && (
          <button
            className="tree-search-clear"
            onClick={() => setSearch('')}
            aria-label="Clear search"
          >
            x
          </button>
        )}
      </div>
      <div className="tree-items">
        {filteredEntities.length === 0 ? (
          <div className="tree-item-empty">
            {acquiredEntities.length === 0 ? 'No acquired entities' : 'No matches'}
          </div>
        ) : (
          filteredEntities.map((entity) => {
            const isSelected = selectedEntityId === entity.id;

            return (
              <button
                key={entity.id}
                className={`tree-item ${isSelected ? 'selected' : ''}`}
                onClick={() => handleEntityClick(entity.id)}
                onContextMenu={(e) => onEntityContextMenu(e, entity)}
                title={entity.name}
                draggable
                onDragStart={(e) => {
                  e.dataTransfer.setData('acquiredEntityId', entity.id);
                  e.dataTransfer.effectAllowed = 'copy';
                }}
              >
                <span className="tree-item-icon">🏢</span>
                <span className="tree-item-label">
                  {entity.name}
                  {formatAcquisitionYear(entity.acquisitionDate)}
                </span>
              </button>
            );
          })
        )}
      </div>
    </TreeSection>
  );
};
