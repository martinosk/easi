import { Chip, Group } from '@mantine/core';
import React, { useMemo } from 'react';
import type { ArtifactCreator } from '../utils/filterByCreator';
import { TreeFilterSection } from './shared/TreeFilterSection';

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

  return (
    <TreeFilterSection
      label="Created by"
      hasSelection={selectedCreatorIds.length > 0}
      onClear={() => onSelectionChange([])}
    >
      <Chip.Group multiple value={selectedCreatorIds} onChange={onSelectionChange}>
        <Group gap={4}>
          {creatorOptions.map((option) => (
            <Chip key={option.id} value={option.id} size="xs">
              {option.label}
            </Chip>
          ))}
        </Group>
      </Chip.Group>
    </TreeFilterSection>
  );
};
