import React from 'react';
import { Handle, Position } from '@xyflow/react';
import { useCurrentView } from '../../hooks/useCurrentView';

type HexColor = string;
type ColorScheme = 'maturity' | 'classic' | 'custom';
type MaturityLevel = 'genesis' | 'custom build' | 'product' | 'commodity';

export interface CapabilityNodeData {
  label: string;
  level: string;
  maturityLevel?: string;
  isSelected: boolean;
  customColor?: string;
}

const MATURITY_COLORS: Record<MaturityLevel, HexColor> = {
  'genesis': '#ef4444',
  'custom build': '#f97316',
  'product': '#22c55e',
  'commodity': '#3b82f6',
};

const DEFAULT_MATURITY_COLOR: HexColor = '#6b7280';
const CLASSIC_COLOR: HexColor = '#f9c268';
const DEFAULT_CUSTOM_COLOR: HexColor = '#E0E0E0';
const SELECTED_BORDER_COLOR: HexColor = '#374151';

const getMaturityColor = (maturityLevel?: string): HexColor => {
  const level = maturityLevel?.toLowerCase() as MaturityLevel | undefined;
  return level && level in MATURITY_COLORS ? MATURITY_COLORS[level] : DEFAULT_MATURITY_COLOR;
};

const getColorByScheme = (colorScheme: ColorScheme, maturityLevel?: string): HexColor => {
  if (colorScheme === 'classic') {
    return CLASSIC_COLOR;
  }
  return getMaturityColor(maturityLevel);
};

const getBackgroundGradient = (baseColor: HexColor): string => {
  return `linear-gradient(135deg, ${baseColor} 0%, ${baseColor}dd 100%)`;
};

const hasValidCustomColor = (customColor: HexColor | undefined): boolean => {
  return customColor !== undefined && customColor.trim() !== '';
};

interface ColorConfig {
  colorScheme: ColorScheme;
  customColor: HexColor | undefined;
  maturityLevel: string | undefined;
}

const resolveBaseColor = (config: ColorConfig): HexColor => {
  if (config.colorScheme === 'custom') {
    return hasValidCustomColor(config.customColor) ? config.customColor! : DEFAULT_CUSTOM_COLOR;
  }
  return getColorByScheme(config.colorScheme, config.maturityLevel);
};

const resolveBorderColor = (isSelected: boolean, baseColor: HexColor): HexColor => {
  return isSelected ? SELECTED_BORDER_COLOR : baseColor;
};

export const CapabilityNode: React.FC<{ data: CapabilityNodeData; id: string }> = ({ data, id }) => {
  const { currentView } = useCurrentView();
  const colorScheme = (currentView?.colorScheme || 'maturity') as ColorScheme;

  const baseColor = resolveBaseColor({ colorScheme, customColor: data.customColor, maturityLevel: data.maturityLevel });
  const borderColor = resolveBorderColor(data.isSelected, baseColor);

  const nodeClassName = `capability-node ${data.isSelected ? 'capability-node-selected' : ''} ${colorScheme === 'classic' ? 'classic-text' : ''}`;

  return (
    <div
      className={nodeClassName}
      style={{
        background: getBackgroundGradient(baseColor),
        borderColor: borderColor,
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
          <svg className="capability-node-icon" width="16" height="16" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
            <rect x="1" y="1" width="6" height="6" rx="1" stroke="currentColor" strokeWidth="1.5" fill="none"/>
            <rect x="9" y="1" width="6" height="6" rx="1" stroke="currentColor" strokeWidth="1.5" fill="none"/>
            <rect x="1" y="9" width="6" height="6" rx="1" stroke="currentColor" strokeWidth="1.5" fill="none"/>
            <rect x="9" y="9" width="6" height="6" rx="1" stroke="currentColor" strokeWidth="1.5" fill="none"/>
          </svg>
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
