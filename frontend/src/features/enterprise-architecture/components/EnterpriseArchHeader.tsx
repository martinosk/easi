import { Button, Group, Stack, Text, Title } from '@mantine/core';
import React from 'react';

interface EnterpriseArchHeaderProps {
  canWrite: boolean;
  onCreateNew: () => void;
  isDockPanelOpen: boolean;
  onToggleDockPanel: () => void;
  activeTab?: string;
  showTabActions?: boolean;
}

function LinkIcon() {
  return (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
      <path
        d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path
        d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}

function PlusIcon() {
  return (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
      <path d="M12 5V19M5 12H19" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}

export const EnterpriseArchHeader = React.memo<EnterpriseArchHeaderProps>(
  ({ canWrite, onCreateNew, isDockPanelOpen, onToggleDockPanel, showTabActions = true }) => {
    return (
      <Group justify="space-between" align="flex-start" mb="xl">
        <Stack gap={4}>
          <Title order={1}>Enterprise Architecture</Title>
          <Text c="dimmed">
            Manage enterprise capabilities, analyze maturity gaps, and discover unlinked domain capabilities.
          </Text>
        </Stack>
        {showTabActions && (
          <Group gap="sm">
            <Button
              variant={isDockPanelOpen ? 'filled' : 'default'}
              leftSection={<LinkIcon />}
              onClick={onToggleDockPanel}
              data-testid="toggle-dock-panel-btn"
              aria-pressed={isDockPanelOpen}
            >
              {isDockPanelOpen ? 'Hide Linking Panel' : 'Link Capabilities'}
            </Button>
            {canWrite && (
              <Button leftSection={<PlusIcon />} onClick={onCreateNew} data-testid="create-capability-btn">
                Create Capability
              </Button>
            )}
          </Group>
        )}
      </Group>
    );
  },
);

EnterpriseArchHeader.displayName = 'EnterpriseArchHeader';
