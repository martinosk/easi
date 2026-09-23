import { Button, Group, Stack } from '@mantine/core';
import type React from 'react';
import type { View, ViewOriginEntity } from '../../../api/types';
import { hasLink } from '../../../utils/hateoas';
import { useCurrentView } from '../../views/hooks/useCurrentView';
import { useRemoveOriginEntityFromView } from '../../views/hooks/useViews';

export interface OriginEntityViewMembership {
  view: View;
  membership: ViewOriginEntity;
}

function removableMembership(view: View, entityId: string): ViewOriginEntity | undefined {
  const membership = view.originEntities.find((item) => item.originEntityId === entityId);
  return membership && hasLink(membership, 'x-remove') ? membership : undefined;
}

export function useOriginEntityViewMembership(entityId: string): OriginEntityViewMembership | null {
  const { currentView } = useCurrentView();
  if (!currentView) return null;
  const membership = removableMembership(currentView, entityId);
  return membership ? { view: currentView, membership } : null;
}

export const OriginEntityViewMembershipSection: React.FC<OriginEntityViewMembership> = ({ view, membership }) => {
  const removeFromView = useRemoveOriginEntityFromView();

  return (
    <Stack gap="sm" data-testid="view-membership-section">
      <Group justify="flex-start">
        <Button
          variant="default"
          size="xs"
          onClick={() => removeFromView.mutate({ viewId: view.id, originEntityId: membership.originEntityId })}
        >
          Remove from View
        </Button>
      </Group>
    </Stack>
  );
};
