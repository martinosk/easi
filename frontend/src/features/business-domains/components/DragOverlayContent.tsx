import { DragOverlay } from '@dnd-kit/core';
import type { Capability } from '../../../api/types';

interface DragOverlayContentProps {
  activeCapability: Capability | null;
}

export function DragOverlayContent({ activeCapability }: DragOverlayContentProps) {
  return (
    <DragOverlay>
      {activeCapability && (
        <div
          style={{
            backgroundColor: '#3b82f6',
            color: 'white',
            padding: '0.75rem 1rem',
            borderRadius: '0.5rem',
            boxShadow: '0 4px 6px -1px rgba(0, 0, 0, 0.1)',
            fontWeight: 500,
          }}
        >
          {activeCapability.name}
        </div>
      )}
    </DragOverlay>
  );
}
