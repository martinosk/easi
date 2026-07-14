import type { CapabilityId, ComponentId } from '../../../api/types';
import type { CapabilityTreeNode } from '../../capabilities/hooks/useCapabilityTree';
import { JourneyProgressBar } from '../../journeys/components/JourneyProgressBar';
import type { CapabilityJourney } from '../../journeys/types';
import { journeyKindLabel } from '../../journeys/utils/journeyFormat';
import type { AssessedRealization } from '../hooks/domainBoardViewModel';
import { capabilityJourneyStatus } from '../lens/boardLens';
import { AppChip } from './AppChip';
import classes from './LensCardBody.module.css';
import { PlanChip } from './PlanChip';
import { SubCapabilityBreakdown } from './SubCapabilityBreakdown';

function fromLabel(journey: CapabilityJourney): string {
  return journey.fromApplications.map((app) => app.componentName).join(', ');
}

interface DoneBodyProps {
  journey: CapabilityJourney;
  realizations: AssessedRealization[];
  onChipClick: (componentId: ComponentId) => void;
}

function DoneBody({ journey, realizations, onChipClick }: DoneBodyProps) {
  const source = fromLabel(journey);
  const standardChips = realizations.filter((realization) => realization.role === 'standard');

  return (
    <>
      <div className={classes.row}>
        {source && <span className={classes.fromLabel}>{source} → </span>}
        {standardChips.length > 0 ? (
          standardChips.map((realization) => (
            <AppChip key={realization.id} realization={realization} onClick={onChipClick} showGrade={false} />
          ))
        ) : (
          <PlanChip label={journey.toApplication.componentName} variant="standard" />
        )}
      </div>
      <JourneyProgressBar journey={journey} />
    </>
  );
}

function ActiveBody({ journey }: { journey: CapabilityJourney }) {
  return (
    <>
      <div className={classes.row}>
        <PlanChip label={fromLabel(journey) || '—'} variant="legacy" />
        <span className={classes.arrow}>→</span>
        <PlanChip label={journey.toApplication.componentName} variant="future" />
        <span className={classes.kind}>{journeyKindLabel(journey.kind)}</span>
      </div>
      <JourneyProgressBar journey={journey} />
    </>
  );
}

function SteadyBody({
  realizations,
  onChipClick,
}: {
  realizations: AssessedRealization[];
  onChipClick: (componentId: ComponentId) => void;
}) {
  return (
    <div className={classes.row}>
      {realizations.length === 0 ? (
        <span className={classes.empty}>unmapped</span>
      ) : (
        realizations.map((realization) => (
          <AppChip key={realization.id} realization={realization} onClick={onChipClick} showGrade={false} />
        ))
      )}
      <span className={classes.noChange}>no change planned</span>
    </div>
  );
}

export interface JourneyLensBodyProps {
  node: CapabilityTreeNode;
  journey: CapabilityJourney | undefined;
  realizations: AssessedRealization[];
  hideSubCapabilities?: boolean;
  getRealizationsForCapability: (capabilityId: CapabilityId) => AssessedRealization[];
  onChipClick: (componentId: ComponentId) => void;
}

export function JourneyLensBody({
  node,
  journey,
  realizations,
  hideSubCapabilities,
  getRealizationsForCapability,
  onChipClick,
}: JourneyLensBodyProps) {
  if (!journey) return <SteadyBody realizations={realizations} onChipClick={onChipClick} />;

  return (
    <>
      {capabilityJourneyStatus(journey) === 'done' ? (
        <DoneBody journey={journey} realizations={realizations} onChipClick={onChipClick} />
      ) : (
        <ActiveBody journey={journey} />
      )}
      {!hideSubCapabilities && (
        <SubCapabilityBreakdown
          node={node}
          journey={journey}
          getRealizationsForCapability={getRealizationsForCapability}
        />
      )}
    </>
  );
}
