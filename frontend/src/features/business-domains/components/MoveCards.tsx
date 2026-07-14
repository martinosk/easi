import { Box, UnstyledButton } from '@mantine/core';
import type { ComponentId } from '../../../api/types';
import type { CapabilityJourney } from '../../journeys/types';
import { formatTargetPeriod } from '../../journeys/utils/journeyFormat';
import type { AssessedRealization } from '../hooks/domainBoardViewModel';
import { AppChip } from './AppChip';
import { type TraceEnd, useBoardLens } from './BoardLensContext';
import classes from './MoveCards.module.css';
import { PlanChip } from './PlanChip';

function useMoveCard(journey: CapabilityJourney, end: TraceEnd) {
  const { tracedMoveId, activateTrace, registerTraceRef, openCapabilityById } = useBoardLens();
  return {
    traced: tracedMoveId === journey.id,
    setRef: (element: HTMLDivElement | null) => registerTraceRef(journey.id, end, element),
    open: () => openCapabilityById(journey.capabilityId),
    trace: (event: React.MouseEvent) => {
      event.stopPropagation();
      activateTrace(journey.id, end);
    },
  };
}

function cardActivation(open: () => void) {
  return (event: React.KeyboardEvent) => {
    if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault();
      open();
    }
  };
}

export interface GhostCardProps {
  journey: CapabilityJourney;
  realizations: AssessedRealization[];
  onChipClick: (componentId: ComponentId) => void;
}

export function GhostCard({ journey, realizations, onChipClick }: GhostCardProps) {
  const { traced, setRef, open, trace } = useMoveCard(journey, 'source');
  const destination = journey.move?.targetDomainName ?? '';

  return (
    <Box
      ref={setRef}
      className={[classes.card, classes.leaving, traced ? classes.traced : ''].filter(Boolean).join(' ')}
      role="button"
      tabIndex={0}
      onClick={open}
      onKeyDown={cardActivation(open)}
      data-testid={`move-ghost-${journey.capabilityId}`}
      data-traced={traced || undefined}
    >
      <div className={classes.row1}>
        <span className={classes.name}>{journey.capabilityName}</span>
        <span className={classes.level}>L1</span>
        <span className={[classes.pill, classes.pillIdle].join(' ')}>moving out</span>
      </div>
      <div className={classes.row2}>
        {realizations.map((realization) => (
          <AppChip key={realization.id} realization={realization} onClick={onChipClick} showGrade={false} />
        ))}
        <UnstyledButton
          component="button"
          className={classes.traceLink}
          onClick={trace}
          data-testid={`move-trace-source-${journey.id}`}
        >
          to {destination} ↗
        </UnstyledButton>
      </div>
    </Box>
  );
}

function ArrivingJourneyCard({ journey }: { journey: CapabilityJourney }) {
  const { traced, setRef, open, trace } = useMoveCard(journey, 'dest');
  const source = useBoardLens().index.sourceDomainName(journey.capabilityId) ?? '';

  return (
    <Box
      ref={setRef}
      className={[classes.card, classes.incoming, traced ? classes.traced : ''].filter(Boolean).join(' ')}
      role="button"
      tabIndex={0}
      onClick={open}
      onKeyDown={cardActivation(open)}
      data-testid={`move-arriving-${journey.capabilityId}`}
      data-traced={traced || undefined}
    >
      <div className={classes.row1}>
        <span className={classes.name}>{journey.move?.resultingName}</span>
        <span className={classes.level}>L3</span>
        <span className={[classes.pill, classes.pillIncoming].join(' ')}>
          arriving {formatTargetPeriod(journey.targetPeriod)}
        </span>
      </div>
      <div className={classes.row2}>
        <PlanChip label={journey.toApplication.componentName} variant="future" />
        <UnstyledButton
          component="button"
          className={classes.traceLink}
          onClick={trace}
          data-testid={`move-trace-dest-${journey.id}`}
        >
          from {source} ↗
        </UnstyledButton>
      </div>
    </Box>
  );
}

function ArrivingTargetCard({ journey }: { journey: CapabilityJourney }) {
  const { open } = useMoveCard(journey, 'dest');
  const source = useBoardLens().index.sourceDomainName(journey.capabilityId) ?? '';

  return (
    <Box
      className={[classes.card, classes.done].join(' ')}
      role="button"
      tabIndex={0}
      onClick={open}
      onKeyDown={cardActivation(open)}
      data-testid={`move-arriving-${journey.capabilityId}`}
    >
      <div className={classes.row1}>
        <span className={classes.name}>{journey.move?.resultingName}</span>
        <span className={classes.level}>L3</span>
      </div>
      <div className={classes.row2}>
        <PlanChip label={journey.toApplication.componentName} variant="standard" />
        <span className={classes.movedFlag}>moved from {source}</span>
      </div>
    </Box>
  );
}

export function ArrivingMoves({ journeys }: { journeys: CapabilityJourney[] }) {
  const { lens } = useBoardLens();
  if (lens === 'now' || journeys.length === 0) return null;

  return (
    <>
      {journeys.map((journey) =>
        lens === 'journey' ? (
          <ArrivingJourneyCard key={journey.id} journey={journey} />
        ) : (
          <ArrivingTargetCard key={journey.id} journey={journey} />
        ),
      )}
    </>
  );
}
