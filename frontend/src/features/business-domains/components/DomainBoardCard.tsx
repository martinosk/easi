import { ActionIcon, Group, Paper, Text } from '@mantine/core';
import { IconDotsVertical } from '@tabler/icons-react';
import type { BusinessDomain, Capability, CapabilityId, ComponentId } from '../../../api/types';
import { nodeMatchesSearch } from '../hooks/boardSearch';
import type { DomainBoardViewModel } from '../hooks/domainBoardViewModel';
import { useBoardLens } from './BoardLensContext';
import classes from './DomainBoardCard.module.css';
import { L1Group } from './L1Group';
import { ArrivingMoves } from './MoveCards';

export interface DomainBoardCardProps {
  viewModel: DomainBoardViewModel;
  searchQuery: string;
  selectedCapabilities: Set<CapabilityId>;
  forceOpenL1Ids?: Set<CapabilityId>;
  isHighlighted?: boolean;
  isDropTarget?: boolean;
  getColorForValue: (maturityValue: number) => string;
  onCapabilityClick: (capability: Capability, event: React.MouseEvent) => void;
  onCapabilityContextMenu: (capability: Capability, event: React.MouseEvent) => void;
  onChipClick: (componentId: ComponentId) => void;
  onDomainMenu: (event: React.MouseEvent, domain: BusinessDomain) => void;
  onDragOver: (event: React.DragEvent) => void;
  onDragLeave: () => void;
  onDrop: (event: React.DragEvent) => void;
}

export function DomainBoardCard({
  viewModel,
  searchQuery,
  selectedCapabilities,
  forceOpenL1Ids,
  isHighlighted = false,
  isDropTarget = false,
  getColorForValue,
  onCapabilityClick,
  onCapabilityContextMenu,
  onChipClick,
  onDomainMenu,
  onDragOver,
  onDragLeave,
  onDrop,
}: DomainBoardCardProps) {
  const { domain, l1Groups, totalCapabilityCount, totalAppCount } = viewModel;
  const { index } = useBoardLens();

  const visibleGroups = searchQuery
    ? l1Groups.filter((group) => nodeMatchesSearch(group.node, searchQuery, viewModel.getRealizationsForCapability))
    : l1Groups;
  const arrivingAtTopLevel = index.getArrivingMovesForDomain(domain.id);

  return (
    <Paper
      p="md"
      radius="lg"
      className={[classes.card, isDropTarget ? classes.dropTarget : '', isHighlighted ? classes.highlighted : '']
        .filter(Boolean)
        .join(' ')}
      data-testid={`domain-card-${domain.id}`}
      onContextMenu={(e) => {
        e.preventDefault();
        onDomainMenu(e, domain);
      }}
      onDragOver={onDragOver}
      onDragLeave={onDragLeave}
      onDrop={onDrop}
    >
      <Group className={classes.headerRow} wrap="nowrap">
        <Text className={classes.domainName}>{domain.name}</Text>
        <Text className={classes.stats}>
          <span className={classes.statCount}>{totalCapabilityCount}</span> capabilities ·{' '}
          <span className={classes.statCount}>{totalAppCount}</span> apps
        </Text>
        <ActionIcon
          variant="subtle"
          color="gray"
          size="sm"
          aria-label={`Actions for ${domain.name}`}
          onClick={(e) => onDomainMenu(e, domain)}
        >
          <IconDotsVertical size={16} />
        </ActionIcon>
      </Group>

      {visibleGroups.length === 0 && arrivingAtTopLevel.length === 0 ? (
        <Text className={classes.emptyState}>
          {l1Groups.length === 0 ? 'No capabilities assigned to this domain yet.' : 'No matches in this domain.'}
        </Text>
      ) : (
        <div className={classes.groupsList}>
          {visibleGroups.map((group) => (
            <L1Group
              key={group.node.capability.id}
              node={group.node}
              distinctAppCount={group.distinctAppCount}
              searchQuery={searchQuery}
              forceOpen={forceOpenL1Ids?.has(group.node.capability.id)}
              selectedCapabilities={selectedCapabilities}
              getColorForValue={getColorForValue}
              getRealizationsForCapability={viewModel.getRealizationsForCapability}
              onCapabilityClick={onCapabilityClick}
              onCapabilityContextMenu={onCapabilityContextMenu}
              onChipClick={onChipClick}
            />
          ))}
          <ArrivingMoves journeys={arrivingAtTopLevel} />
        </div>
      )}
    </Paper>
  );
}
