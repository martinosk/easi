import { Group, UnstyledButton } from '@mantine/core';
import { IconChevronRight } from '@tabler/icons-react';
import { useState } from 'react';
import type { Capability, CapabilityId, CapabilityRealization, ComponentId } from '../../../api/types';
import type { CapabilityTreeNode } from '../../capabilities/hooks/useCapabilityTree';
import { nodeMatchesSearch } from '../hooks/boardSearch';
import { AppChip } from './AppChip';
import { BoardCapabilityCard } from './BoardCapabilityCard';
import classes from './L1Group.module.css';

interface OwnAppsRowProps {
  capabilityId: CapabilityId;
  realizations: CapabilityRealization[];
  onChipClick: (componentId: ComponentId) => void;
}

function OwnAppsRow({ capabilityId, realizations, onChipClick }: OwnAppsRowProps) {
  if (realizations.length === 0) return null;

  return (
    <Group gap="xs" wrap="wrap" className={classes.ownApps} data-testid={`l1-own-apps-${capabilityId}`}>
      {realizations.map((realization) => (
        <AppChip key={realization.id} realization={realization} onClick={onChipClick} />
      ))}
    </Group>
  );
}

interface L1GroupBodyProps {
  node: CapabilityTreeNode;
  ownRealizations: CapabilityRealization[];
  visibleChildren: CapabilityTreeNode[];
  selectedCapabilities: Set<CapabilityId>;
  getColorForValue: (maturityValue: number) => string;
  getRealizationsForCapability: (capabilityId: CapabilityId) => CapabilityRealization[];
  onCapabilityClick: (capability: Capability, event: React.MouseEvent) => void;
  onCapabilityContextMenu: (capability: Capability, event: React.MouseEvent) => void;
  onChipClick: (componentId: ComponentId) => void;
}

function L1GroupBody({
  node,
  ownRealizations,
  visibleChildren,
  selectedCapabilities,
  getColorForValue,
  getRealizationsForCapability,
  onCapabilityClick,
  onCapabilityContextMenu,
  onChipClick,
}: L1GroupBodyProps) {
  const cardProps = (child: CapabilityTreeNode) => ({
    node: child,
    isSelected: selectedCapabilities.has(child.capability.id),
    getColorForValue,
    getRealizationsForCapability,
    onClick: onCapabilityClick,
    onContextMenu: onCapabilityContextMenu,
    onChipClick,
  });

  return (
    <div className={classes.body}>
      {node.children.length > 0 && (
        <OwnAppsRow capabilityId={node.capability.id} realizations={ownRealizations} onChipClick={onChipClick} />
      )}
      {node.children.length === 0 ? (
        <BoardCapabilityCard {...cardProps(node)} />
      ) : (
        visibleChildren.map((child) => <BoardCapabilityCard key={child.capability.id} {...cardProps(child)} />)
      )}
    </div>
  );
}

export interface L1GroupProps {
  node: CapabilityTreeNode;
  distinctAppCount: number;
  searchQuery: string;
  forceOpen?: boolean;
  selectedCapabilities: Set<CapabilityId>;
  getColorForValue: (maturityValue: number) => string;
  getRealizationsForCapability: (capabilityId: CapabilityId) => CapabilityRealization[];
  onCapabilityClick: (capability: Capability, event: React.MouseEvent) => void;
  onCapabilityContextMenu: (capability: Capability, event: React.MouseEvent) => void;
  onChipClick: (componentId: ComponentId) => void;
}

export function L1Group({
  node,
  distinctAppCount,
  searchQuery,
  forceOpen = false,
  selectedCapabilities,
  getColorForValue,
  getRealizationsForCapability,
  onCapabilityClick,
  onCapabilityContextMenu,
  onChipClick,
}: L1GroupProps) {
  const [manualOpen, setManualOpen] = useState(false);
  const isOpen = manualOpen || Boolean(searchQuery) || forceOpen;
  const ownRealizations = getRealizationsForCapability(node.capability.id);

  const visibleChildren = searchQuery
    ? node.children.filter((child) => nodeMatchesSearch(child, searchQuery, getRealizationsForCapability))
    : node.children;

  return (
    <div data-testid={`l1-group-${node.capability.id}`}>
      <UnstyledButton
        component="button"
        className={classes.header}
        aria-expanded={isOpen}
        onClick={() => setManualOpen((open) => !open)}
      >
        <IconChevronRight size={12} className={[classes.chevron, isOpen ? classes.chevronOpen : ''].join(' ')} />
        {node.capability.name}
        <span className={classes.subTag}>L1 · {node.children.length} sub</span>
        <span className={[classes.appCount, distinctAppCount > 3 ? classes.appCountMulti : ''].join(' ')}>
          {distinctAppCount} apps
        </span>
      </UnstyledButton>
      {isOpen && (
        <L1GroupBody
          node={node}
          ownRealizations={ownRealizations}
          visibleChildren={visibleChildren}
          selectedCapabilities={selectedCapabilities}
          getColorForValue={getColorForValue}
          getRealizationsForCapability={getRealizationsForCapability}
          onCapabilityClick={onCapabilityClick}
          onCapabilityContextMenu={onCapabilityContextMenu}
          onChipClick={onChipClick}
        />
      )}
    </div>
  );
}
