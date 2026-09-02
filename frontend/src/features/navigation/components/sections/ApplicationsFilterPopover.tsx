import { Chip, Group } from '@mantine/core';
import React from 'react';
import type { HostingClassification, OwnershipState } from '../../../../api/types';
import { HOSTING_CLASSIFICATION_LABELS } from '../../../components/components/ComponentHostingSection';
import { OWNERSHIP_STATE_LABELS } from '../../../components/components/ComponentOwnershipSection';
import { TreeFilterPopover } from '../shared/TreeFilterPopover';
import { TreeFilterSection } from '../shared/TreeFilterSection';

interface SingleChoiceFilterProps<T extends string> {
  label: string;
  labels: Record<T, string>;
  value: T | null;
  onChange: (value: T | null) => void;
}

function SingleChoiceFilter<T extends string>({ label, labels, value, onChange }: SingleChoiceFilterProps<T>) {
  return (
    <TreeFilterSection label={label} hasSelection={value !== null} onClear={() => onChange(null)}>
      <Chip.Group value={value ?? ''} onChange={(next) => onChange(next as T)}>
        <Group gap={4}>
          {(Object.keys(labels) as T[]).map((option) => (
            <Chip key={option} value={option} size="xs">
              {labels[option]}
            </Chip>
          ))}
        </Group>
      </Chip.Group>
    </TreeFilterSection>
  );
}

interface ApplicationsFilterPopoverProps {
  ownership: OwnershipState | null;
  onOwnershipChange: (value: OwnershipState | null) => void;
  hosting: HostingClassification | null;
  onHostingChange: (value: HostingClassification | null) => void;
}

export const ApplicationsFilterPopover: React.FC<ApplicationsFilterPopoverProps> = ({
  ownership,
  onOwnershipChange,
  hosting,
  onHostingChange,
}) => {
  const activeCount = Number(ownership !== null) + Number(hosting !== null);
  const clearAll = () => {
    onOwnershipChange(null);
    onHostingChange(null);
  };

  return (
    <TreeFilterPopover ariaLabel="Toggle application filters" activeCount={activeCount} onClearAll={clearAll}>
      <SingleChoiceFilter
        label="Ownership"
        labels={OWNERSHIP_STATE_LABELS}
        value={ownership}
        onChange={onOwnershipChange}
      />
      <SingleChoiceFilter
        label="Hosting"
        labels={HOSTING_CLASSIFICATION_LABELS}
        value={hosting}
        onChange={onHostingChange}
      />
    </TreeFilterPopover>
  );
};
