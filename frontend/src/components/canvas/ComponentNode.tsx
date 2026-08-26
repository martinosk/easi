import { Handle, Position } from '@xyflow/react';
import React from 'react';
import { DEFAULT_CUSTOM_COLOR } from '../../constants/maturityColors';
import { useCurrentView } from '../../features/views/hooks/useCurrentView';
import classes from './ComponentNode.module.css';
import { getContrastTextColor } from './contrastText';
import handles from './nodeHandles.module.css';

type ColorScheme = 'maturity' | 'classic' | 'custom';

export interface ComponentNodeData {
  label: string;
  description?: string;
  isSelected: boolean;
  customColor?: string;
}

const SCHEME_CLASS_NAMES: Partial<Record<ColorScheme, string>> = {
  maturity: classes.maturity,
  classic: classes.classic,
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

  const nodeClassName = [classes.node, SCHEME_CLASS_NAMES[colorScheme], isSelected ? classes.selected : '']
    .filter(Boolean)
    .join(' ');

  return (
    <div
      className={nodeClassName}
      style={customFill ? { backgroundColor: customFill, color: getContrastTextColor(customFill) } : undefined}
      data-component-id={id}
      data-testid="component-node"
    >
      <Handle type="source" position={Position.Top} id="top" className={handles.top} />
      <Handle type="target" position={Position.Top} id="top" className={handles.top} />

      <Handle type="source" position={Position.Left} id="left" className={handles.left} />
      <Handle type="target" position={Position.Left} id="left" className={handles.left} />

      <div>
        <div className={classes.header} data-testid="component-node-header">
          {data.label}
        </div>
        {data.description && <div className={classes.description}>{data.description}</div>}
      </div>

      <Handle
        type="source"
        position={Position.Right}
        id="right"
        className={handles.right}
        data-testid="component-handle-right"
      />
      <Handle
        type="target"
        position={Position.Right}
        id="right"
        className={handles.right}
        data-testid="component-handle-right"
      />

      <Handle type="source" position={Position.Bottom} id="bottom" className={handles.bottom} />
      <Handle type="target" position={Position.Bottom} id="bottom" className={handles.bottom} />
    </div>
  );
};
