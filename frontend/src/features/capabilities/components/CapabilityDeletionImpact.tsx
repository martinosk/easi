import { Alert, Button, Group, ScrollArea, Skeleton, Stack, Text } from '@mantine/core';
import type { ImpactRealization } from '../../../api/types';
import type { CapabilityDeletionPlan } from '../hooks/useCapabilityDeletionPlan';

export const plural = (count: number, singular: string, pluralForm: string) => (count === 1 ? singular : pluralForm);

export const ImpactLoading = () => (
  <Stack gap="sm" data-testid="delete-impact-loading">
    <Skeleton height={20} width="80%" />
    <Skeleton height={20} width="60%" />
    <Skeleton height={40} />
  </Stack>
);

export const ImpactError = ({ onRetry }: { onRetry: () => void }) => (
  <Alert color="red" data-testid="delete-impact-error">
    <Group justify="space-between">
      <Text size="sm">Could not load what this delete will remove.</Text>
      <Button size="xs" variant="light" color="red" onClick={onRetry} data-testid="delete-impact-retry">
        Retry
      </Button>
    </Group>
  </Alert>
);

interface RealizationsToRemoveProps {
  realizations: ImpactRealization[];
  capabilityNames: Map<string, string>;
}

export const RealizationsToRemove = ({ realizations, capabilityNames }: RealizationsToRemoveProps) => {
  if (realizations.length === 0) return null;

  return (
    <Stack gap="xs" data-testid="realizations-to-remove">
      <Text fw={600} size="sm">
        {realizations.length} application {plural(realizations.length, 'realization', 'realizations')} will be removed
      </Text>
      <ScrollArea.Autosize mah={150}>
        <Stack gap={4}>
          {realizations.map((r) => (
            <Text key={r.id} size="sm" data-testid="realization-to-remove">
              <Text span fw={600}>
                {r.componentName}
              </Text>{' '}
              stops realising{' '}
              <Text span fw={600}>
                {capabilityNames.get(r.capabilityId) ?? r.capabilityName ?? r.capabilityId}
              </Text>
            </Text>
          ))}
        </Stack>
      </ScrollArea.Autosize>
      <Text size="sm" c="dimmed">
        The applications themselves are kept.
      </Text>
    </Stack>
  );
};

export const CapabilityDeletionImpactSection = ({ plan }: { plan: CapabilityDeletionPlan }) => {
  if (plan.isLoading) return <ImpactLoading />;
  if (plan.isError) return <ImpactError onRetry={plan.retry} />;
  return <RealizationsToRemove realizations={plan.realizations} capabilityNames={plan.capabilityNames} />;
};
