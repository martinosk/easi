import { CloseButton, Group, Paper, Stack, Title } from '@mantine/core';
import React from 'react';
import type { Capability } from '../../../api/types';
import type { CapabilityLinkStatusResponse } from '../types';
import { DomainCapabilityPanel } from './DomainCapabilityPanel';
import classes from './DomainCapabilityDockPanel.module.css';

interface DomainCapabilityDockPanelProps {
  capabilities: Capability[];
  linkStatuses: Map<string, CapabilityLinkStatusResponse>;
  isLoading: boolean;
  onClose: () => void;
}

export const DomainCapabilityDockPanel = React.memo<DomainCapabilityDockPanelProps>(
  ({ capabilities, linkStatuses, isLoading, onClose }) => {
    return (
      <Paper withBorder radius="lg" shadow="sm" className={classes.panel}>
        <Stack gap={0} h="100%">
          <Group justify="space-between" px="md" py="sm" className={classes.header}>
            <Title order={4}>Link Capabilities</Title>
            <CloseButton onClick={onClose} aria-label="Close dock panel" />
          </Group>
          <DomainCapabilityPanel capabilities={capabilities} linkStatuses={linkStatuses} isLoading={isLoading} />
        </Stack>
      </Paper>
    );
  },
);

DomainCapabilityDockPanel.displayName = 'DomainCapabilityDockPanel';
