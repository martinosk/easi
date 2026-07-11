import { Handle, Position } from '@xyflow/react';
import React from 'react';
import type { OriginEntityType } from '../../constants/entityIdentifiers';

export type { OriginEntityType };

export interface OriginEntityNodeData {
  label: string;
  entityType: OriginEntityType;
  isSelected: boolean;
  subtitle?: string;
}

const ENTITY_ICONS: Record<OriginEntityType, React.ReactNode> = {
  acquired: (
    <svg width="16" height="16" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
      <rect x="2" y="4" width="12" height="10" rx="1" stroke="currentColor" strokeWidth="1.5" fill="none" />
      <path d="M4 4V3C4 2.44772 4.44772 2 5 2H11C11.5523 2 12 2.44772 12 3V4" stroke="currentColor" strokeWidth="1.5" />
      <line x1="5" y1="7" x2="11" y2="7" stroke="currentColor" strokeWidth="1.5" />
      <line x1="5" y1="10" x2="9" y2="10" stroke="currentColor" strokeWidth="1.5" />
    </svg>
  ),
  vendor: (
    <svg width="16" height="16" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
      <rect x="1" y="5" width="14" height="10" rx="1" stroke="currentColor" strokeWidth="1.5" fill="none" />
      <path d="M1 5L8 1L15 5" stroke="currentColor" strokeWidth="1.5" />
      <rect x="6" y="9" width="4" height="6" stroke="currentColor" strokeWidth="1.5" fill="none" />
    </svg>
  ),
  team: (
    <svg width="16" height="16" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
      <circle cx="8" cy="5" r="3" stroke="currentColor" strokeWidth="1.5" fill="none" />
      <path d="M2 14C2 11.2386 4.68629 9 8 9C11.3137 9 14 11.2386 14 14" stroke="currentColor" strokeWidth="1.5" />
    </svg>
  ),
};

const ENTITY_LABELS: Record<OriginEntityType, string> = {
  acquired: 'Acquired Entity',
  vendor: 'Vendor',
  team: 'Internal Team',
};

export const OriginEntityNode: React.FC<{ data: OriginEntityNodeData; id: string; selected?: boolean }> = ({
  data,
  id,
  selected,
}) => {
  const isSelected = data.isSelected || !!selected;

  const nodeClassName = [
    'origin-entity-node',
    `origin-entity-node-${data.entityType}`,
    isSelected ? 'origin-entity-node-selected' : '',
  ]
    .filter(Boolean)
    .join(' ');

  return (
    <div className={nodeClassName} data-origin-entity-id={id}>
      <Handle
        type="source"
        position={Position.Top}
        id="top"
        className="origin-entity-handle origin-entity-handle-top"
      />
      <Handle
        type="target"
        position={Position.Top}
        id="top"
        className="origin-entity-handle origin-entity-handle-top"
      />

      <Handle
        type="source"
        position={Position.Left}
        id="left"
        className="origin-entity-handle origin-entity-handle-left"
      />
      <Handle
        type="target"
        position={Position.Left}
        id="left"
        className="origin-entity-handle origin-entity-handle-left"
      />

      <div className="origin-entity-node-content">
        <div className="origin-entity-node-header">
          <span className="origin-entity-node-icon">{ENTITY_ICONS[data.entityType]}</span>
          <span className="origin-entity-node-label">{data.label}</span>
        </div>
        <div className="origin-entity-node-type">
          {ENTITY_LABELS[data.entityType]}
          {data.subtitle && ` - ${data.subtitle}`}
        </div>
      </div>

      <Handle
        type="source"
        position={Position.Right}
        id="right"
        className="origin-entity-handle origin-entity-handle-right"
      />
      <Handle
        type="target"
        position={Position.Right}
        id="right"
        className="origin-entity-handle origin-entity-handle-right"
      />

      <Handle
        type="source"
        position={Position.Bottom}
        id="bottom"
        className="origin-entity-handle origin-entity-handle-bottom"
      />
      <Handle
        type="target"
        position={Position.Bottom}
        id="bottom"
        className="origin-entity-handle origin-entity-handle-bottom"
      />
    </div>
  );
};
