import { Alert, Box, Collapse, Group, Loader, Paper, Stack, Text, UnstyledButton } from '@mantine/core';
import { useDisclosure } from '@mantine/hooks';
import type { AuditEntry } from '../../../api/types';
import { useAuditHistory } from '../hooks/useAuditHistory';

interface AuditHistorySectionProps {
  aggregateId: string;
}

function formatDate(dateString: string): string {
  const date = new Date(dateString);
  return date.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}

function formatEventDataValue(value: unknown): string {
  if (value === null || value === undefined) {
    return '-';
  }
  if (typeof value === 'object') {
    return JSON.stringify(value);
  }
  return String(value);
}

function formatFieldName(key: string): string {
  return key
    .replace(/([a-z])([A-Z])/g, '$1 $2')
    .replace(/_/g, ' ')
    .replace(/^./, (str) => str.toUpperCase());
}

interface ExpandChevronProps {
  expanded: boolean;
}

function ExpandChevron({ expanded }: ExpandChevronProps) {
  return (
    <Text
      component="span"
      c="dimmed"
      aria-hidden
      style={{
        display: 'inline-block',
        transform: expanded ? 'rotate(90deg)' : 'none',
        transition: 'transform 0.2s',
      }}
    >
      ▸
    </Text>
  );
}

interface AuditEntryCardProps {
  entry: AuditEntry;
}

function AuditEntryCard({ entry }: AuditEntryCardProps) {
  const [opened, { toggle }] = useDisclosure(false);
  const eventDataEntries = Object.entries(entry.eventData || {});
  const hasDetails = eventDataEntries.length > 0;

  return (
    <Paper withBorder radius="sm" bg="gray.0">
      <UnstyledButton
        onClick={toggle}
        aria-expanded={opened}
        w="100%"
        px="sm"
        py="xs"
        ta="left"
      >
        <Group justify="space-between" wrap="nowrap" align="flex-start">
          <Stack gap={2}>
            <Text size="sm" fw={500}>
              {entry.displayName}
            </Text>
            <Group gap="xs" wrap="nowrap">
              <Text size="xs" c="gray.6">
                {entry.actorEmail}
              </Text>
              <Text size="xs" c="gray.3">
                •
              </Text>
              <Text size="xs" c="dimmed">
                {formatDate(entry.occurredAt)}
              </Text>
            </Group>
          </Stack>
          {hasDetails && <ExpandChevron expanded={opened} />}
        </Group>
      </UnstyledButton>
      {hasDetails && (
        <Collapse in={opened}>
          <Stack gap="xs" p="sm" bg="white">
            {eventDataEntries.map(([key, value]) => (
              <Group key={key} gap="sm" wrap="nowrap" align="flex-start">
                <Text size="xs" fw={500} c="gray.6" miw={120}>
                  {formatFieldName(key)}
                </Text>
                <Text size="xs" c="gray.8" style={{ wordBreak: 'break-word', flex: 1 }}>
                  {formatEventDataValue(value)}
                </Text>
              </Group>
            ))}
          </Stack>
        </Collapse>
      )}
    </Paper>
  );
}

interface AuditEntriesProps {
  entries: AuditEntry[];
  isLoading: boolean;
  error: unknown;
}

function AuditEntries({ entries, isLoading, error }: AuditEntriesProps) {
  if (isLoading) {
    return (
      <Group justify="center" gap="sm" py="md">
        <Loader size="xs" />
        <Text size="sm" c="dimmed">
          Loading history...
        </Text>
      </Group>
    );
  }
  if (error) {
    return (
      <Alert color="red" variant="light" ta="center">
        Failed to load history
      </Alert>
    );
  }
  if (entries.length === 0) {
    return (
      <Text size="sm" c="dimmed" ta="center" py="md">
        No history available
      </Text>
    );
  }
  return (
    <Stack gap="xs">
      {entries.map((entry) => (
        <AuditEntryCard key={entry.eventId} entry={entry} />
      ))}
    </Stack>
  );
}

export function AuditHistorySection({ aggregateId }: AuditHistorySectionProps) {
  const [opened, { toggle }] = useDisclosure(false);
  const { data, isLoading, error } = useAuditHistory(aggregateId);

  const entries = data?.entries || [];
  const entryCount = entries.length;

  return (
    <Paper withBorder radius="sm" mt="lg" style={{ overflow: 'hidden' }}>
      <UnstyledButton
        onClick={toggle}
        aria-expanded={opened}
        aria-label={`History, ${entryCount} events`}
        w="100%"
        px="md"
        py="sm"
        bg="gray.0"
        ta="left"
      >
        <Group justify="space-between" wrap="nowrap">
          <Text size="sm" fw={600} c="gray.7">
            History{' '}
            {entryCount > 0 && (
              <Text component="span" c="dimmed" fw={400}>
                ({entryCount})
              </Text>
            )}
          </Text>
          <ExpandChevron expanded={opened} />
        </Group>
      </UnstyledButton>
      <Collapse in={opened}>
        <Box p="sm" style={{ borderTop: '1px solid var(--mantine-color-gray-2)' }}>
          <AuditEntries entries={entries} isLoading={isLoading} error={error} />
        </Box>
      </Collapse>
    </Paper>
  );
}
