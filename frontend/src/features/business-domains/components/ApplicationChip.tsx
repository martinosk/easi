import { UnstyledButton } from '@mantine/core';
import type { CapabilityRealization, ComponentId } from '../../../api/types';
import classes from './ApplicationChip.module.css';

export interface ApplicationChipProps {
  realization: CapabilityRealization;
  onClick: (componentId: ComponentId) => void;
}

const REALIZATION_LEVEL_CLASS: Record<CapabilityRealization['realizationLevel'], string> = {
  Full: classes.realizationFull,
  Partial: classes.realizationPartial,
  Planned: classes.realizationPlanned,
};

const ORIGIN_CLASS: Record<CapabilityRealization['origin'], string> = {
  Direct: classes.originDirect,
  Inherited: classes.originInherited,
};

export function ApplicationChip({ realization, onClick }: ApplicationChipProps) {
  const componentName = realization.componentName || realization.componentId;
  const isInherited = realization.origin === 'Inherited';

  const tooltipText =
    isInherited && realization.sourceCapabilityName
      ? `${componentName} (inherited from ${realization.sourceCapabilityName})`
      : componentName;

  const chipClassName = [classes.chip, REALIZATION_LEVEL_CLASS[realization.realizationLevel], ORIGIN_CLASS[realization.origin]].join(' ');

  return (
    <UnstyledButton
      component="button"
      onClick={(e) => {
        e.stopPropagation();
        onClick(realization.componentId);
      }}
      title={tooltipText}
      className={chipClassName}
    >
      {isInherited && <span className={classes.inheritedIcon}>↓</span>}
      <span className={classes.chipName}>{componentName}</span>
    </UnstyledButton>
  );
}
