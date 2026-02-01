import React from 'react';
import { Handle, Position } from '@xyflow/react';
import { useCurrentView } from '../../features/views/hooks/useCurrentView';
import { useMaturityColorScale } from '../../hooks/useMaturityColorScale';
import { CLASSIC_COLOR, DEFAULT_CUSTOM_COLOR, SELECTED_BORDER_COLOR, deriveMaturityValue } from '../../constants/maturityColors';

type HexColor = string;
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

const getColorByScheme = (
  colorScheme: ColorScheme,
  maturityValue: number | undefined,
  maturityLevel: string | undefined,
  getColorForValue: (value: number) => string
): HexColor => {
  if (colorScheme === 'classic') {
    return CLASSIC_COLOR;
  }

  const effectiveValue = maturityValue ?? deriveMaturityValue(maturityLevel);
  return getColorForValue(effectiveValue);
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
  maturityValue: number | undefined;
  getColorForValue: (value: number) => string;
}

const resolveBaseColor = (config: ColorConfig): HexColor => {
  if (config.colorScheme === 'custom') {
    return hasValidCustomColor(config.customColor) ? config.customColor! : DEFAULT_CUSTOM_COLOR;
  }
  return getColorByScheme(config.colorScheme, config.maturityValue, config.maturityLevel, config.getColorForValue);
};

const resolveBorderColor = (isSelected: boolean, baseColor: HexColor): HexColor => {
  return isSelected ? SELECTED_BORDER_COLOR : baseColor;
};

export const CapabilityNode: React.FC<{ data: CapabilityNodeData; id: string; selected?: boolean }> = ({ data, id, selected }) => {
  const { currentView } = useCurrentView();
  const colorScheme = (currentView?.colorScheme || 'maturity') as ColorScheme;
  const { getColorForValue, getSectionNameForValue } = useMaturityColorScale();
  const isSelected = data.isSelected || !!selected;

  const effectiveMaturityValue = data.maturityValue ?? deriveMaturityValue(data.maturityLevel);
  const sectionName = getSectionNameForValue(effectiveMaturityValue);

  const baseColor = resolveBaseColor({
    colorScheme,
    customColor: data.customColor,
    maturityLevel: data.maturityLevel,
    maturityValue: data.maturityValue,
    getColorForValue,
  });
  const borderColor = resolveBorderColor(isSelected, baseColor);

  const nodeClassName = `capability-node ${isSelected ? 'capability-node-selected' : ''} ${colorScheme === 'classic' ? 'classic-text' : ''}`;

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
          {sectionName}
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
