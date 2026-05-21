import { Button, Center, Paper, Stack, Text, Title } from '@mantine/core';
import React from 'react';

interface EnterpriseCapabilitiesEmptyStateProps {
  onCreateNew: () => void;
  canWrite?: boolean;
}

function EmptyIcon() {
  return (
    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg" width="48" height="48" aria-hidden="true">
      <path
        d="M19 3H5C3.89543 3 3 3.89543 3 5V19C3 20.1046 3.89543 21 5 21H19C20.1046 21 21 20.1046 21 19V5C21 3.89543 20.1046 3 19 3Z"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path d="M3 9H21" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
      <path d="M9 21V9" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}

export const EnterpriseCapabilitiesEmptyState = React.memo<EnterpriseCapabilitiesEmptyStateProps>(
  ({ onCreateNew, canWrite = false }) => {
    return (
      <Paper withBorder radius="lg" p="xl" data-testid="empty-state">
        <Center>
          <Stack gap="md" align="center" maw={420}>
            <Text c="gray.4">
              <EmptyIcon />
            </Text>
            <Title order={3}>{canWrite ? 'No Enterprise Capabilities Yet' : 'No Enterprise Capabilities'}</Title>
            <Text size="sm" c="dimmed" ta="center">
              {canWrite
                ? 'Enterprise capabilities help you group related domain capabilities across your organization. Create your first enterprise capability to start organizing your architecture.'
                : 'No enterprise capabilities have been created yet.'}
            </Text>
            {canWrite && (
              <Button onClick={onCreateNew} data-testid="create-first-capability-btn">
                Create Enterprise Capability
              </Button>
            )}
          </Stack>
        </Center>
      </Paper>
    );
  },
);

EnterpriseCapabilitiesEmptyState.displayName = 'EnterpriseCapabilitiesEmptyState';
