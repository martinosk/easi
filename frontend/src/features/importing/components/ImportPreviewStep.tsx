import { Alert, Button, Group, List, Paper, Stack, Text, Title } from '@mantine/core';
import type { ImportPreview } from '../types';

interface ImportPreviewStepProps {
  preview: ImportPreview;
  eaOwnerName?: string;
  onConfirm: () => void;
  onCancel: () => void;
  isLoading: boolean;
}

interface SupportedSummaryProps {
  supported: ImportPreview['supported'];
}

function SupportedSummary({ supported }: SupportedSummaryProps) {
  return (
    <Paper p="md" withBorder radius="md">
      <Stack gap="sm">
        <Title order={5}>Will Import</Title>
        <List data-testid="supported-list">
          <List.Item>
            <Text component="span" fw={700}>
              {supported.capabilities}
            </Text>{' '}
            Capabilities
          </List.Item>
          <List.Item>
            <Text component="span" fw={700}>
              {supported.components}
            </Text>{' '}
            Components
          </List.Item>
          <List.Item>
            <Text component="span" fw={700}>
              {supported.valueStreams}
            </Text>{' '}
            Value streams
          </List.Item>
          <List.Item>
            <Text component="span" fw={700}>
              {supported.parentChildRelationships}
            </Text>{' '}
            Parent-child relationships
          </List.Item>
          <List.Item>
            <Text component="span" fw={700}>
              {supported.realizations}
            </Text>{' '}
            Capability realizations
          </List.Item>
          <List.Item>
            <Text component="span" fw={700}>
              {supported.componentRelationships}
            </Text>{' '}
            Component relationships
          </List.Item>
          <List.Item>
            <Text component="span" fw={700}>
              {supported.capabilityToValueStreamMappings}
            </Text>{' '}
            Capability-to-value-stream mappings
          </List.Item>
        </List>
      </Stack>
    </Paper>
  );
}

interface SettingsSummaryProps {
  eaOwnerName: string;
}

function SettingsSummary({ eaOwnerName }: SettingsSummaryProps) {
  return (
    <Paper p="md" withBorder radius="md" data-testid="import-settings">
      <Stack gap="sm">
        <Title order={5}>Import Settings</Title>
        <List>
          <List.Item>
            EA Owner for capabilities:{' '}
            <Text component="span" fw={700}>
              {eaOwnerName}
            </Text>
          </List.Item>
        </List>
      </Stack>
    </Paper>
  );
}

interface UnsupportedGroupProps {
  heading: string;
  items: Record<string, number>;
  testId: string;
}

function UnsupportedGroup({ heading, items, testId }: UnsupportedGroupProps) {
  if (Object.keys(items).length === 0) return null;
  return (
    <Stack gap="xs">
      <Text fw={600} size="sm">
        {heading}
      </Text>
      <List data-testid={testId}>
        {Object.entries(items).map(([type, count]) => (
          <List.Item key={type}>
            <Text component="span" fw={700}>
              {count}
            </Text>{' '}
            {type}
          </List.Item>
        ))}
      </List>
    </Stack>
  );
}

interface UnsupportedSummaryProps {
  unsupported: ImportPreview['unsupported'];
}

function UnsupportedSummary({ unsupported }: UnsupportedSummaryProps) {
  return (
    <Alert color="yellow" title="Will NOT Import">
      <Stack gap="sm">
        <Text size="sm">The following unsupported elements will be skipped:</Text>
        <UnsupportedGroup heading="Elements:" items={unsupported.elements} testId="unsupported-elements" />
        <UnsupportedGroup
          heading="Relationships:"
          items={unsupported.relationships}
          testId="unsupported-relationships"
        />
      </Stack>
    </Alert>
  );
}

export function ImportPreviewStep({
  preview,
  eaOwnerName,
  onConfirm,
  onCancel,
  isLoading,
}: ImportPreviewStepProps) {
  const { supported, unsupported } = preview;
  const hasUnsupported =
    Object.keys(unsupported.elements).length > 0 || Object.keys(unsupported.relationships).length > 0;

  return (
    <Stack gap="md">
      <Text c="dimmed" size="sm">
        Review what will be imported from the file.
      </Text>

      <SupportedSummary supported={supported} />
      {eaOwnerName && <SettingsSummary eaOwnerName={eaOwnerName} />}
      {hasUnsupported && <UnsupportedSummary unsupported={unsupported} />}

      <Group justify="flex-end" gap="sm">
        <Button variant="default" onClick={onCancel} disabled={isLoading} data-testid="cancel-button">
          Cancel
        </Button>
        <Button onClick={onConfirm} loading={isLoading} data-testid="confirm-button">
          Confirm Import
        </Button>
      </Group>
    </Stack>
  );
}
