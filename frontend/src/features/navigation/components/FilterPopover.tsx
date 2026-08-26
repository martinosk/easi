import { ActionIcon, Button, Group, Indicator, Popover } from '@mantine/core';
import React from 'react';
import type { ArtifactCreator } from '../utils/filterByCreator';
import { CreatedByFilter } from './CreatedByFilter';
import { DomainFilter } from './DomainFilter';
import classes from './FilterPopover.module.css';

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
  const activeCount = selectedCreatorIds.length + selectedDomainIds.length;

  return (
    <Popover withinPortal position="bottom-start" offset={4} shadow="md" classNames={{ dropdown: classes.dropdown }}>
      <Indicator
        label={activeCount > 0 ? activeCount : undefined}
        disabled={activeCount === 0}
        size={16}
        offset={2}
        color="blue"
      >
        <Popover.Target>
          <ActionIcon
            variant={hasActiveFilters ? 'light' : 'subtle'}
            color={hasActiveFilters ? 'blue' : 'gray'}
            size="sm"
            aria-label="Toggle filters"
          >
            {FILTER_ICON}
          </ActionIcon>
        </Popover.Target>
      </Indicator>

      <Popover.Dropdown p={0}>
        <Group justify="space-between" className={classes.header}>
          <span className={classes.title}>Filters</span>
          {hasActiveFilters && onClearAllFilters && (
            <Button variant="subtle" size="compact-xs" onClick={onClearAllFilters}>
              Clear all
            </Button>
          )}
        </Group>
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
      </Popover.Dropdown>
    </Popover>
  );
};
