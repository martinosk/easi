import { Box, Center, Group, Loader, Stack, Text } from '@mantine/core';
import React from 'react';
import type { Capability } from '../../../api/types';
import type { CapabilityLinkStatusResponse, EnterpriseCapability, EnterpriseCapabilityId } from '../types';
import { DomainCapabilityDockPanel } from './DomainCapabilityDockPanel';
import { EnterpriseCapabilitiesEmptyState } from './EnterpriseCapabilitiesEmptyState';
import { EnterpriseCapabilitiesTable } from './EnterpriseCapabilitiesTable';
import { EnterpriseCapabilityDetailPanel } from './EnterpriseCapabilityDetailPanel';
import classes from './EnterpriseArchContent.module.css';

interface EnterpriseArchContentProps {
  isLoading: boolean;
  error: string | null;
  capabilities: EnterpriseCapability[];
  selectedCapability: EnterpriseCapability | null;
  canWrite: boolean;
  onSelect: (capability: EnterpriseCapability) => void;
  onDelete: (capability: EnterpriseCapability) => void;
  onCreateNew: () => void;
  isDockPanelOpen: boolean;
  domainCapabilities: Capability[];
  linkStatuses: Map<string, CapabilityLinkStatusResponse>;
  isLoadingDomainCapabilities: boolean;
  onCloseDockPanel: () => void;
  onLinkCapability: (enterpriseCapabilityId: EnterpriseCapabilityId, domainCapability: Capability) => void;
}

export const EnterpriseArchContent = React.memo<EnterpriseArchContentProps>(
  ({
    isLoading,
    error,
    capabilities,
    selectedCapability,
    canWrite,
    onSelect,
    onDelete,
    onCreateNew,
    isDockPanelOpen,
    domainCapabilities,
    linkStatuses,
    isLoadingDomainCapabilities,
    onCloseDockPanel,
    onLinkCapability,
  }) => {
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

    const hasAnyPanel = !!(selectedCapability || isDockPanelOpen);
    const hasBothPanels = !!(selectedCapability && isDockPanelOpen);

    return (
      <Group align="flex-start" gap="lg" wrap="nowrap" className={classes.layout}>
        <Box
          flex={hasBothPanels ? 1 : hasAnyPanel ? 2 : 1}
          miw={0}
        >
          <EnterpriseCapabilitiesTable
            capabilities={capabilities}
            selectedId={selectedCapability?.id}
            onSelect={onSelect}
            onDelete={onDelete}
            isDockPanelOpen={isDockPanelOpen}
            onLinkCapability={onLinkCapability}
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
        {isDockPanelOpen && (
          <DomainCapabilityDockPanel
            capabilities={domainCapabilities}
            linkStatuses={linkStatuses}
            isLoading={isLoadingDomainCapabilities}
            onClose={onCloseDockPanel}
          />
        )}
      </Group>
    );
  },
);

EnterpriseArchContent.displayName = 'EnterpriseArchContent';
