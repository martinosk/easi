import { Button, Group } from '@mantine/core';
import React from 'react';

interface RealizationActionsProps {
  canEdit: boolean;
  onEditClick: () => void;
}

export const RealizationActions: React.FC<RealizationActionsProps> = ({ canEdit, onEditClick }) => {
  if (!canEdit) return null;

  return (
    <Group gap="sm">
      <Button variant="default" size="xs" onClick={onEditClick}>
        Edit
      </Button>
    </Group>
  );
};
