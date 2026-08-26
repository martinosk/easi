import { Chip, Group } from '@mantine/core';
import React from 'react';
import { UNASSIGNED_DOMAIN } from '../utils/filterByDomain';
import { TreeFilterSection } from './shared/TreeFilterSection';

interface DomainFilterProps {
  domains: Array<{ id: string; name: string }>;
  selectedDomainIds: string[];
  onSelectionChange: (domainIds: string[]) => void;
}

export const DomainFilter: React.FC<DomainFilterProps> = ({ domains, selectedDomainIds, onSelectionChange }) => (
  <TreeFilterSection
    label="Assigned to domain"
    hasSelection={selectedDomainIds.length > 0}
    onClear={() => onSelectionChange([])}
  >
    <Chip.Group multiple value={selectedDomainIds} onChange={onSelectionChange}>
      <Group gap={4}>
        <Chip value={UNASSIGNED_DOMAIN} size="xs">
          Unassigned
        </Chip>
        {domains.map((domain) => (
          <Chip key={domain.id} value={domain.id} size="xs">
            {domain.name}
          </Chip>
        ))}
      </Group>
    </Chip.Group>
  </TreeFilterSection>
);
