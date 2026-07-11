import { Handle, Position } from '@xyflow/react';
import React from 'react';
import { DEFAULT_CUSTOM_COLOR, deriveMaturityValue } from '../../constants/maturityColors';
import { useCurrentView } from '../../features/views/hooks/useCurrentView';
import { useMaturityColorScale } from '../../hooks/useMaturityColorScale';
import { getContrastTextColor } from './contrastText';

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

const SCHEME_CLASS_NAMES: Record<ColorScheme, string> = {
  maturity: 'capability-node--maturity',
  classic: 'capability-node--classic classic-text',
  custom: 'capability-node--custom',
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

  const nodeClassName = [
    'capability-node',
    SCHEME_CLASS_NAMES[colorScheme],
    isSelected ? 'capability-node-selected' : '',
  ]
    .filter(Boolean)
    .join(' ');

  return (
    <div
      className={nodeClassName}
      style={
        dataDrivenFill ? { backgroundColor: dataDrivenFill, color: getContrastTextColor(dataDrivenFill) } : undefined
      }
      data-capability-id={id}
    >
      <Handle type="source" position={Position.Top} id="top" className="capability-handle capability-handle-top" />
      <Handle type="target" position={Position.Top} id="top" className="capability-handle capability-handle-top" />

      <Handle type="source" position={Position.Left} id="left" className="capability-handle capability-handle-left" />
      <Handle type="target" position={Position.Left} id="left" className="capability-handle capability-handle-left" />

      <div className="capability-node-content">
        <div className="capability-node-header">
          <svg
            className="capability-node-icon"
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
          <span className="capability-node-level">{data.level}:</span>
          <span className="capability-node-name">{data.label}</span>
        </div>
        <div className="capability-node-maturity">{sectionName}</div>
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
