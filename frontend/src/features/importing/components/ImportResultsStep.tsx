import { Alert, Button, Group, List, Paper, Stack, Text, Title } from '@mantine/core';
import type { ImportError, ImportResult } from '../types';

interface ImportResultsStepProps {
  result: ImportResult;
  onClose: () => void;
}

interface SummaryProps {
  result: ImportResult;
}

function ResultsSummary({ result }: SummaryProps) {
  const { capabilitiesCreated, componentsCreated, valueStreamsCreated, realizationsCreated, capabilityMappings, domainAssignments } =
    result;
  return (
    <Paper p="md" withBorder radius="md">
      <Stack gap="sm">
        <Title order={5}>Summary</Title>
        <List data-testid="results-summary">
          <List.Item>
            <Text component="span" fw={700}>
              {capabilitiesCreated}
            </Text>{' '}
            Capabilities created
          </List.Item>
          <List.Item>
            <Text component="span" fw={700}>
              {componentsCreated}
            </Text>{' '}
            Components created
          </List.Item>
          <List.Item>
            <Text component="span" fw={700}>
              {valueStreamsCreated}
            </Text>{' '}
            Value streams created
          </List.Item>
          <List.Item>
            <Text component="span" fw={700}>
              {realizationsCreated}
            </Text>{' '}
            Realizations created
          </List.Item>
          {capabilityMappings > 0 && (
            <List.Item>
              <Text component="span" fw={700}>
                {capabilityMappings}
              </Text>{' '}
              Capability-to-value-stream mappings
            </List.Item>
          )}
          {domainAssignments > 0 && (
            <List.Item>
              <Text component="span" fw={700}>
                {domainAssignments}
              </Text>{' '}
              Domain assignments made
            </List.Item>
          )}
        </List>
      </Stack>
    </Paper>
  );
}

interface ErrorRowProps {
  error: ImportError;
}

function ErrorRow({ error }: ErrorRowProps) {
  return (
    <Paper p="sm" withBorder radius="sm" bg="red.0">
      <Stack gap={4}>
        <Text fw={600} size="sm">
          {error.sourceName} ({error.sourceElement})
        </Text>
        <Text size="sm" c="red">
          {error.error}
        </Text>
        <Text size="xs" c="dimmed">
          Action: {error.action}
        </Text>
      </Stack>
    </Paper>
  );
}

interface ErrorsSectionProps {
  errors: ImportError[];
}

function ErrorsSection({ errors }: ErrorsSectionProps) {
  return (
    <Alert color="red" title={`Errors (${errors.length})`}>
      <Stack gap="xs" data-testid="error-list">
        {errors.map((error) => (
          <ErrorRow key={`${error.sourceElement}-${error.sourceName}-${error.error}`} error={error} />
        ))}
      </Stack>
    </Alert>
  );
}

export function ImportResultsStep({ result, onClose }: ImportResultsStepProps) {
  const hasErrors = result.errors.length > 0;

  return (
    <Stack gap="md">
      <Title order={4}>{hasErrors ? 'Import Completed with Errors' : 'Import Complete'}</Title>

      <ResultsSummary result={result} />
      {hasErrors && <ErrorsSection errors={result.errors} />}

      <Group justify="flex-end" gap="sm">
        <Button onClick={onClose} data-testid="close-button">
          Close
        </Button>
      </Group>
    </Stack>
  );
}
