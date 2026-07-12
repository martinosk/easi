import { UnstyledButton } from '@mantine/core';
import type { CapabilityRealization, ComponentId, TimeGrade } from '../../../api/types';
import type { AssessedRealization } from '../hooks/domainBoardViewModel';
import classes from './AppChip.module.css';

export interface AppChipProps {
  realization: AssessedRealization;
  onClick: (componentId: ComponentId) => void;
}

const LEVEL_CLASS: Record<CapabilityRealization['realizationLevel'], string> = {
  Full: classes.full,
  Partial: classes.partial,
  Planned: classes.planned,
};

const GRADE_CLASS: Record<TimeGrade, string> = {
  Invest: classes.gradeInvest,
  Tolerate: classes.gradeTolerate,
  Migrate: classes.gradeMigrate,
  Eliminate: classes.gradeEliminate,
};

const GRADE_LETTER: Record<TimeGrade, string> = {
  Invest: 'I',
  Tolerate: 'T',
  Migrate: 'M',
  Eliminate: 'E',
};

function GradeBadge({ componentId, grade }: { componentId: ComponentId; grade: TimeGrade }) {
  return (
    <span
      className={[classes.gradeBadge, GRADE_CLASS[grade]].join(' ')}
      title={`${grade} — for this capability`}
      data-testid={`app-chip-grade-${componentId}`}
    >
      {GRADE_LETTER[grade]}
    </span>
  );
}

export function AppChip({ realization, onClick }: AppChipProps) {
  const componentName = realization.componentName || realization.componentId;
  const isInherited = realization.origin === 'Inherited';
  const assessedGrade = isInherited ? undefined : realization.timeGrade;

  const title =
    isInherited && realization.sourceCapabilityName
      ? `${componentName} (inherited from ${realization.sourceCapabilityName})`
      : componentName;

  const chipClassName = [classes.chip, LEVEL_CLASS[realization.realizationLevel], isInherited ? classes.inherited : '']
    .filter(Boolean)
    .join(' ');

  return (
    <UnstyledButton
      component="button"
      onClick={(e) => {
        e.stopPropagation();
        onClick(realization.componentId);
      }}
      title={title}
      className={chipClassName}
      data-testid={`app-chip-${realization.componentId}`}
    >
      <span className={classes.chipName}>{componentName}</span>
      {assessedGrade && <GradeBadge componentId={realization.componentId} grade={assessedGrade} />}
    </UnstyledButton>
  );
}
