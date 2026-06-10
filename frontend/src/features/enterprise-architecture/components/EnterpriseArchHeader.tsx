import { Button, Group, Stack, Text, Title } from '@mantine/core';
import React from 'react';

interface EnterpriseArchHeaderProps {
  canWrite: boolean;
  onCreateNew: () => void;
  activeTab?: string;
  showTabActions?: boolean;
}

function PlusIcon() {
  return (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
      <path d="M12 5V19M5 12H19" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}

export const EnterpriseArchHeader = React.memo<EnterpriseArchHeaderProps>(
  ({ canWrite, onCreateNew, showTabActions = true }) => {
    return (
      <Group justify="space-between" align="flex-start" mb="xl">
        <Stack gap={4}>
          <Title order={1}>Enterprise Architecture</Title>
          <Text c="dimmed">
            Manage enterprise capabilities and analyze maturity gaps. An enterprise capability&apos;s composition is
            defined by its active direction&apos;s sources.
          </Text>
        </Stack>
        {showTabActions && canWrite && (
          <Group gap="sm">
            <Button leftSection={<PlusIcon />} onClick={onCreateNew} data-testid="create-capability-btn">
              Create Capability
            </Button>
          </Group>
        )}
      </Group>
    );
  },
);

EnterpriseArchHeader.displayName = 'EnterpriseArchHeader';
