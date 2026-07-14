import { UnstyledButton } from '@mantine/core';
import { IconChevronRight } from '@tabler/icons-react';
import { useState } from 'react';
import type { Capability, CapabilityId, ComponentId } from '../../../api/types';
import type { CapabilityTreeNode } from '../../capabilities/hooks/useCapabilityTree';
import { nodeMatchesSearch } from '../hooks/boardSearch';
import type { AssessedRealization } from '../hooks/domainBoardViewModel';
import { isMoveJourney } from '../lens/boardLens';
import { l1HasChange } from '../lens/journeyIndex';
import { BoardCapabilityCard } from './BoardCapabilityCard';
import { useBoardLens } from './BoardLensContext';
import classes from './L1Group.module.css';

interface AppCountProps {
  distinctAppCount: number;
}

function AppCount({ distinctAppCount }: AppCountProps) {
  return (
    <span className={[classes.appCount, distinctAppCount > 3 ? classes.appCountMulti : ''].join(' ')}>
      {distinctAppCount} apps
    </span>
  );
}

interface L1CompactHeaderProps {
  capability: Capability;
  subCount: number;
  distinctAppCount: number;
  isSelected: boolean;
  onToggle: () => void;
}

function L1CompactHeader({ capability, subCount, distinctAppCount, isSelected, onToggle }: L1CompactHeaderProps) {
  return (
    <UnstyledButton
      component="button"
      className={[classes.header, isSelected ? classes.selected : ''].filter(Boolean).join(' ')}
      aria-expanded={false}
      aria-label={`Expand ${capability.name}`}
      onClick={onToggle}
      data-testid={`l1-toggle-${capability.id}`}
      data-selected={isSelected || undefined}
    >
      <IconChevronRight size={12} className={classes.chevron} />
      <span className={classes.name}>{capability.name}</span>
      <span className={classes.subTag}>{subCount > 0 ? `L1 · ${subCount} sub` : 'L1'}</span>
      <AppCount distinctAppCount={distinctAppCount} />
    </UnstyledButton>
  );
}

interface L1CollapseToggleProps {
  capabilityId: CapabilityId;
  subCount: number;
  distinctAppCount: number;
  onToggle: () => void;
}

function L1CollapseToggle({ capabilityId, subCount, distinctAppCount, onToggle }: L1CollapseToggleProps) {
  return (
    <UnstyledButton
      component="button"
      className={classes.childToggle}
      aria-expanded
      aria-label="Collapse capability"
      onClick={(event) => {
        event.stopPropagation();
        onToggle();
      }}
      data-testid={`l1-toggle-${capabilityId}`}
    >
      <IconChevronRight size={12} className={[classes.chevron, classes.chevronOpen].join(' ')} />
      {subCount > 0 && (
        <span>
          {subCount} sub-capabilit{subCount === 1 ? 'y' : 'ies'}
        </span>
      )}
      <AppCount distinctAppCount={distinctAppCount} />
    </UnstyledButton>
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

interface ExpandedL1GroupProps extends L1GroupProps {
  isMove: boolean;
  onToggle: () => void;
}

function ExpandedL1Group({ isMove, onToggle, ...props }: ExpandedL1GroupProps) {
  const { node, distinctAppCount, searchQuery, selectedCapabilities, getRealizationsForCapability } = props;
  const hasChildren = node.children.length > 0;
  const visibleChildren =
    hasChildren && searchQuery
      ? node.children.filter((child) => nodeMatchesSearch(child, searchQuery, getRealizationsForCapability))
      : node.children;

  return (
    <>
      <BoardCapabilityCard
        node={node}
        isSelected={selectedCapabilities.has(node.capability.id)}
        subCapabilities={
          isMove ? undefined : (
            <L1CollapseToggle
              capabilityId={node.capability.id}
              subCount={node.children.length}
              distinctAppCount={distinctAppCount}
              onToggle={onToggle}
            />
          )
        }
        getColorForValue={props.getColorForValue}
        getRealizationsForCapability={getRealizationsForCapability}
        onClick={props.onCapabilityClick}
        onContextMenu={props.onCapabilityContextMenu}
        onChipClick={props.onChipClick}
      />
      {hasChildren && (
        <div className={classes.childCards}>
          {visibleChildren.map((child) => (
            <BoardCapabilityCard
              key={child.capability.id}
              node={child}
              isSelected={selectedCapabilities.has(child.capability.id)}
              getColorForValue={props.getColorForValue}
              getRealizationsForCapability={getRealizationsForCapability}
              onClick={props.onCapabilityClick}
              onContextMenu={props.onCapabilityContextMenu}
              onChipClick={props.onChipClick}
            />
          ))}
        </div>
      )}
    </>
  );
}

export function L1Group(props: L1GroupProps) {
  const { node, distinctAppCount, searchQuery, forceOpen = false, selectedCapabilities } = props;
  const [manualOpen, setManualOpen] = useState(false);
  const lensContext = useBoardLens();
  const { lens, index } = lensContext;
  const isMove = isMoveJourney(index.getJourney(node.capability.id));

  if (lens === 'target' && isMove) return null;

  const { isOpen, dimmedGroup } = computeL1GroupState({ node, lensContext, searchQuery, forceOpen, manualOpen });
  const toggle = () => setManualOpen((open) => !open);

  return (
    <div
      className={dimmedGroup ? classes.dimmedGroup : undefined}
      data-testid={`l1-group-${node.capability.id}`}
      data-dimmed={dimmedGroup || undefined}
    >
      {isOpen || isMove ? (
        <ExpandedL1Group {...props} isMove={isMove} onToggle={toggle} />
      ) : (
        <L1CompactHeader
          capability={node.capability}
          subCount={node.children.length}
          distinctAppCount={distinctAppCount}
          isSelected={selectedCapabilities.has(node.capability.id)}
          onToggle={toggle}
        />
      )}
    </div>
  );
}
