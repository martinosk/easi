import { ActionIcon, Button, Indicator } from '@mantine/core';
import React, { useCallback, useEffect, useRef, useState } from 'react';
import type { ArtifactCreator } from '../utils/filterByCreator';
import { CreatedByFilter } from './CreatedByFilter';
import { DomainFilter } from './DomainFilter';

interface FilterPopoverProps {
  artifactCreators: ArtifactCreator[];
  users: Array<{ id: string; name?: string; email: string }>;
  selectedCreatorIds: string[];
  onCreatorSelectionChange?: (creatorIds: string[]) => void;
  domains: Array<{ id: string; name: string }>;
  selectedDomainIds: string[];
  onDomainSelectionChange?: (domainIds: string[]) => void;
  hasActiveFilters: boolean;
  onClearAllFilters?: () => void;
}

function useClickOutside(ref: React.RefObject<HTMLDivElement | null>, isOpen: boolean, onClose: () => void) {
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (ref.current && !ref.current.contains(event.target as Node)) {
        onClose();
      }
    }

    if (isOpen) {
      document.addEventListener('mousedown', handleClickOutside);
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [isOpen, ref, onClose]);
}

const FILTER_ICON = (
  <svg
    width="14"
    height="14"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    strokeWidth="2"
    strokeLinecap="round"
    strokeLinejoin="round"
  >
    <polygon points="22 3 2 3 10 12.46 10 19 14 21 14 12.46 22 3" />
  </svg>
);

export const FilterPopover: React.FC<FilterPopoverProps> = ({
  artifactCreators,
  users,
  selectedCreatorIds,
  onCreatorSelectionChange,
  domains,
  selectedDomainIds,
  onDomainSelectionChange,
  hasActiveFilters,
  onClearAllFilters,
}) => {
  const [isOpen, setIsOpen] = useState(false);
  const popoverRef = useRef<HTMLDivElement>(null);

  const close = useCallback(() => setIsOpen(false), []);
  useClickOutside(popoverRef, isOpen, close);

  const activeCount = selectedCreatorIds.length + selectedDomainIds.length;

  return (
    <div className="filter-popover" ref={popoverRef}>
      <Indicator
        label={activeCount > 0 ? activeCount : undefined}
        disabled={activeCount === 0}
        size={16}
        offset={2}
        color="blue"
      >
        <ActionIcon
          variant={hasActiveFilters ? 'light' : 'subtle'}
          color={hasActiveFilters ? 'blue' : 'gray'}
          size="sm"
          onClick={() => setIsOpen(!isOpen)}
          aria-expanded={isOpen}
          aria-haspopup="true"
          aria-label="Toggle filters"
        >
          {FILTER_ICON}
        </ActionIcon>
      </Indicator>

      {isOpen && (
        <div className="filter-popover-panel">
          <div className="filter-popover-header">
            <span className="filter-popover-title">Filters</span>
            {hasActiveFilters && onClearAllFilters && (
              <Button variant="subtle" size="compact-xs" onClick={onClearAllFilters}>
                Clear all
              </Button>
            )}
          </div>
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
        </div>
      )}
    </div>
  );
};
