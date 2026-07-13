import { Button, Group, Stack, Textarea, TextInput } from '@mantine/core';
import { useState } from 'react';
import { useUpdateJourneyDetails } from '../hooks/useJourneys';
import type { CapabilityJourney } from '../types';
import { isPeriodPaired, PeriodFields, type PeriodValue, toTargetPeriod } from './periodFields';

export function EditJourneyDetailsForm({ journey, onDone }: { journey: CapabilityJourney; onDone: () => void }) {
  const mutation = useUpdateJourneyDetails();
  const [note, setNote] = useState(journey.note);
  const [period, setPeriod] = useState<PeriodValue>({
    year: journey.targetPeriod?.year,
    quarter: journey.targetPeriod?.quarter,
  });
  const [resultingName, setResultingName] = useState(journey.move?.resultingName ?? '');

  const valid = isPeriodPaired(period) && (!journey.move || resultingName.trim().length > 0);

  const save = async () => {
    await mutation.mutateAsync({
      journey,
      request: {
        note: note.trim(),
        targetPeriod: toTargetPeriod(period),
        ...(journey.move ? { resultingName: resultingName.trim() } : {}),
      },
    });
    onDone();
  };

  return (
    <Stack gap="md">
      <Textarea
        label="Note"
        maxLength={2000}
        autosize
        minRows={2}
        value={note}
        onChange={(e) => setNote(e.currentTarget.value)}
        data-testid="details-note-input"
      />
      <PeriodFields value={period} onChange={setPeriod} />
      {journey.move && (
        <TextInput
          label="Resulting name"
          withAsterisk
          value={resultingName}
          onChange={(e) => setResultingName(e.currentTarget.value)}
          data-testid="details-resulting-name-input"
        />
      )}
      <Group justify="flex-end" gap="sm">
        <Button variant="default" onClick={onDone} disabled={mutation.isPending}>
          Cancel
        </Button>
        <Button onClick={save} disabled={!valid || mutation.isPending} loading={mutation.isPending} data-testid="save-details-btn">
          Save
        </Button>
      </Group>
    </Stack>
  );
}
