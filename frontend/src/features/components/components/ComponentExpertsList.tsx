import React from 'react';
import { Box, Text, Stack, Button, Group, CloseButton } from '@mantine/core';
import type { Expert, ComponentId } from '../../../api/types';
import { hasLink } from '../../../utils/hateoas';
import { useRemoveComponentExpert } from '../hooks/useComponents';

interface ComponentExpertsListProps {
  componentId: ComponentId;
  experts?: Expert[];
  canAddExpert?: boolean;
  onAddClick: () => void;
  disabled?: boolean;
}

export const ComponentExpertsList: React.FC<ComponentExpertsListProps> = ({
  componentId,
  experts,
  canAddExpert,
  onAddClick,
  disabled,
}) => {
  const removeExpertMutation = useRemoveComponentExpert();

  const handleRemove = (expert: Expert) => {
    removeExpertMutation.mutate({ componentId, expert });
  };

  return (
    <Box>
      <Text size="sm" fw={500} mb="xs">
        Experts
      </Text>
      {experts?.length ? (
        <Stack gap="xs">
          {experts.map((expert, i) => (
            <Group key={i} justify="space-between" wrap="nowrap">
              <Text size="sm" c="dimmed">
                {expert.name} ({expert.role}) - {expert.contact}
              </Text>
              {hasLink(expert, 'x-remove') && (
                <CloseButton
                  size="sm"
                  onClick={() => handleRemove(expert)}
                  disabled={disabled || removeExpertMutation.isPending}
                  data-testid={`remove-expert-${i}`}
                />
              )}
            </Group>
          ))}
        </Stack>
      ) : (
        <Text size="sm" c="dimmed">
          No experts added
        </Text>
      )}
      {canAddExpert && (
        <Button
          variant="subtle"
          size="compact-sm"
          onClick={onAddClick}
          disabled={disabled}
          mt="xs"
          data-testid="add-component-expert-button"
        >
          + Add Expert
        </Button>
      )}
    </Box>
  );
};
