import { useMemo } from 'react';
import { useDroppable } from '@dnd-kit/core';
import { SortableContext, useSortable, rectSortingStrategy } from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';
import type { Capability, Position, CapabilityId } from '../../../api/types';
import './visualization.css';

const L1_COLOR = '#3b82f6';

export interface PositionMap {
  [capabilityId: string]: Position;
}

export interface DomainGridProps {
  capabilities: Capability[];
  onCapabilityClick: (capability: Capability) => void;
  positions?: PositionMap;
  onPositionChange?: (capabilityId: CapabilityId, x: number, y: number) => void;
}

interface SortableCapabilityItemProps {
  capability: Capability;
  onClick: () => void;
}

function SortableCapabilityItem({ capability, onClick }: SortableCapabilityItemProps) {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({ id: capability.id });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    backgroundColor: L1_COLOR,
    color: 'white',
    padding: '1.5rem',
    borderRadius: '0.5rem',
    border: 'none',
    cursor: 'grab',
    minHeight: '100px',
    fontWeight: 500,
    textAlign: 'left' as const,
    opacity: isDragging ? 0.5 : 1,
  };

  return (
    <button
      ref={setNodeRef}
      type="button"
      className="capability-item"
      onClick={onClick}
      data-sortable="true"
      style={style}
      {...attributes}
      {...listeners}
    >
      {capability.name}
    </button>
  );
}

interface StaticCapabilityItemProps {
  capability: Capability;
  onClick: () => void;
}

function StaticCapabilityItem({ capability, onClick }: StaticCapabilityItemProps) {
  return (
    <button
      type="button"
      className="capability-item"
      onClick={onClick}
      style={{
        backgroundColor: L1_COLOR,
        color: 'white',
        padding: '1.5rem',
        borderRadius: '0.5rem',
        border: 'none',
        cursor: 'pointer',
        minHeight: '100px',
        fontWeight: 500,
        textAlign: 'left',
      }}
    >
      {capability.name}
    </button>
  );
}

export function DomainGrid({ capabilities, onCapabilityClick, positions }: DomainGridProps) {
  const { setNodeRef, isOver } = useDroppable({
    id: 'domain-grid-droppable',
  });

  const l1Capabilities = useMemo(() => {
    const filtered = capabilities.filter((cap) => cap.level === 'L1');

    if (positions && Object.keys(positions).length > 0) {
      return [...filtered].sort((a, b) => {
        const posA = positions[a.id];
        const posB = positions[b.id];

        if (posA && posB) {
          if (posA.y !== posB.y) return posA.y - posB.y;
          return posA.x - posB.x;
        }
        if (posA) return -1;
        if (posB) return 1;
        return a.name.localeCompare(b.name);
      });
    }

    return filtered.sort((a, b) => a.name.localeCompare(b.name));
  }, [capabilities, positions]);

  const sortableIds = useMemo(() => l1Capabilities.map((c) => c.id), [l1Capabilities]);

  const hasSortablePositions = positions && Object.keys(positions).length > 0;

  return (
    <div
      ref={setNodeRef}
      className="domain-grid"
      data-dnd-context="true"
      style={{
        minHeight: '200px',
        border: isOver ? '2px dashed #3b82f6' : '2px dashed transparent',
        borderRadius: '0.5rem',
        transition: 'border-color 0.2s',
      }}
    >
      {hasSortablePositions ? (
        <SortableContext items={sortableIds} strategy={rectSortingStrategy}>
          <div
            style={{
              display: 'grid',
              gridTemplateColumns: 'repeat(auto-fill, minmax(200px, 1fr))',
              gap: '1rem',
              padding: '1rem',
            }}
          >
            {l1Capabilities.map((capability) => (
              <SortableCapabilityItem
                key={capability.id}
                capability={capability}
                onClick={() => onCapabilityClick(capability)}
              />
            ))}
          </div>
        </SortableContext>
      ) : (
        <div
          style={{
            display: 'grid',
            gridTemplateColumns: 'repeat(auto-fill, minmax(200px, 1fr))',
            gap: '1rem',
            padding: '1rem',
          }}
        >
          {l1Capabilities.map((capability) => (
            <StaticCapabilityItem
              key={capability.id}
              capability={capability}
              onClick={() => onCapabilityClick(capability)}
            />
          ))}
        </div>
      )}
    </div>
  );
}
