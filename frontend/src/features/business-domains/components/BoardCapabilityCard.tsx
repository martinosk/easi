import { Box, Group, Text, UnstyledButton } from '@mantine/core';
import { IconChevronRight } from '@tabler/icons-react';
import { useState } from 'react';
import type { Capability, CapabilityId, CapabilityRealization, ComponentId } from '../../../api/types';
import type { CapabilityTreeNode } from '../../capabilities/hooks/useCapabilityTree';
import { AppChip } from './AppChip';
import classes from './BoardCapabilityCard.module.css';

function flattenDescendants(node: CapabilityTreeNode): CapabilityTreeNode[] {
  return node.children.flatMap((child) => [child, ...flattenDescendants(child)]);
}

function activationKeyHandler(
  capability: Capability,
  onClick: (capability: Capability, event: React.MouseEvent) => void,
) {
  return (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      onClick(capability, e as unknown as React.MouseEvent);
    }
  };
}

interface CardRealizationsProps {
  realizations: CapabilityRealization[];
  onChipClick: (componentId: ComponentId) => void;
}

function CardRealizations({ realizations, onChipClick }: CardRealizationsProps) {
  if (realizations.length === 0) {
    return (
      <Text className={classes.empty} data-testid="capability-card-empty-realizations">
        no realising application mapped
      </Text>
    );
  }

  return (
    <Group gap="xs" wrap="wrap" className={classes.row2}>
      {realizations.map((realization) => (
        <AppChip key={realization.id} realization={realization} onClick={onChipClick} />
      ))}
      {realizations.length > 1 && <Text className={classes.multiFlag}>{realizations.length} apps</Text>}
    </Group>
  );
}

interface ChildRowProps {
  node: CapabilityTreeNode;
  getRealizationsForCapability: (capabilityId: CapabilityId) => CapabilityRealization[];
  onClick: (capability: Capability, event: React.MouseEvent) => void;
  onContextMenu: (capability: Capability, event: React.MouseEvent) => void;
  onChipClick: (componentId: ComponentId) => void;
}

function ChildRow({ node, getRealizationsForCapability, onClick, onContextMenu, onChipClick }: ChildRowProps) {
  const realizations = getRealizationsForCapability(node.capability.id);

  return (
    <Box
      className={classes.childRow}
      role="button"
      tabIndex={0}
      onClick={(e) => onClick(node.capability, e)}
      onKeyDown={activationKeyHandler(node.capability, onClick)}
      onContextMenu={(e) => {
        e.preventDefault();
        e.stopPropagation();
        onContextMenu(node.capability, e);
      }}
      data-testid={`capability-card-${node.capability.id}`}
    >
      <span className={classes.childDot} />
      <Text fw={500}>{node.capability.name}</Text>
      <span className={classes.childLevel}>{node.capability.level}</span>
      <div className={classes.childApps}>
        {realizations.map((realization) => (
          <AppChip key={realization.id} realization={realization} onClick={onChipClick} />
        ))}
      </div>
    </Box>
  );
}

interface ChildrenExpanderProps {
  descendants: CapabilityTreeNode[];
  isOpen: boolean;
  onToggle: () => void;
  getRealizationsForCapability: (capabilityId: CapabilityId) => CapabilityRealization[];
  onClick: (capability: Capability, event: React.MouseEvent) => void;
  onContextMenu: (capability: Capability, event: React.MouseEvent) => void;
  onChipClick: (componentId: ComponentId) => void;
}

function ChildrenExpander({
  descendants,
  isOpen,
  onToggle,
  getRealizationsForCapability,
  onClick,
  onContextMenu,
  onChipClick,
}: ChildrenExpanderProps) {
  return (
    <>
      <UnstyledButton
        component="button"
        className={classes.expander}
        aria-expanded={isOpen}
        onClick={(e) => {
          e.stopPropagation();
          onToggle();
        }}
      >
        <IconChevronRight size={12} className={[classes.chevron, isOpen ? classes.chevronOpen : ''].join(' ')} />
        {descendants.length} sub-capabilit{descendants.length === 1 ? 'y' : 'ies'}
      </UnstyledButton>
      {isOpen && (
        <Box className={classes.children}>
          {descendants.map((child) => (
            <ChildRow
              key={child.capability.id}
              node={child}
              getRealizationsForCapability={getRealizationsForCapability}
              onClick={onClick}
              onContextMenu={onContextMenu}
              onChipClick={onChipClick}
            />
          ))}
        </Box>
      )}
    </>
  );
}

export interface BoardCapabilityCardProps {
  node: CapabilityTreeNode;
  isSelected: boolean;
  getColorForValue: (maturityValue: number) => string;
  getRealizationsForCapability: (capabilityId: CapabilityId) => CapabilityRealization[];
  onClick: (capability: Capability, event: React.MouseEvent) => void;
  onContextMenu: (capability: Capability, event: React.MouseEvent) => void;
  onChipClick: (componentId: ComponentId) => void;
}

export function BoardCapabilityCard({
  node,
  isSelected,
  getColorForValue,
  getRealizationsForCapability,
  onClick,
  onContextMenu,
  onChipClick,
}: BoardCapabilityCardProps) {
  const [childrenOpen, setChildrenOpen] = useState(false);
  const { capability } = node;
  const descendants = flattenDescendants(node);
  const realizations = getRealizationsForCapability(capability.id);
  const borderLeftColor =
    capability.maturityValue !== undefined ? getColorForValue(capability.maturityValue) : undefined;

  return (
    <Box
      className={[classes.card, isSelected ? classes.selected : ''].filter(Boolean).join(' ')}
      style={borderLeftColor ? { borderLeftColor } : undefined}
      role="button"
      tabIndex={0}
      onClick={(e) => onClick(capability, e)}
      onKeyDown={activationKeyHandler(capability, onClick)}
      onContextMenu={(e) => {
        e.preventDefault();
        onContextMenu(capability, e);
      }}
      data-testid={`capability-card-${capability.id}`}
      data-selected={isSelected || undefined}
    >
      <div className={classes.row1}>
        <Text className={classes.name}>{capability.name}</Text>
        <span className={classes.levelTag}>{capability.level}</span>
      </div>
      <CardRealizations realizations={realizations} onChipClick={onChipClick} />
      {descendants.length > 0 && (
        <ChildrenExpander
          descendants={descendants}
          isOpen={childrenOpen}
          onToggle={() => setChildrenOpen((open) => !open)}
          getRealizationsForCapability={getRealizationsForCapability}
          onClick={onClick}
          onContextMenu={onContextMenu}
          onChipClick={onChipClick}
        />
      )}
    </Box>
  );
}
