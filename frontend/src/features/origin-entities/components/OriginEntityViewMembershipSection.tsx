import { Button, Group, Stack, Text } from '@mantine/core';
import type React from 'react';
import type { View, ViewOriginEntity } from '../../../api/types';
import { hasLink } from '../../../utils/hateoas';
import { useCurrentView } from '../../views/hooks/useCurrentView';
import { useRemoveOriginEntityFromView } from '../../views/hooks/useViews';

interface OriginEntityViewMembershipSectionProps {
  entityId: string;
}

function removableMembership(view: View, entityId: string): ViewOriginEntity | undefined {
  const membership = view.originEntities.find((item) => item.originEntityId === entityId);
  return membership && hasLink(membership, 'x-remove') ? membership : undefined;
}

export const OriginEntityViewMembershipSection: React.FC<OriginEntityViewMembershipSectionProps> = ({ entityId }) => {
  const { currentView } = useCurrentView();
  const removeFromView = useRemoveOriginEntityFromView();
  if (!currentView) return null;
  if (!removableMembership(currentView, entityId)) return null;

  return (
    <Stack gap="sm" data-testid="view-membership-section">
      <Text size="sm" fw={500}>
        In this view
      </Text>
      <Group justify="flex-start">
        <Button
          variant="default"
          size="xs"
          onClick={() => removeFromView.mutate({ viewId: currentView.id, originEntityId: entityId })}
        >
          Remove from View
        </Button>
      </Group>
    </Stack>
  );
};
