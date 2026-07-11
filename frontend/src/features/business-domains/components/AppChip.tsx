import { UnstyledButton } from '@mantine/core';
import type { CapabilityRealization, ComponentId } from '../../../api/types';
import classes from './AppChip.module.css';

export interface AppChipProps {
  realization: CapabilityRealization;
  onClick: (componentId: ComponentId) => void;
}

const LEVEL_CLASS: Record<CapabilityRealization['realizationLevel'], string> = {
  Full: classes.full,
  Partial: classes.partial,
  Planned: classes.planned,
};

export function AppChip({ realization, onClick }: AppChipProps) {
  const componentName = realization.componentName || realization.componentId;
  const isInherited = realization.origin === 'Inherited';

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
    </UnstyledButton>
  );
}
