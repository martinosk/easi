import { ApplicationChip } from './ApplicationChip';
import type { CapabilityRealization, ComponentId } from '../../../api/types';

export interface ApplicationChipListProps {
  realizations: CapabilityRealization[];
  onApplicationClick: (componentId: ComponentId) => void;
}

const MAX_VISIBLE_CHIPS = 5;

export function ApplicationChipList({
  realizations,
  onApplicationClick,
}: ApplicationChipListProps) {
  const visibleRealizations = realizations.slice(0, MAX_VISIBLE_CHIPS);
  const overflowCount = realizations.length - MAX_VISIBLE_CHIPS;

  if (realizations.length === 0) {
    return null;
  }

  return (
    <div
      style={{
        display: 'flex',
        flexWrap: 'wrap',
        gap: '0.375rem',
        alignItems: 'center',
      }}
    >
      {visibleRealizations.map((realization) => (
        <ApplicationChip
          key={realization.id}
          realization={realization}
          onClick={onApplicationClick}
        />
      ))}
      {overflowCount > 0 && (
        <span
          style={{
            fontSize: '0.75rem',
            color: '#6b7280',
            fontWeight: 500,
            padding: '0.25rem',
          }}
        >
          +{overflowCount} more
        </span>
      )}
    </div>
  );
}
