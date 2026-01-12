import React, { useMemo } from 'react';
import type { Capability, View, ViewCapability } from '../../../../api/types';
import { TreeSection } from '../TreeSection';
import { buildCapabilityTree, getLevelNumber, hasCustomColor } from '../../utils/treeUtils';
import { deriveMaturityValue } from '../../../../constants/maturityColors';
import { useMaturityColorScale } from '../../../../hooks/useMaturityColorScale';
import type { CapabilityTreeNode } from '../../types';

interface ColorIndicatorProps {
  customColor: string | undefined;
}

const ColorIndicator: React.FC<ColorIndicatorProps> = ({ customColor }) => (
  <div
    data-testid="custom-color-indicator"
    style={{
      width: '10px',
      height: '10px',
      borderRadius: '2px',
      backgroundColor: customColor,
      display: 'inline-block',
      marginLeft: '8px',
      border: '1px solid rgba(0,0,0,0.1)',
    }}
  />
);

interface ExpandButtonProps {
  hasChildren: boolean;
  isExpanded: boolean;
  onClick: (e: React.MouseEvent) => void;
}

const ExpandButton: React.FC<ExpandButtonProps> = ({ hasChildren, isExpanded, onClick }) => {
  if (!hasChildren) {
    return <span className="capability-expand-placeholder" />;
  }
  return (
    <button className="capability-expand-btn" onClick={onClick}>
      {isExpanded ? '\u25BC' : '\u25B6'}
    </button>
  );
};

interface CapabilitiesSectionProps {
  capabilities: Capability[];
  currentView: View | null;
  isExpanded: boolean;
  onToggle: () => void;
  onAddCapability?: () => void;
  onCapabilitySelect?: (capabilityId: string) => void;
  onCapabilityContextMenu: (e: React.MouseEvent, capability: Capability) => void;
  expandedCapabilities: Set<string>;
  toggleCapabilityExpanded: (capabilityId: string) => void;
  selectedCapabilityId: string | null;
  setSelectedCapabilityId: (id: string | null) => void;
}

export const CapabilitiesSection: React.FC<CapabilitiesSectionProps> = ({
  capabilities,
  currentView,
  isExpanded,
  onToggle,
  onAddCapability,
  onCapabilitySelect,
  onCapabilityContextMenu,
  expandedCapabilities,
  toggleCapabilityExpanded,
  selectedCapabilityId,
  setSelectedCapabilityId,
}) => {
  const { getColorForValue, getSectionNameForValue } = useMaturityColorScale();
  const capabilityTree = useMemo(() => buildCapabilityTree(capabilities), [capabilities]);

  const handleCapabilityClick = (capabilityId: string) => {
    setSelectedCapabilityId(capabilityId);
    if (onCapabilitySelect) {
      onCapabilitySelect(capabilityId);
    }
  };

  const getCapabilityNodeData = (node: CapabilityTreeNode) => {
    const { capability } = node;
    const viewCapability = currentView?.capabilities.find((vc: ViewCapability) => vc.capabilityId === capability.id);
    const isOnCanvas = !!viewCapability;
    const customColor = viewCapability?.customColor;
    const colorScheme = currentView?.colorScheme ?? 'maturity';

    return {
      hasChildNodes: node.children.length > 0,
      isExpanded: expandedCapabilities.has(capability.id),
      levelNum: getLevelNumber(capability.level),
      isSelected: selectedCapabilityId === capability.id,
      isOnCanvas,
      showColorIndicator: hasCustomColor(currentView?.colorScheme, customColor),
      title: isOnCanvas ? (capability.description || capability.name) : `${capability.description || capability.name} (not in view)`,
      customColor,
      colorScheme,
    };
  };

  const buildCapabilityItemClassName = (levelNum: number, isSelected: boolean, isOnCanvas: boolean): string => {
    return [
      'capability-tree-item',
      `capability-level-${levelNum}`,
      isSelected && 'selected',
      !isOnCanvas && 'not-in-view',
    ].filter(Boolean).join(' ');
  };

  const renderCapabilityNode = (node: CapabilityTreeNode): React.ReactNode => {
    const { capability, children } = node;
    const nodeData = getCapabilityNodeData(node);

    const effectiveMaturityValue = capability.maturityValue ?? deriveMaturityValue(capability.maturityLevel);
    const maturityColor = nodeData.colorScheme === 'classic' ? '#f9c268' : getColorForValue(effectiveMaturityValue);
    const sectionName = capability.maturitySection?.name || getSectionNameForValue(effectiveMaturityValue);
    const maturityTooltip = `${sectionName} (${effectiveMaturityValue})`;

    const handleExpandClick = (e: React.MouseEvent) => {
      e.stopPropagation();
      toggleCapabilityExpanded(capability.id);
    };

    const handleDragStart = (e: React.DragEvent) => {
      e.dataTransfer.setData('capabilityId', capability.id);
      e.dataTransfer.effectAllowed = 'copy';
    };

    return (
      <div key={capability.id}>
        <div
          className={buildCapabilityItemClassName(nodeData.levelNum, nodeData.isSelected, nodeData.isOnCanvas)}
          draggable
          onDragStart={handleDragStart}
          onClick={() => handleCapabilityClick(capability.id)}
          onContextMenu={(e) => onCapabilityContextMenu(e, capability)}
          title={nodeData.title}
        >
          <ExpandButton
            hasChildren={nodeData.hasChildNodes}
            isExpanded={nodeData.isExpanded}
            onClick={handleExpandClick}
          />
          <span className="capability-level-badge">{capability.level}:</span>
          <span className="capability-name">{capability.name}</span>
          <span
            className="capability-maturity-indicator"
            style={{ backgroundColor: maturityColor }}
            title={maturityTooltip}
          />
          {nodeData.showColorIndicator && <ColorIndicator customColor={nodeData.customColor} />}
        </div>
        {nodeData.hasChildNodes && nodeData.isExpanded && (
          <div className="capability-children">
            {children.map(renderCapabilityNode)}
          </div>
        )}
      </div>
    );
  };

  return (
    <TreeSection
      label="Capabilities"
      count={capabilities.length}
      isExpanded={isExpanded}
      onToggle={onToggle}
      onAdd={onAddCapability}
      addTitle="Create new capability"
      addTestId="create-capability-button"
    >
      <div className="tree-items">
        {capabilityTree.length === 0 ? (
          <div className="tree-item-empty">No capabilities</div>
        ) : (
          capabilityTree.map(renderCapabilityNode)
        )}
      </div>
    </TreeSection>
  );
};
