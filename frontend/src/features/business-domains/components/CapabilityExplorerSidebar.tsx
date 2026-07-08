import { Box, Stack, Text } from '@mantine/core';
import type { BusinessDomain, Capability, CapabilityId } from '../../../api/types';
import { CapabilityExplorer } from './CapabilityExplorer';
import classes from './CapabilityExplorerSidebar.module.css';

interface CapabilityExplorerSidebarProps {
  visualizedDomain: BusinessDomain | null;
  capabilities: Capability[];
  assignedCapabilityIds: Set<CapabilityId>;
  isLoading: boolean;
  onDragStart?: (capability: Capability) => void;
  onDragEnd?: () => void;
}

export function CapabilityExplorerSidebar({
  visualizedDomain,
  capabilities,
  assignedCapabilityIds,
  isLoading,
  onDragStart,
  onDragEnd,
}: CapabilityExplorerSidebarProps) {
  return (
    <Stack gap="md" p="md" h="100%" className={classes.root}>
      <Text size="sm" c="dimmed">
        {visualizedDomain
          ? 'Drag L1 capabilities to the grid to assign them'
          : 'Select a domain to visualize, then drag capabilities to assign them'}
      </Text>
      <Box flex={1} className={classes.tree}>
        <CapabilityExplorer
          capabilities={capabilities}
          assignedCapabilityIds={assignedCapabilityIds}
          isLoading={isLoading}
          onDragStart={onDragStart}
          onDragEnd={onDragEnd}
        />
      </Box>
    </Stack>
  );
}
