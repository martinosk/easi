import type { StageCapabilityMapping } from '../../../api/types';
import { hasLink } from '../../../utils/hateoas';

interface CapabilityChipProps {
  mapping: StageCapabilityMapping;
  canRemove: boolean;
  onRemove: (mapping: StageCapabilityMapping) => void;
}

export function CapabilityChip({ mapping, canRemove, onRemove }: CapabilityChipProps) {
  return (
    <span className="cap-chip" data-testid={`cap-chip-${mapping.capabilityId}`}>
      <span className="cap-chip-name">{mapping.capabilityName || mapping.capabilityId}</span>
      {canRemove && hasLink(mapping, 'delete') && (
        <button
          type="button"
          className="cap-chip-remove"
          onClick={(e) => { e.stopPropagation(); onRemove(mapping); }}
          title="Remove capability"
        >
          &times;
        </button>
      )}
    </span>
  );
}
