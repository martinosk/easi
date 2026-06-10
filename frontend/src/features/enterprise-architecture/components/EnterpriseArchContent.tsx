import { Box, Center, Group, Loader, Stack, Text } from '@mantine/core';
import React from 'react';
import type { EnterpriseCapability } from '../types';
import classes from './EnterpriseArchContent.module.css';
import { EnterpriseCapabilitiesEmptyState } from './EnterpriseCapabilitiesEmptyState';
import { EnterpriseCapabilitiesTable } from './EnterpriseCapabilitiesTable';
import { EnterpriseCapabilityDetailPanel } from './EnterpriseCapabilityDetailPanel';

interface EnterpriseArchContentProps {
  isLoading: boolean;
  error: string | null;
  capabilities: EnterpriseCapability[];
  selectedCapability: EnterpriseCapability | null;
  canWrite: boolean;
  onSelect: (capability: EnterpriseCapability) => void;
  onDelete: (capability: EnterpriseCapability) => void;
  onCreateNew: () => void;
}

export const EnterpriseArchContent = React.memo<EnterpriseArchContentProps>(
  ({ isLoading, error, capabilities, selectedCapability, canWrite, onSelect, onDelete, onCreateNew }) => {
    if (isLoading) {
      return (
        <Center py="xl">
          <Stack align="center" gap="sm">
            <Loader />
            <Text c="dimmed">Loading enterprise capabilities...</Text>
          </Stack>
        </Center>
      );
    }

    if (error) {
      return (
        <Text c="red" data-testid="capabilities-error">
          {error}
        </Text>
      );
    }

    if (capabilities.length === 0) {
      return <EnterpriseCapabilitiesEmptyState onCreateNew={onCreateNew} canWrite={canWrite} />;
    }

    return (
      <Group align="flex-start" gap="lg" wrap="nowrap" className={classes.layout}>
        <Box flex={selectedCapability ? 2 : 1} miw={0}>
          <EnterpriseCapabilitiesTable
            capabilities={capabilities}
            selectedId={selectedCapability?.id}
            onSelect={onSelect}
            onDelete={onDelete}
          />
        </Box>
        {selectedCapability && (
          <Box flex={1} miw={0}>
            <EnterpriseCapabilityDetailPanel
              capability={selectedCapability}
              onClose={() => onSelect(selectedCapability)}
            />
          </Box>
        )}
      </Group>
    );
  },
);

EnterpriseArchContent.displayName = 'EnterpriseArchContent';
