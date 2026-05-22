import { Box } from '@mantine/core';
import type { Capability, CapabilityId, CapabilityRealization, ComponentId } from '../../../../api/types';
import type { Breakpoint } from '../../../../hooks/useResponsive';
import { getResponsiveValue, RESPONSIVE_GRID_COLUMNS, RESPONSIVE_SPACING } from '../../../../hooks/useResponsive';
import { ApplicationChipList } from '../ApplicationChipList';
import type { DepthLevel } from '../DepthSelector';
import classes from './CapabilityItem.module.css';
import { type CapabilityNode, levelToNumber } from './gridUtils';

const LEVEL_CLASS: Record<Capability['level'], string> = {
  L1: classes.levelL1,
  L2: classes.levelL2,
  L3: classes.levelL3,
  L4: classes.levelL4,
};

function getResponsiveGridColumns(level: Capability['level'], breakpoint: Breakpoint): string {
  const columns = RESPONSIVE_GRID_COLUMNS[level] || RESPONSIVE_GRID_COLUMNS.L3;
  return getResponsiveValue(columns, breakpoint) || columns.base;
}

function getVisibleChildren(children: CapabilityNode[], depth: DepthLevel): CapabilityNode[] {
  return children.filter((child) => levelToNumber(child.capability.level) <= depth);
}

function getCapabilityRealizations(
  showApplications: boolean,
  getRealizationsForCapability: ((id: CapabilityId) => CapabilityRealization[]) | undefined,
  capabilityId: CapabilityId,
): CapabilityRealization[] {
  if (!showApplications || !getRealizationsForCapability) return [];
  return getRealizationsForCapability(capabilityId);
}

export interface CapabilityItemProps {
  node: CapabilityNode;
  depth: DepthLevel;
  onClick: (capability: Capability, event: React.MouseEvent) => void;
  onContextMenu?: (capability: Capability, event: React.MouseEvent) => void;
  showApplications?: boolean;
  getRealizationsForCapability?: (capabilityId: CapabilityId) => CapabilityRealization[];
  onApplicationClick?: (componentId: ComponentId) => void;
  selectedCapabilities?: Set<CapabilityId>;
  breakpoint: Breakpoint;
}

interface ChildrenGridProps {
  nodes: CapabilityNode[];
  level: Capability['level'];
  depth: DepthLevel;
  onClick: (capability: Capability, event: React.MouseEvent) => void;
  onContextMenu?: (capability: Capability, event: React.MouseEvent) => void;
  showApplications: boolean;
  getRealizationsForCapability?: (capabilityId: CapabilityId) => CapabilityRealization[];
  onApplicationClick?: (componentId: ComponentId) => void;
  selectedCapabilities?: Set<CapabilityId>;
  breakpoint: Breakpoint;
}

function ChildrenGrid({
  nodes,
  level,
  depth,
  onClick,
  onContextMenu,
  showApplications,
  getRealizationsForCapability,
  onApplicationClick,
  selectedCapabilities,
  breakpoint,
}: ChildrenGridProps) {
  const gridGap = getResponsiveValue(RESPONSIVE_SPACING.gridGap, breakpoint) || '0.5rem';

  return (
    <div
      className={classes.childrenGrid}
      style={{
        gridTemplateColumns: getResponsiveGridColumns(level, breakpoint),
        gap: gridGap,
      }}
    >
      {nodes.map((child) => (
        <CapabilityItem
          key={child.capability.id}
          node={child}
          depth={depth}
          onClick={onClick}
          onContextMenu={onContextMenu}
          showApplications={showApplications}
          getRealizationsForCapability={getRealizationsForCapability}
          onApplicationClick={onApplicationClick}
          selectedCapabilities={selectedCapabilities}
          breakpoint={breakpoint}
        />
      ))}
    </div>
  );
}

function buildHandlers(
  capability: Capability,
  onClick: CapabilityItemProps['onClick'],
  onContextMenu: CapabilityItemProps['onContextMenu'],
) {
  return {
    onClick: (e: React.MouseEvent) => {
      e.stopPropagation();
      onClick(capability, e);
    },
    onContextMenu: (e: React.MouseEvent) => {
      e.preventDefault();
      e.stopPropagation();
      if (capability.level === 'L1') {
        onContextMenu?.(capability, e);
      }
    },
  };
}

function tileClassName(level: Capability['level'], isSelected: boolean): string {
  return [classes.tile, LEVEL_CLASS[level], isSelected ? classes.selected : ''].filter(Boolean).join(' ');
}

interface CapabilityItemBodyProps {
  name: string;
  hasContent: boolean;
  realizations: CapabilityRealization[];
  onApplicationClick?: (componentId: ComponentId) => void;
  childrenSection: React.ReactNode;
}

function CapabilityItemBody({
  name,
  hasContent,
  realizations,
  onApplicationClick,
  childrenSection,
}: CapabilityItemBodyProps) {
  return (
    <>
      <Box className={classes.title} mb={hasContent ? 'xs' : 0}>
        {name}
      </Box>
      {realizations.length > 0 && onApplicationClick && (
        <Box mb={childrenSection ? 'xs' : 0}>
          <ApplicationChipList realizations={realizations} onApplicationClick={onApplicationClick} />
        </Box>
      )}
      {childrenSection}
    </>
  );
}

export function CapabilityItem({
  node,
  depth,
  onClick,
  onContextMenu,
  showApplications = false,
  getRealizationsForCapability,
  onApplicationClick,
  selectedCapabilities,
  breakpoint,
}: CapabilityItemProps) {
  const { capability, children } = node;
  const realizations = getCapabilityRealizations(showApplications, getRealizationsForCapability, capability.id);
  const visibleChildren = getVisibleChildren(children, depth);
  const hasContent = visibleChildren.length > 0 || realizations.length > 0;
  const isSelected = selectedCapabilities?.has(capability.id) ?? false;
  const handlers = buildHandlers(capability, onClick, onContextMenu);

  const childrenSection =
    visibleChildren.length > 0 ? (
      <ChildrenGrid
        nodes={visibleChildren}
        level={capability.level}
        depth={depth}
        onClick={onClick}
        onContextMenu={onContextMenu}
        showApplications={showApplications}
        getRealizationsForCapability={getRealizationsForCapability}
        onApplicationClick={onApplicationClick}
        selectedCapabilities={selectedCapabilities}
        breakpoint={breakpoint}
      />
    ) : null;

  return (
    <Box
      className={tileClassName(capability.level, isSelected)}
      data-testid={`capability-${capability.id}`}
      data-level={capability.level}
      data-selected={isSelected || undefined}
      {...handlers}
    >
      <CapabilityItemBody
        name={capability.name}
        hasContent={hasContent}
        realizations={realizations}
        onApplicationClick={onApplicationClick}
        childrenSection={childrenSection}
      />
    </Box>
  );
}
