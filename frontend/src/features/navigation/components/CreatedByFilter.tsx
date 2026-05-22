import { Chip, Group, UnstyledButton } from '@mantine/core';
import React, { useMemo } from 'react';
import type { ArtifactCreator } from '../utils/filterByCreator';

interface CreatedByFilterProps {
  artifactCreators: ArtifactCreator[];
  users: Array<{ id: string; name?: string; email: string }>;
  selectedCreatorIds: string[];
  onSelectionChange: (creatorIds: string[]) => void;
}

export const CreatedByFilter: React.FC<CreatedByFilterProps> = ({
  artifactCreators,
  users,
  selectedCreatorIds,
  onSelectionChange,
}) => {
  const creatorOptions = useMemo(() => {
    const uniqueCreatorIds = [...new Set(artifactCreators.map((ac) => ac.creatorId))];
    const userMap = new Map(users.map((u) => [u.id, u]));

    return uniqueCreatorIds.map((creatorId) => {
      const user = userMap.get(creatorId);
      return {
        id: creatorId,
        label: user?.name || user?.email || creatorId,
      };
    });
  }, [artifactCreators, users]);

  const handleClear = () => onSelectionChange([]);

  return (
    <div className="tree-filter">
      <div className="tree-filter-header">
        <span className="tree-filter-label">Created by</span>
        {selectedCreatorIds.length > 0 && (
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
      <Chip.Group multiple value={selectedCreatorIds} onChange={onSelectionChange}>
        <Group gap={4}>
          {creatorOptions.map((option) => (
            <Chip key={option.id} value={option.id} size="xs">
              {option.label}
            </Chip>
          ))}
        </Group>
      </Chip.Group>
    </div>
  );
};
