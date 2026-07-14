import { UnstyledButton } from '@mantine/core';
import type { CapabilityRealization, ComponentId, TimeGrade } from '../../../api/types';
import type { RealizationRole } from '../../architecture-direction/types';
import type { AssessedRealization } from '../hooks/domainBoardViewModel';
import classes from './AppChip.module.css';

export interface AppChipProps {
  realization: AssessedRealization;
  onClick: (componentId: ComponentId) => void;
  showGrade?: boolean;
}

const LEVEL_CLASS: Record<CapabilityRealization['realizationLevel'], string> = {
  Full: classes.full,
  Partial: classes.partial,
  Planned: classes.planned,
};

const ROLE_CLASS: Record<RealizationRole, string> = {
  standard: classes.roleStandard,
  legacy: classes.roleLegacy,
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

interface AppChipView {
  componentName: string;
  title: string;
  chipClassName: string;
  assessedGrade?: TimeGrade;
}

function resolveAppChipView(realization: AssessedRealization, showGrade: boolean): AppChipView {
  const componentName = realization.componentName || realization.componentId;
  const isInherited = realization.origin === 'Inherited';
  const role = isInherited ? undefined : realization.role;
  const tintClass = role ? ROLE_CLASS[role] : LEVEL_CLASS[realization.realizationLevel];

  return {
    componentName,
    assessedGrade: isInherited || !showGrade ? undefined : realization.timeGrade,
    title:
      isInherited && realization.sourceCapabilityName
        ? `${componentName} (inherited from ${realization.sourceCapabilityName})`
        : componentName,
    chipClassName: [classes.chip, tintClass, isInherited ? classes.inherited : ''].filter(Boolean).join(' '),
  };
}

export function AppChip({ realization, onClick, showGrade = true }: AppChipProps) {
  const { componentName, title, chipClassName, assessedGrade } = resolveAppChipView(realization, showGrade);

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
