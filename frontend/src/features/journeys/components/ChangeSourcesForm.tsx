import { Button, Group, MultiSelect, Stack, Text } from '@mantine/core';
import { useMemo, useState } from 'react';
import type { CapabilityRealization } from '../../../api/types';
import { sourceCardinalityMessage, violatesSourceCardinality } from '../../../lib/schemas/journey';
import { useChangeJourneySourceApplications } from '../hooks/useJourneys';
import type { CapabilityJourney } from '../types';

function sourceOptions(journey: CapabilityJourney, realizations: CapabilityRealization[]) {
  const map = new Map<string, string>();
  for (const app of journey.fromApplications) map.set(app.componentId, app.componentName);
  for (const realization of realizations) {
    map.set(String(realization.componentId), realization.componentName ?? String(realization.componentId));
  }
  map.delete(journey.toApplication.componentId);
  return [...map].map(([value, label]) => ({ value, label }));
}

export function ChangeSourcesForm({
  journey,
  realizations,
  onDone,
}: {
  journey: CapabilityJourney;
  realizations: CapabilityRealization[];
  onDone: () => void;
}) {
  const mutation = useChangeJourneySourceApplications();
  const [ids, setIds] = useState(journey.fromApplications.map((app) => app.componentId));
  const options = useMemo(() => sourceOptions(journey, realizations), [journey, realizations]);

  const valid = !violatesSourceCardinality(journey.kind, ids.length);

  const save = async () => {
    await mutation.mutateAsync({ journey, request: { componentIds: ids } });
    onDone();
  };

  return (
    <Stack gap="md">
      <MultiSelect
        label="Source applications"
        data={options}
        value={ids}
        onChange={setIds}
        searchable
        data-testid="change-sources-input"
      />
      {!valid && (
        <Text size="xs" c="red" data-testid="change-sources-error">
          {sourceCardinalityMessage(journey.kind)}
        </Text>
      )}
      <Group justify="flex-end" gap="sm">
        <Button variant="default" onClick={onDone} disabled={mutation.isPending}>
          Cancel
        </Button>
        <Button onClick={save} disabled={!valid || mutation.isPending} loading={mutation.isPending} data-testid="save-sources-btn">
          Save
        </Button>
      </Group>
    </Stack>
  );
}
