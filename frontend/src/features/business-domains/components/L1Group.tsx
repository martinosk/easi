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

interface L1HeaderProps {
  capability: Capability;
  subCount: number;
  distinctAppCount: number;
  isOpen: boolean;
  isSelected: boolean;
  onToggle: () => void;
  onOpen: (capability: Capability, event: React.MouseEvent) => void;
}

function L1Header({ capability, subCount, distinctAppCount, isOpen, isSelected, onToggle, onOpen }: L1HeaderProps) {
  return (
    <div className={classes.header}>
      <UnstyledButton
        component="button"
        className={classes.toggle}
        aria-expanded={isOpen}
        aria-label={`${isOpen ? 'Collapse' : 'Expand'} sub-capabilities`}
        onClick={onToggle}
        data-testid={`l1-toggle-${capability.id}`}
      >
        <IconChevronRight size={12} className={[classes.chevron, isOpen ? classes.chevronOpen : ''].join(' ')} />
      </UnstyledButton>
      <UnstyledButton
        component="button"
        className={[classes.label, isSelected ? classes.selected : ''].filter(Boolean).join(' ')}
        onClick={(e) => onOpen(capability, e)}
        data-testid={`l1-open-${capability.id}`}
        data-selected={isSelected || undefined}
      >
        <span className={classes.name}>{capability.name}</span>
        <span className={classes.subTag}>L1 · {subCount} sub</span>
        <span className={[classes.appCount, distinctAppCount > 3 ? classes.appCountMulti : ''].join(' ')}>
          {distinctAppCount} apps
        </span>
      </UnstyledButton>
    </div>
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
  return (
    <div className={classes.body}>
      <OwnAppsRow capabilityId={node.capability.id} realizations={ownRealizations} onChipClick={onChipClick} />
      {visibleChildren.map((child) => (
        <BoardCapabilityCard
          key={child.capability.id}
          node={child}
          isSelected={selectedCapabilities.has(child.capability.id)}
          getColorForValue={getColorForValue}
          getRealizationsForCapability={getRealizationsForCapability}
          onClick={onCapabilityClick}
          onContextMenu={onCapabilityContextMenu}
          onChipClick={onChipClick}
        />
      ))}
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

  if (node.children.length === 0) {
    return (
      <div data-testid={`l1-group-${node.capability.id}`}>
        <BoardCapabilityCard
          node={node}
          isSelected={selectedCapabilities.has(node.capability.id)}
          getColorForValue={getColorForValue}
          getRealizationsForCapability={getRealizationsForCapability}
          onClick={onCapabilityClick}
          onContextMenu={onCapabilityContextMenu}
          onChipClick={onChipClick}
        />
      </div>
    );
  }

  const visibleChildren = searchQuery
    ? node.children.filter((child) => nodeMatchesSearch(child, searchQuery, getRealizationsForCapability))
    : node.children;

  return (
    <div data-testid={`l1-group-${node.capability.id}`}>
      <L1Header
        capability={node.capability}
        subCount={node.children.length}
        distinctAppCount={distinctAppCount}
        isOpen={isOpen}
        isSelected={selectedCapabilities.has(node.capability.id)}
        onToggle={() => setManualOpen((open) => !open)}
        onOpen={onCapabilityClick}
      />
      {isOpen && (
        <L1GroupBody
          node={node}
          ownRealizations={getRealizationsForCapability(node.capability.id)}
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
