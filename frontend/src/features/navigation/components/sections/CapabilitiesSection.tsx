import { Tooltip } from '@mantine/core';
import React, { useCallback, useMemo, useState } from 'react';
import type { Capability, View, ViewCapability } from '../../../../api/types';
import { deriveMaturityValue } from '../../../../constants/maturityColors';
import { useMaturityColorScale } from '../../../../hooks/useMaturityColorScale';
import { CapabilityTree } from '../../../capabilities/components/CapabilityTree';
import type { CapabilityTreeNode } from '../../../capabilities/hooks/useCapabilityTree';
import { buildCapabilityTree } from '../../../capabilities/hooks/useCapabilityTree';
import { OnePagerIncompleteIndicator } from '../../../one-pagers/components/OnePagerIncompleteIndicator';
import type { TreeSelectedItem } from '../../hooks/useTreeMultiSelect';
import type { TreeMultiSelectProps } from '../../types';
import { hasCustomColor } from '../../utils/treeUtils';
import { TreeSection } from '../TreeSection';
import classes from './CapabilitiesSection.module.css';

interface MaturityDotProps {
  capability: Capability;
  colorScheme: string;
}

const MaturityDot: React.FC<MaturityDotProps> = ({ capability, colorScheme }) => {
  const { getColorForValue, getSectionNameForValue } = useMaturityColorScale();
  const value = capability.maturityValue ?? deriveMaturityValue(capability.maturityLevel);
  const sectionName = capability.maturitySection?.name || getSectionNameForValue(value);

  if (colorScheme === 'classic') return null;

  return (
    <Tooltip label={`Maturity: ${sectionName}`} withArrow>
      <span className={classes.maturityDot} style={{ backgroundColor: getColorForValue(value) }} />
    </Tooltip>
  );
};

interface CapabilitiesSectionProps {
  capabilities: Capability[];
  currentView: View | null;
  capabilitiesInView?: Set<string>;
  isExpanded: boolean;
  onToggle: () => void;
  onAddCapability?: () => void;
  onCapabilitySelect?: (capabilityId: string) => void;
  onCapabilityContextMenu: (e: React.MouseEvent, capability: Capability) => void;
  expandedCapabilities: Set<string>;
  toggleCapabilityExpanded: (capabilityId: string) => void;
  selectedCapabilityId: string | null;
  setSelectedCapabilityId: (id: string | null) => void;
  multiSelect: TreeMultiSelectProps;
}

function defaultCapabilitiesInView(currentView: View | null): Set<string> {
  return new Set((currentView?.capabilities ?? []).map((vc) => vc.capabilityId));
}

function toSelectedItem(capability: Capability): TreeSelectedItem {
  return { id: capability.id, name: capability.name, type: 'capability', links: capability._links };
}

export const CapabilitiesSection: React.FC<CapabilitiesSectionProps> = ({
  capabilities,
  currentView,
  capabilitiesInView,
  isExpanded,
  onToggle,
  onAddCapability,
  onCapabilitySelect,
  onCapabilityContextMenu,
  expandedCapabilities,
  toggleCapabilityExpanded,
  selectedCapabilityId,
  setSelectedCapabilityId,
  multiSelect,
}) => {
  const [visibleItems, setVisibleItems] = useState<TreeSelectedItem[]>([]);
  const effectiveCapabilitiesInView = useMemo(
    () => capabilitiesInView ?? defaultCapabilitiesInView(currentView),
    [capabilitiesInView, currentView],
  );
  const tree = useMemo(() => buildCapabilityTree(capabilities, { orphanRoots: 'any-level' }), [capabilities]);

  const handleVisibleNodesChange = useCallback((nodes: CapabilityTreeNode[]) => {
    setVisibleItems(nodes.map((node) => toSelectedItem(node.capability)));
  }, []);

  const handleCapabilityClick = (capability: Capability, event: React.MouseEvent) => {
    const result = multiSelect.handleItemClick(toSelectedItem(capability), 'capabilities', visibleItems, event);
    if (result === 'single') {
      setSelectedCapabilityId(capability.id);
      onCapabilitySelect?.(capability.id);
    }
  };

  const handleContextMenu = (e: React.MouseEvent, capability: Capability) => {
    const handled = multiSelect.handleContextMenu(e, capability.id, multiSelect.selectedItems);
    if (!handled) {
      onCapabilityContextMenu(e, capability);
    }
  };

  const handleDragStart = (e: React.DragEvent, capability: Capability) => {
    const handled = multiSelect.handleDragStart(e, capability.id);
    if (!handled) {
      e.dataTransfer.setData('capabilityId', capability.id);
      e.dataTransfer.effectAllowed = 'copy';
    }
  };

  const getRowProps = (node: CapabilityTreeNode) => {
    const { capability } = node;
    const isOnCanvas = effectiveCapabilitiesInView.has(capability.id);
    const baseTitle = capability.description || capability.name;

    return {
      draggable: true,
      selected: selectedCapabilityId === capability.id || multiSelect.isMultiSelected(capability.id),
      dimmed: !isOnCanvas,
      title: isOnCanvas ? baseTitle : `${baseTitle} (not in view)`,
      testId: `capability-tree-item-${capability.id}`,
      onClick: (e: React.MouseEvent) => handleCapabilityClick(capability, e),
      onContextMenu: (e: React.MouseEvent) => handleContextMenu(e, capability),
      onDragStart: (e: React.DragEvent) => handleDragStart(e, capability),
    };
  };

  const renderRight = (node: CapabilityTreeNode) => {
    const { capability } = node;
    const viewCapability = currentView?.capabilities.find((vc: ViewCapability) => vc.capabilityId === capability.id);
    const colorScheme = currentView?.colorScheme ?? 'maturity';

    return (
      <>
        <MaturityDot capability={capability} colorScheme={colorScheme} />
        <OnePagerIncompleteIndicator id={capability.id} onePagerComplete={capability.onePagerComplete} />
        {hasCustomColor(currentView?.colorScheme, viewCapability?.customColor) && (
          <Tooltip label="Custom colour in this view" withArrow>
            <span
              data-testid="custom-color-indicator"
              className={classes.colorSwatch}
              style={{ backgroundColor: viewCapability?.customColor }}
            />
          </Tooltip>
        )}
      </>
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
      <CapabilityTree
        tree={tree}
        expandedIds={expandedCapabilities}
        onToggleExpanded={toggleCapabilityExpanded}
        getRowProps={getRowProps}
        renderRight={renderRight}
        onVisibleNodesChange={handleVisibleNodesChange}
      />
    </TreeSection>
  );
};
