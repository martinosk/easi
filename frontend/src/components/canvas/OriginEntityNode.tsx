import React from 'react';
import { Handle, Position } from '@xyflow/react';
import type { OriginEntityType } from '../../constants/entityIdentifiers';

type HexColor = string;

export type { OriginEntityType };

export interface OriginEntityNodeData {
  label: string;
  entityType: OriginEntityType;
  isSelected: boolean;
  subtitle?: string;
}

const ENTITY_COLORS: Record<OriginEntityType, HexColor> = {
  acquired: '#8b5cf6',
  vendor: '#ec4899',
  team: '#14b8a6',
};

const ENTITY_ICONS: Record<OriginEntityType, React.ReactNode> = {
  acquired: (
    <svg width="16" height="16" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
      <rect x="2" y="4" width="12" height="10" rx="1" stroke="currentColor" strokeWidth="1.5" fill="none"/>
      <path d="M4 4V3C4 2.44772 4.44772 2 5 2H11C11.5523 2 12 2.44772 12 3V4" stroke="currentColor" strokeWidth="1.5"/>
      <line x1="5" y1="7" x2="11" y2="7" stroke="currentColor" strokeWidth="1.5"/>
      <line x1="5" y1="10" x2="9" y2="10" stroke="currentColor" strokeWidth="1.5"/>
    </svg>
  ),
  vendor: (
    <svg width="16" height="16" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
      <rect x="1" y="5" width="14" height="10" rx="1" stroke="currentColor" strokeWidth="1.5" fill="none"/>
      <path d="M1 5L8 1L15 5" stroke="currentColor" strokeWidth="1.5"/>
      <rect x="6" y="9" width="4" height="6" stroke="currentColor" strokeWidth="1.5" fill="none"/>
    </svg>
  ),
  team: (
    <svg width="16" height="16" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
      <circle cx="8" cy="5" r="3" stroke="currentColor" strokeWidth="1.5" fill="none"/>
      <path d="M2 14C2 11.2386 4.68629 9 8 9C11.3137 9 14 11.2386 14 14" stroke="currentColor" strokeWidth="1.5"/>
    </svg>
  ),
};

const ENTITY_LABELS: Record<OriginEntityType, string> = {
  acquired: 'Acquired Entity',
  vendor: 'Vendor',
  team: 'Internal Team',
};

const SELECTED_BORDER_COLOR: HexColor = '#374151';

const getBackgroundGradient = (baseColor: HexColor): string => {
  return `linear-gradient(135deg, ${baseColor} 0%, ${baseColor}dd 100%)`;
};

export const OriginEntityNode: React.FC<{ data: OriginEntityNodeData; id: string; selected?: boolean }> = ({ data, id, selected }) => {
  const baseColor = ENTITY_COLORS[data.entityType];
  const isSelected = data.isSelected || !!selected;
  const borderColor = isSelected ? SELECTED_BORDER_COLOR : baseColor;

  const nodeClassName = `origin-entity-node origin-entity-node-${data.entityType} ${isSelected ? 'origin-entity-node-selected' : ''}`;

  return (
    <div
      className={nodeClassName}
      style={{
        background: getBackgroundGradient(baseColor),
        borderColor: borderColor,
        borderWidth: isSelected ? 3 : 2,
        borderStyle: 'solid',
        borderRadius: '8px',
        padding: '12px 16px',
        minWidth: '150px',
        maxWidth: '220px',
        color: 'white',
        cursor: 'pointer',
      }}
      data-origin-entity-id={id}
    >
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

      <div className="origin-entity-node-content" style={{ display: 'flex', flexDirection: 'column', gap: '4px' }}>
        <div className="origin-entity-node-header" style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
          <span style={{ display: 'flex', alignItems: 'center' }}>
            {ENTITY_ICONS[data.entityType]}
          </span>
          <span style={{ fontSize: '14px', fontWeight: 600, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
            {data.label}
          </span>
        </div>
        <div className="origin-entity-node-type" style={{ fontSize: '11px', opacity: 0.9 }}>
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
