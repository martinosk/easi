import { CloseButton } from '@mantine/core';
import type { StageCapabilityMapping } from '../../../api/types';
import { hasLink } from '../../../utils/hateoas';
import classes from './CapabilityChip.module.css';

interface CapabilityChipProps {
  mapping: StageCapabilityMapping;
  canRemove: boolean;
  onRemove: (mapping: StageCapabilityMapping) => void;
}

export function CapabilityChip({ mapping, canRemove, onRemove }: CapabilityChipProps) {
  const showRemove = canRemove && hasLink(mapping, 'delete');
  return (
    <span className={classes.chip} data-testid={`cap-chip-${mapping.capabilityId}`}>
      <span className={classes.name}>{mapping.capabilityName || mapping.capabilityId}</span>
      {showRemove && (
        <CloseButton
          size="xs"
          className={classes.remove}
          onClick={(e) => {
            e.stopPropagation();
            onRemove(mapping);
          }}
          title="Remove capability"
          aria-label="Remove capability"
        />
      )}
    </span>
  );
}
