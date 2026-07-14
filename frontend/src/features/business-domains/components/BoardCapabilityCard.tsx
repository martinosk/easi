import { Box, Text } from '@mantine/core';
import type { CSSProperties } from 'react';
import type { Capability, CapabilityId, ComponentId } from '../../../api/types';
import type { CapabilityTreeNode } from '../../capabilities/hooks/useCapabilityTree';
import type { AssessedRealization } from '../hooks/domainBoardViewModel';
import { type BoardJourneyStatus, type BoardLens, capabilityJourneyStatus, isMoveJourney } from '../lens/boardLens';
import { capabilityHasChange } from '../lens/journeyIndex';
import classes from './BoardCapabilityCard.module.css';
import { useBoardLens } from './BoardLensContext';
import { activationKeyHandler } from './boardCardKeyboard';
import { JourneyLensBody } from './JourneyCardBody';
import { ArrivingMoves, GhostCard } from './MoveCards';
import { NowCardContent } from './NowCardContent';
import { TargetCardBody } from './TargetCardBody';

const PILL_CONFIG: Record<BoardJourneyStatus, { className: string; label: string } | null> = {
  steady: null,
  'not-started': { className: classes.pillIdle, label: 'not started' },
  'in-flight': { className: classes.pillFlight, label: 'in flight' },
  done: { className: classes.pillDone, label: 'done' },
  'planned-move': { className: classes.pillIncoming, label: 'move planned' },
};

const STATUS_BORDER: Partial<Record<BoardJourneyStatus, string>> = {
  done: classes.statusDone,
  'in-flight': classes.statusFlight,
};

function StatusPill({ status }: { status: BoardJourneyStatus }) {
  const config = PILL_CONFIG[status];
  if (!config) return null;
  return (
    <span className={[classes.pill, config.className].join(' ')} data-testid={`status-pill-${status}`}>
      {config.label}
    </span>
  );
}

export interface BoardCapabilityCardProps {
  node: CapabilityTreeNode;
  isSelected: boolean;
  getColorForValue: (maturityValue: number) => string;
  getRealizationsForCapability: (capabilityId: CapabilityId) => AssessedRealization[];
  onClick: (capability: Capability, event: React.MouseEvent) => void;
  onContextMenu: (capability: Capability, event: React.MouseEvent) => void;
  onChipClick: (componentId: ComponentId) => void;
}

interface ShellProps extends BoardCapabilityCardProps {
  lens: BoardLens;
  status: BoardJourneyStatus;
  dimmed: boolean;
  children: React.ReactNode;
}

function shellChrome({
  capability,
  lens,
  status,
  isSelected,
  dimmed,
  getColorForValue,
}: {
  capability: Capability;
  lens: BoardLens;
  status: BoardJourneyStatus;
  isSelected: boolean;
  dimmed: boolean;
  getColorForValue: (maturityValue: number) => string;
}): { className: string; style?: CSSProperties } {
  const maturityBorder =
    lens === 'now' && capability.maturityValue !== undefined ? getColorForValue(capability.maturityValue) : undefined;
  const statusBorderClass = lens === 'now' ? '' : (STATUS_BORDER[status] ?? '');
  const className = [classes.card, statusBorderClass, isSelected ? classes.selected : '', dimmed ? classes.dimmed : '']
    .filter(Boolean)
    .join(' ');
  return { className, style: maturityBorder ? { borderLeftColor: maturityBorder } : undefined };
}

function CapabilityCardShell({
  node,
  lens,
  status,
  isSelected,
  dimmed,
  getColorForValue,
  onClick,
  onContextMenu,
  children,
}: ShellProps) {
  const { capability } = node;
  const { className, style } = shellChrome({ capability, lens, status, isSelected, dimmed, getColorForValue });

  return (
    <Box
      className={className}
      style={style}
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
      data-dimmed={dimmed || undefined}
    >
      <div className={classes.row1}>
        <Text className={classes.name}>{capability.name}</Text>
        <span className={classes.levelTag}>{capability.level}</span>
        {lens === 'journey' && <StatusPill status={status} />}
      </div>
      {children}
    </Box>
  );
}

export function BoardCapabilityCard(props: BoardCapabilityCardProps) {
  const { node, getRealizationsForCapability, onChipClick } = props;
  const { lens, changesOnly, index } = useBoardLens();
  const { capability } = node;
  const journey = index.getJourney(capability.id);
  const arriving = index.getArrivingMovesForParent(capability.id);
  const realizations = getRealizationsForCapability(capability.id);

  if (isMoveJourney(journey)) {
    if (lens === 'journey') {
      return (
        <>
          <GhostCard journey={journey} realizations={realizations} onChipClick={onChipClick} />
          <ArrivingMoves journeys={arriving} />
        </>
      );
    }
    if (lens === 'target') return <ArrivingMoves journeys={arriving} />;
  }

  const status = capabilityJourneyStatus(journey);
  const dimmed = changesOnly && lens !== 'now' && !capabilityHasChange(capability.id, index);

  const body =
    lens === 'now' ? (
      <NowCardContent
        node={node}
        realizations={realizations}
        getRealizationsForCapability={getRealizationsForCapability}
        onClick={props.onClick}
        onContextMenu={props.onContextMenu}
        onChipClick={onChipClick}
      />
    ) : lens === 'target' ? (
      <TargetCardBody journey={journey} realizations={realizations} onChipClick={onChipClick} />
    ) : (
      <JourneyLensBody
        node={node}
        journey={journey}
        realizations={realizations}
        getRealizationsForCapability={getRealizationsForCapability}
        onChipClick={onChipClick}
      />
    );

  return (
    <>
      <CapabilityCardShell {...props} lens={lens} status={status} dimmed={dimmed}>
        {body}
      </CapabilityCardShell>
      <ArrivingMoves journeys={arriving} />
    </>
  );
}
