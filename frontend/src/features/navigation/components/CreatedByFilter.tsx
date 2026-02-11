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

  const selectedSet = useMemo(() => new Set(selectedCreatorIds), [selectedCreatorIds]);

  const handleToggle = (creatorId: string) => {
    if (selectedSet.has(creatorId)) {
      onSelectionChange(selectedCreatorIds.filter((id) => id !== creatorId));
    } else {
      onSelectionChange([...selectedCreatorIds, creatorId]);
    }
  };

  const handleClear = () => {
    onSelectionChange([]);
  };

  return (
    <div className="tree-filter">
      <div className="tree-filter-header">
        <span className="tree-filter-label">Created by</span>
        {selectedCreatorIds.length > 0 && (
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
        {creatorOptions.map((option) => (
          <button
            key={option.id}
            className={`tree-filter-option ${selectedSet.has(option.id) ? 'selected' : ''}`}
            onClick={() => handleToggle(option.id)}
          >
            {option.label}
          </button>
        ))}
      </div>
    </div>
  );
};
