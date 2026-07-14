import type { ComponentId } from '../../../api/types';
import type { CapabilityJourney } from '../../journeys/types';
import type { AssessedRealization } from '../hooks/domainBoardViewModel';
import { capabilityJourneyStatus } from '../lens/boardLens';
import { AppChip } from './AppChip';
import classes from './LensCardBody.module.css';
import { PlanChip } from './PlanChip';

export interface TargetCardBodyProps {
  journey: CapabilityJourney | undefined;
  realizations: AssessedRealization[];
  onChipClick: (componentId: ComponentId) => void;
}

function isProjectedTransition(journey: CapabilityJourney | undefined): journey is CapabilityJourney {
  return !!journey && journey.kind !== 'move' && capabilityJourneyStatus(journey) !== 'done';
}

export function TargetCardBody({ journey, realizations, onChipClick }: TargetCardBodyProps) {
  if (isProjectedTransition(journey)) {
    return (
      <div className={classes.row}>
        <PlanChip label={journey.toApplication.componentName} variant="standard" />
        {journey.kind === 'consolidation' && <span className={classes.consolidatedFlag}>consolidated</span>}
      </div>
    );
  }

  const standardChips = realizations.filter((realization) => realization.role === 'standard');
  const chips = standardChips.length > 0 ? standardChips : realizations;

  if (chips.length === 0) {
    return (
      <div className={classes.row}>
        <span className={classes.empty} data-testid="target-no-standard">
          no standard defined
        </span>
      </div>
    );
  }

  return (
    <div className={classes.row}>
      {chips.map((realization) => (
        <AppChip key={realization.id} realization={realization} onClick={onChipClick} showGrade={false} />
      ))}
    </div>
  );
}
