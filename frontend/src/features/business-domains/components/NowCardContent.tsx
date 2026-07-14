import { Box, Group, Text, UnstyledButton } from '@mantine/core';
import { IconChevronRight } from '@tabler/icons-react';
import { useState } from 'react';
import type { Capability, CapabilityId, ComponentId } from '../../../api/types';
import type { CapabilityTreeNode } from '../../capabilities/hooks/useCapabilityTree';
import type { AssessedRealization } from '../hooks/domainBoardViewModel';
import { AppChip } from './AppChip';
import classes from './BoardCapabilityCard.module.css';
import { activationKeyHandler } from './boardCardKeyboard';

function flattenDescendants(node: CapabilityTreeNode): CapabilityTreeNode[] {
  return node.children.flatMap((child) => [child, ...flattenDescendants(child)]);
}

interface CardRealizationsProps {
  realizations: AssessedRealization[];
  onChipClick: (componentId: ComponentId) => void;
}

export function CardRealizations({ realizations, onChipClick }: CardRealizationsProps) {
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
  getRealizationsForCapability: (capabilityId: CapabilityId) => AssessedRealization[];
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

export interface NowCardContentProps {
  node: CapabilityTreeNode;
  realizations: AssessedRealization[];
  getRealizationsForCapability: (capabilityId: CapabilityId) => AssessedRealization[];
  onClick: (capability: Capability, event: React.MouseEvent) => void;
  onContextMenu: (capability: Capability, event: React.MouseEvent) => void;
  onChipClick: (componentId: ComponentId) => void;
}

export function NowCardContent({
  node,
  realizations,
  getRealizationsForCapability,
  onClick,
  onContextMenu,
  onChipClick,
}: NowCardContentProps) {
  const [childrenOpen, setChildrenOpen] = useState(false);
  const descendants = flattenDescendants(node);

  return (
    <>
      <CardRealizations realizations={realizations} onChipClick={onChipClick} />
      {descendants.length > 0 && (
        <>
          <UnstyledButton
            component="button"
            className={classes.expander}
            aria-expanded={childrenOpen}
            onClick={(e) => {
              e.stopPropagation();
              setChildrenOpen((open) => !open);
            }}
          >
            <IconChevronRight
              size={12}
              className={[classes.chevron, childrenOpen ? classes.chevronOpen : ''].join(' ')}
            />
            {descendants.length} sub-capabilit{descendants.length === 1 ? 'y' : 'ies'}
          </UnstyledButton>
          {childrenOpen && (
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
      )}
    </>
  );
}
