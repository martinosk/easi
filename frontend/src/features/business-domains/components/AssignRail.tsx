import { Stack, Text } from '@mantine/core';
import type { Capability, CapabilityId } from '../../../api/types';
import classes from './AssignRail.module.css';
import { CapabilityExplorer } from './CapabilityExplorer';

export interface AssignRailProps {
  allCapabilities: Capability[];
  isLoading: boolean;
  globalAssignedCapabilityIds: Set<CapabilityId>;
  onDragStart: (capability: Capability) => void;
  onDragEnd: () => void;
}

export function AssignRail({
  allCapabilities,
  isLoading,
  globalAssignedCapabilityIds,
  onDragStart,
  onDragEnd,
}: AssignRailProps) {
  return (
    <Stack gap="md" p="md" className={classes.rail} data-testid="assign-rail">
      <Text size="sm" c="dimmed">
        Drag an L1 capability onto a domain card to assign it
      </Text>
      <CapabilityExplorer
        capabilities={allCapabilities}
        assignedCapabilityIds={globalAssignedCapabilityIds}
        isLoading={isLoading}
        onDragStart={onDragStart}
        onDragEnd={onDragEnd}
      />
    </Stack>
  );
}
