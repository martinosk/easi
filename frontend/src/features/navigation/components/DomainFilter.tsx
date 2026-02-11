import React, { useMemo } from 'react';
import { UNASSIGNED_DOMAIN } from '../utils/filterByDomain';

interface DomainFilterProps {
  domains: Array<{ id: string; name: string }>;
  selectedDomainIds: string[];
  onSelectionChange: (domainIds: string[]) => void;
}

export const DomainFilter: React.FC<DomainFilterProps> = ({
  domains,
  selectedDomainIds,
  onSelectionChange,
}) => {
  const selectedSet = useMemo(() => new Set(selectedDomainIds), [selectedDomainIds]);

  const handleToggle = (domainId: string) => {
    if (selectedSet.has(domainId)) {
      onSelectionChange(selectedDomainIds.filter((id) => id !== domainId));
    } else {
      onSelectionChange([...selectedDomainIds, domainId]);
    }
  };

  const handleClear = () => {
    onSelectionChange([]);
  };

  return (
    <div className="tree-filter">
      <div className="tree-filter-header">
        <span className="tree-filter-label">Assigned to domain</span>
        {selectedDomainIds.length > 0 && (
          <button
            className="tree-filter-clear"
            onClick={handleClear}
            aria-label="Clear filter"
          >
            Clear
          </button>
        )}
      </div>
      <div className="tree-filter-options">
        <button
          className={`tree-filter-option ${selectedSet.has(UNASSIGNED_DOMAIN) ? 'selected' : ''}`}
          onClick={() => handleToggle(UNASSIGNED_DOMAIN)}
        >
          Unassigned
        </button>
        {domains.map((domain) => (
          <button
            key={domain.id}
            className={`tree-filter-option ${selectedSet.has(domain.id) ? 'selected' : ''}`}
            onClick={() => handleToggle(domain.id)}
          >
            {domain.name}
          </button>
        ))}
      </div>
    </div>
  );
};
