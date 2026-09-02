import React from 'react';
import type { ArtifactCreator } from '../utils/filterByCreator';
import { CreatedByFilter } from './CreatedByFilter';
import { DomainFilter } from './DomainFilter';
import { TreeFilterPopover } from './shared/TreeFilterPopover';

interface FilterPopoverProps {
  artifactCreators: ArtifactCreator[];
  users: Array<{ id: string; name?: string; email: string }>;
  selectedCreatorIds: string[];
  onCreatorSelectionChange?: (creatorIds: string[]) => void;
  domains: Array<{ id: string; name: string }>;
  selectedDomainIds: string[];
  onDomainSelectionChange?: (domainIds: string[]) => void;
  onClearAllFilters?: () => void;
}

export const FilterPopover: React.FC<FilterPopoverProps> = ({
  artifactCreators,
  users,
  selectedCreatorIds,
  onCreatorSelectionChange,
  domains,
  selectedDomainIds,
  onDomainSelectionChange,
  onClearAllFilters,
}) => (
  <TreeFilterPopover
    ariaLabel="Toggle filters"
    activeCount={selectedCreatorIds.length + selectedDomainIds.length}
    onClearAll={onClearAllFilters}
  >
    {onCreatorSelectionChange && (
      <CreatedByFilter
        artifactCreators={artifactCreators}
        users={users}
        selectedCreatorIds={selectedCreatorIds}
        onSelectionChange={onCreatorSelectionChange}
      />
    )}
    {onDomainSelectionChange && (
      <DomainFilter
        domains={domains}
        selectedDomainIds={selectedDomainIds}
        onSelectionChange={onDomainSelectionChange}
      />
    )}
  </TreeFilterPopover>
);
