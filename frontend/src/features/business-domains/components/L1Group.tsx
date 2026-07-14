import { Group, UnstyledButton } from '@mantine/core';
import { IconChevronRight } from '@tabler/icons-react';
import { useState } from 'react';
import type { Capability, CapabilityId, ComponentId } from '../../../api/types';
import type { CapabilityTreeNode } from '../../capabilities/hooks/useCapabilityTree';
import { nodeMatchesSearch } from '../hooks/boardSearch';
import type { AssessedRealization } from '../hooks/domainBoardViewModel';
import { isMoveJourney } from '../lens/boardLens';
import { l1HasChange } from '../lens/journeyIndex';
import { AppChip } from './AppChip';
import { BoardCapabilityCard } from './BoardCapabilityCard';
import { useBoardLens } from './BoardLensContext';
import classes from './L1Group.module.css';

interface OwnAppsRowProps {
  capabilityId: CapabilityId;
  realizations: AssessedRealization[];
  showGrade: boolean;
  onChipClick: (componentId: ComponentId) => void;
}

function OwnAppsRow({ capabilityId, realizations, showGrade, onChipClick }: OwnAppsRowProps) {
  if (realizations.length === 0) return null;

  return (
    <Group gap="xs" wrap="wrap" className={classes.ownApps} data-testid={`l1-own-apps-${capabilityId}`}>
      {realizations.map((realization) => (
        <AppChip key={realization.id} realization={realization} onClick={onChipClick} showGrade={showGrade} />
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
  ownRealizations: AssessedRealization[];
  visibleChildren: CapabilityTreeNode[];
  showGrade: boolean;
  selectedCapabilities: Set<CapabilityId>;
  getColorForValue: (maturityValue: number) => string;
  getRealizationsForCapability: (capabilityId: CapabilityId) => AssessedRealization[];
  onCapabilityClick: (capability: Capability, event: React.MouseEvent) => void;
  onCapabilityContextMenu: (capability: Capability, event: React.MouseEvent) => void;
  onChipClick: (componentId: ComponentId) => void;
}

function L1GroupBody({
  node,
  ownRealizations,
  visibleChildren,
  showGrade,
  selectedCapabilities,
  getColorForValue,
  getRealizationsForCapability,
  onCapabilityClick,
  onCapabilityContextMenu,
  onChipClick,
}: L1GroupBodyProps) {
  return (
    <div className={classes.body}>
      <OwnAppsRow
        capabilityId={node.capability.id}
        realizations={ownRealizations}
        showGrade={showGrade}
        onChipClick={onChipClick}
      />
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
  getRealizationsForCapability: (capabilityId: CapabilityId) => AssessedRealization[];
  onCapabilityClick: (capability: Capability, event: React.MouseEvent) => void;
  onCapabilityContextMenu: (capability: Capability, event: React.MouseEvent) => void;
  onChipClick: (componentId: ComponentId) => void;
}

interface L1GroupOpenState {
  isOpen: boolean;
  dimmedGroup: boolean;
}

interface L1GroupStateParams {
  node: CapabilityTreeNode;
  lensContext: ReturnType<typeof useBoardLens>;
  searchQuery: string;
  forceOpen: boolean;
  manualOpen: boolean;
}

function computeL1GroupState({ node, lensContext, searchQuery, forceOpen, manualOpen }: L1GroupStateParams): L1GroupOpenState {
  const { lens, changesOnly, index } = lensContext;
  const changed = lens !== 'now' && l1HasChange(node, index);
  const dimmedGroup = changesOnly && lens !== 'now' && !changed;
  const isOpen = manualOpen || Boolean(searchQuery) || forceOpen || (changesOnly && changed);
  return { isOpen, dimmedGroup };
}

function ChildlessL1Group(props: L1GroupProps) {
  const { lens, index } = useBoardLens();
  const { node } = props;
  if (lens === 'target' && isMoveJourney(index.getJourney(node.capability.id))) return null;

  return (
    <div data-testid={`l1-group-${node.capability.id}`}>
      <BoardCapabilityCard
        node={node}
        isSelected={props.selectedCapabilities.has(node.capability.id)}
        getColorForValue={props.getColorForValue}
        getRealizationsForCapability={props.getRealizationsForCapability}
        onClick={props.onCapabilityClick}
        onContextMenu={props.onCapabilityContextMenu}
        onChipClick={props.onChipClick}
      />
    </div>
  );
}

export function L1Group(props: L1GroupProps) {
  const {
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
  } = props;
  const [manualOpen, setManualOpen] = useState(false);
  const lensContext = useBoardLens();

  if (node.children.length === 0) return <ChildlessL1Group {...props} />;

  const { isOpen, dimmedGroup } = computeL1GroupState({ node, lensContext, searchQuery, forceOpen, manualOpen });
  const visibleChildren = searchQuery
    ? node.children.filter((child) => nodeMatchesSearch(child, searchQuery, getRealizationsForCapability))
    : node.children;

  return (
    <div
      className={dimmedGroup ? classes.dimmedGroup : undefined}
      data-testid={`l1-group-${node.capability.id}`}
      data-dimmed={dimmedGroup || undefined}
    >
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
          showGrade={lensContext.lens === 'now'}
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
