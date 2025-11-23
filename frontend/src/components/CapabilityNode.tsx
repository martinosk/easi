import React from 'react';
import { Handle, Position } from '@xyflow/react';

export interface CapabilityNodeData {
  label: string;
  level: string;
  maturityLevel?: string;
  isSelected: boolean;
}

const getMaturityColor = (maturityLevel?: string): string => {
  switch (maturityLevel?.toLowerCase()) {
    case 'genesis':
      return '#ef4444';
    case 'custom build':
      return '#f97316';
    case 'product':
      return '#22c55e';
    case 'commodity':
      return '#3b82f6';
    default:
      return '#6b7280';
  }
};

const getMaturityBackgroundGradient = (maturityLevel?: string): string => {
  const baseColor = getMaturityColor(maturityLevel);
  return `linear-gradient(135deg, ${baseColor} 0%, ${baseColor}dd 100%)`;
};

export const CapabilityNode: React.FC<{ data: CapabilityNodeData; id: string }> = ({ data, id }) => {
  const maturityColor = getMaturityColor(data.maturityLevel);

  return (
    <div
      className={`capability-node ${data.isSelected ? 'capability-node-selected' : ''}`}
      style={{
        background: getMaturityBackgroundGradient(data.maturityLevel),
        borderColor: data.isSelected ? '#374151' : maturityColor,
      }}
      data-capability-id={id}
    >
      <Handle
        type="source"
        position={Position.Top}
        id="top"
        className="capability-handle capability-handle-top"
      />
      <Handle
        type="target"
        position={Position.Top}
        id="top"
        className="capability-handle capability-handle-top"
      />

      <Handle
        type="source"
        position={Position.Left}
        id="left"
        className="capability-handle capability-handle-left"
      />
      <Handle
        type="target"
        position={Position.Left}
        id="left"
        className="capability-handle capability-handle-left"
      />

      <div className="capability-node-content">
        <div className="capability-node-header">
          <span className="capability-node-icon">◆</span>
          <span className="capability-node-level">{data.level}:</span>
          <span className="capability-node-name">{data.label}</span>
        </div>
        <div className="capability-node-maturity">
          {data.maturityLevel || 'Unknown'}
        </div>
      </div>

      <Handle
        type="source"
        position={Position.Right}
        id="right"
        className="capability-handle capability-handle-right"
      />
      <Handle
        type="target"
        position={Position.Right}
        id="right"
        className="capability-handle capability-handle-right"
      />

      <Handle
        type="source"
        position={Position.Bottom}
        id="bottom"
        className="capability-handle capability-handle-bottom"
      />
      <Handle
        type="target"
        position={Position.Bottom}
        id="bottom"
        className="capability-handle capability-handle-bottom"
      />
    </div>
  );
};
