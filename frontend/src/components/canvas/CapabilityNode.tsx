import { Handle, Position } from '@xyflow/react';
import React from 'react';
import { DEFAULT_CUSTOM_COLOR, deriveMaturityValue } from '../../constants/maturityColors';
import { useCurrentView } from '../../features/views/hooks/useCurrentView';
import { useMaturityColorScale } from '../../hooks/useMaturityColorScale';
import classes from './CapabilityNode.module.css';
import { getContrastTextColor } from './contrastText';
import handles from './nodeHandles.module.css';

type ColorScheme = 'maturity' | 'classic' | 'custom';

export interface CapabilityNodeData {
  label: string;
  level: string;
  maturityLevel?: string;
  maturityValue?: number;
  maturitySection?: string;
  isSelected: boolean;
  customColor?: string;
}

const SCHEME_CLASS_NAMES: Partial<Record<ColorScheme, string>> = {
  classic: classes.classic,
};

const hasValidCustomColor = (customColor: string | undefined): boolean => {
  return customColor !== undefined && customColor.trim() !== '';
};

const resolveCustomFill = (customColor: string | undefined): string => {
  return hasValidCustomColor(customColor) ? (customColor as string) : DEFAULT_CUSTOM_COLOR;
};

export const CapabilityNode: React.FC<{ data: CapabilityNodeData; id: string; selected?: boolean }> = ({
  data,
  id,
  selected,
}) => {
  const { currentView } = useCurrentView();
  const colorScheme = (currentView?.colorScheme || 'maturity') as ColorScheme;
  const { getColorForValue, getSectionNameForValue } = useMaturityColorScale();
  const isSelected = data.isSelected || !!selected;

  const effectiveMaturityValue = data.maturityValue ?? deriveMaturityValue(data.maturityLevel);
  const sectionName = getSectionNameForValue(effectiveMaturityValue);

  const dataDrivenFill =
    colorScheme === 'custom'
      ? resolveCustomFill(data.customColor)
      : colorScheme === 'maturity'
        ? getColorForValue(effectiveMaturityValue)
        : undefined;

  const nodeClassName = [classes.node, SCHEME_CLASS_NAMES[colorScheme], isSelected ? classes.selected : '']
    .filter(Boolean)
    .join(' ');

  return (
    <div
      className={nodeClassName}
      style={
        dataDrivenFill ? { backgroundColor: dataDrivenFill, color: getContrastTextColor(dataDrivenFill) } : undefined
      }
      data-capability-id={id}
      data-testid="capability-node"
    >
      <Handle type="source" position={Position.Top} id="top" className={handles.top} />
      <Handle type="target" position={Position.Top} id="top" className={handles.top} />

      <Handle type="source" position={Position.Left} id="left" className={handles.left} />
      <Handle type="target" position={Position.Left} id="left" className={handles.left} />

      <div className={classes.content}>
        <div className={classes.header}>
          <svg
            className={classes.icon}
            width="16"
            height="16"
            viewBox="0 0 16 16"
            fill="none"
            xmlns="http://www.w3.org/2000/svg"
          >
            <rect x="1" y="1" width="6" height="6" rx="1" stroke="currentColor" strokeWidth="1.5" fill="none" />
            <rect x="9" y="1" width="6" height="6" rx="1" stroke="currentColor" strokeWidth="1.5" fill="none" />
            <rect x="1" y="9" width="6" height="6" rx="1" stroke="currentColor" strokeWidth="1.5" fill="none" />
            <rect x="9" y="9" width="6" height="6" rx="1" stroke="currentColor" strokeWidth="1.5" fill="none" />
          </svg>
          <span className={classes.level} data-testid="capability-node-level">
            {data.level}:
          </span>
          <span className={classes.name}>{data.label}</span>
        </div>
        <div className={classes.maturitySection} data-testid="capability-node-maturity">
          {sectionName}
        </div>
      </div>

      <Handle type="source" position={Position.Right} id="right" className={handles.right} />
      <Handle type="target" position={Position.Right} id="right" className={handles.right} />

      <Handle type="source" position={Position.Bottom} id="bottom" className={handles.bottom} />
      <Handle type="target" position={Position.Bottom} id="bottom" className={handles.bottom} />
    </div>
  );
};
