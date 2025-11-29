import { useMemo } from 'react';
import { useDroppable } from '@dnd-kit/core';
import type { Capability } from '../../../api/types';

const L1_COLOR = '#3b82f6';

export interface DomainGridProps {
  capabilities: Capability[];
  onCapabilityClick: (capability: Capability) => void;
}

export function DomainGrid({ capabilities, onCapabilityClick }: DomainGridProps) {
  const { setNodeRef, isOver } = useDroppable({
    id: 'domain-grid-droppable',
  });

  const l1Capabilities = useMemo(() => {
    return capabilities
      .filter((cap) => cap.level === 'L1')
      .sort((a, b) => a.name.localeCompare(b.name));
  }, [capabilities]);

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
      <div
        style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fill, minmax(200px, 1fr))',
          gap: '1rem',
          padding: '1rem',
        }}
      >
        {l1Capabilities.map((capability) => (
          <button
            key={capability.id}
            type="button"
            onClick={() => onCapabilityClick(capability)}
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
        ))}
      </div>
    </div>
  );
}
