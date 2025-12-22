import React from 'react';
import { Handle, Position } from '@xyflow/react';
import { useCurrentView } from '../../hooks/useCurrentView';

type HexColor = string;
type ColorScheme = 'maturity' | 'classic' | 'custom';

export interface ComponentNodeData {
  label: string;
  description?: string;
  isSelected: boolean;
  customColor?: string;
}

const COMPONENT_COLORS: Record<ColorScheme, HexColor> = {
  'maturity': '#3b82f6',
  'classic': '#bfd9f0',
  'custom': '#3b82f6',
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

export const ComponentNode: React.FC<{ data: ComponentNodeData; id: string }> = ({ data, id }) => {
  const { currentView } = useCurrentView();
  const colorScheme = (currentView?.colorScheme || 'maturity') as ColorScheme;

  const baseColor = resolveBaseColor({ colorScheme, customColor: data.customColor });
  const borderColor = resolveBorderColor(data.isSelected, baseColor);

  const nodeClassName = `component-node ${data.isSelected ? 'component-node-selected' : ''} ${colorScheme === 'classic' ? 'classic-text' : ''}`;

  return (
    <div
      className={nodeClassName}
      style={{
        background: getBackgroundGradient(baseColor),
        borderColor: borderColor,
      }}
      data-component-id={id}
    >
      <Handle
        type="source"
        position={Position.Top}
        id="top"
        className="component-handle component-handle-top"
      />
      <Handle
        type="target"
        position={Position.Top}
        id="top"
        className="component-handle component-handle-top"
      />

      <Handle
        type="source"
        position={Position.Left}
        id="left"
        className="component-handle component-handle-left"
      />
      <Handle
        type="target"
        position={Position.Left}
        id="left"
        className="component-handle component-handle-left"
      />

      <div className="component-node-content">
        <div className="component-node-header">{data.label}</div>
        {data.description && (
          <div className="component-node-description">{data.description}</div>
        )}
      </div>

      <Handle
        type="source"
        position={Position.Right}
        id="right"
        className="component-handle component-handle-right"
      />
      <Handle
        type="target"
        position={Position.Right}
        id="right"
        className="component-handle component-handle-right"
      />

      <Handle
        type="source"
        position={Position.Bottom}
        id="bottom"
        className="component-handle component-handle-bottom"
      />
      <Handle
        type="target"
        position={Position.Bottom}
        id="bottom"
        className="component-handle component-handle-bottom"
      />
    </div>
  );
};
