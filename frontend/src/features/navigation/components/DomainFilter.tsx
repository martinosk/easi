import { Chip, Group, UnstyledButton } from '@mantine/core';
import React from 'react';
import { UNASSIGNED_DOMAIN } from '../utils/filterByDomain';

interface DomainFilterProps {
  domains: Array<{ id: string; name: string }>;
  selectedDomainIds: string[];
  onSelectionChange: (domainIds: string[]) => void;
}

export const DomainFilter: React.FC<DomainFilterProps> = ({ domains, selectedDomainIds, onSelectionChange }) => {
  const handleClear = () => onSelectionChange([]);

  return (
    <div className="tree-filter">
      <div className="tree-filter-header">
        <span className="tree-filter-label">Assigned to domain</span>
        {selectedDomainIds.length > 0 && (
          <UnstyledButton
            component="button"
            type="button"
            className="tree-filter-clear"
            onClick={handleClear}
            aria-label="Clear filter"
          >
            Clear
          </UnstyledButton>
        )}
      </div>
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
    </div>
  );
};
