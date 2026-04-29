import { Position } from '@xyflow/react';
import React from 'react';
import { useCurrentView } from '../../features/views/hooks/useCurrentView';
import { ConnectorHandle, type ConnectorClickInfo } from './ConnectorHandle';

type HexColor = string;
type ColorScheme = 'maturity' | 'classic' | 'custom';

export interface ComponentNodeData {
  label: string;
  description?: string;
  isSelected: boolean;
  customColor?: string;
  onConnectorClick?: (info: ConnectorClickInfo) => void;
}

const COMPONENT_COLORS: Record<ColorScheme, HexColor> = {
  maturity: '#3b82f6',
  classic: '#bfd9f0',
  custom: '#3b82f6',
};

const DEFAULT_CUSTOM_COLOR: HexColor = '#E0E0E0';
const SELECTED_BORDER_COLOR: HexColor = '#374151';

const getColorByScheme = (colorScheme: ColorScheme): HexColor => {
  return COMPONENT_COLORS[colorScheme];
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
}

const resolveBaseColor = (config: ColorConfig): HexColor => {
  if (config.colorScheme === 'custom') {
    return hasValidCustomColor(config.customColor) ? config.customColor! : DEFAULT_CUSTOM_COLOR;
  }
  return getColorByScheme(config.colorScheme);
};

const resolveBorderColor = (isSelected: boolean, baseColor: HexColor): HexColor => {
  return isSelected ? SELECTED_BORDER_COLOR : baseColor;
};

export const ComponentNode: React.FC<{ data: ComponentNodeData; id: string; selected?: boolean }> = ({
  data,
  id,
  selected,
}) => {
  const { currentView } = useCurrentView();
  const colorScheme = (currentView?.colorScheme || 'maturity') as ColorScheme;
  const isSelected = data.isSelected || !!selected;

  const baseColor = resolveBaseColor({ colorScheme, customColor: data.customColor });
  const borderColor = resolveBorderColor(isSelected, baseColor);

  const nodeClassName = `component-node ${isSelected ? 'component-node-selected' : ''} ${colorScheme === 'classic' ? 'classic-text' : ''}`;

  return (
    <div
      className={nodeClassName}
      style={{
        background: getBackgroundGradient(baseColor),
        borderColor: borderColor,
      }}
      data-component-id={id}
    >
      <ConnectorHandle type="source" position={Position.Top} id="top" className="component-handle component-handle-top" nodeId={id} onConnectorClick={data.onConnectorClick} />
      <ConnectorHandle type="target" position={Position.Top} id="top" className="component-handle component-handle-top" nodeId={id} onConnectorClick={data.onConnectorClick} />

      <ConnectorHandle type="source" position={Position.Left} id="left" className="component-handle component-handle-left" nodeId={id} onConnectorClick={data.onConnectorClick} />
      <ConnectorHandle type="target" position={Position.Left} id="left" className="component-handle component-handle-left" nodeId={id} onConnectorClick={data.onConnectorClick} />

      <div className="component-node-content">
        <div className="component-node-header">{data.label}</div>
        {data.description && <div className="component-node-description">{data.description}</div>}
      </div>

      <ConnectorHandle type="source" position={Position.Right} id="right" className="component-handle component-handle-right" nodeId={id} onConnectorClick={data.onConnectorClick} />
      <ConnectorHandle type="target" position={Position.Right} id="right" className="component-handle component-handle-right" nodeId={id} onConnectorClick={data.onConnectorClick} />

      <ConnectorHandle
        type="source"
        position={Position.Bottom}
        id="bottom"
        className="component-handle component-handle-bottom"
        nodeId={id}
        onConnectorClick={data.onConnectorClick}
      />
      <ConnectorHandle
        type="target"
        position={Position.Bottom}
        id="bottom"
        className="component-handle component-handle-bottom"
        nodeId={id}
        onConnectorClick={data.onConnectorClick}
      />
    </div>
  );
};
