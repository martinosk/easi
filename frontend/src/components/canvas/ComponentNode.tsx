import { Handle, Position } from '@xyflow/react';
import React from 'react';
import { DEFAULT_CUSTOM_COLOR } from '../../constants/maturityColors';
import { useCurrentView } from '../../features/views/hooks/useCurrentView';
import { getContrastTextColor } from './contrastText';

type ColorScheme = 'maturity' | 'classic' | 'custom';

export interface ComponentNodeData {
  label: string;
  description?: string;
  isSelected: boolean;
  customColor?: string;
}

const SCHEME_CLASS_NAMES: Record<ColorScheme, string> = {
  maturity: 'component-node--maturity',
  classic: 'component-node--classic classic-text',
  custom: 'component-node--custom',
};

const hasValidCustomColor = (customColor: string | undefined): boolean => {
  return customColor !== undefined && customColor.trim() !== '';
};

const resolveCustomFill = (customColor: string | undefined): string => {
  return hasValidCustomColor(customColor) ? (customColor as string) : DEFAULT_CUSTOM_COLOR;
};

export const ComponentNode: React.FC<{ data: ComponentNodeData; id: string; selected?: boolean }> = ({
  data,
  id,
  selected,
}) => {
  const { currentView } = useCurrentView();
  const colorScheme = (currentView?.colorScheme || 'maturity') as ColorScheme;
  const isSelected = data.isSelected || !!selected;
  const isCustom = colorScheme === 'custom';
  const customFill = isCustom ? resolveCustomFill(data.customColor) : undefined;

  const nodeClassName = ['component-node', SCHEME_CLASS_NAMES[colorScheme], isSelected ? 'component-node-selected' : '']
    .filter(Boolean)
    .join(' ');

  return (
    <div
      className={nodeClassName}
      style={customFill ? { backgroundColor: customFill, color: getContrastTextColor(customFill) } : undefined}
      data-component-id={id}
    >
      <Handle type="source" position={Position.Top} id="top" className="component-handle component-handle-top" />
      <Handle type="target" position={Position.Top} id="top" className="component-handle component-handle-top" />

      <Handle type="source" position={Position.Left} id="left" className="component-handle component-handle-left" />
      <Handle type="target" position={Position.Left} id="left" className="component-handle component-handle-left" />

      <div className="component-node-content">
        <div className="component-node-header">{data.label}</div>
        {data.description && <div className="component-node-description">{data.description}</div>}
      </div>

      <Handle type="source" position={Position.Right} id="right" className="component-handle component-handle-right" />
      <Handle type="target" position={Position.Right} id="right" className="component-handle component-handle-right" />

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
