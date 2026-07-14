import { Button, Group, NumberInput } from '@mantine/core';
import { useState } from 'react';
import { useUpdateJourneyProgress } from '../hooks/useJourneys';
import type { CapabilityJourney } from '../types';

function clampProgress(raw: number | string): number {
  const parsed = typeof raw === 'number' ? raw : Number(raw);
  if (!Number.isFinite(parsed)) return 0;
  return Math.min(100, Math.max(0, Math.round(parsed)));
}

export function ProgressEditor({ journey, onClose }: { journey: CapabilityJourney; onClose: () => void }) {
  const [value, setValue] = useState<number>(journey.progress ?? 0);
  const mutation = useUpdateJourneyProgress();

  const save = async () => {
    await mutation.mutateAsync({ journey, request: { progress: value } });
    onClose();
  };

  return (
    <Group gap="xs" align="flex-end" data-testid="progress-editor">
      <NumberInput
        label="Progress (%)"
        min={0}
        max={100}
        value={value}
        onChange={(v) => setValue(clampProgress(v))}
        data-testid="progress-input"
      />
      <Button size="xs" onClick={save} loading={mutation.isPending} data-testid="save-progress-btn">
        Save
      </Button>
      <Button size="xs" variant="default" onClick={onClose} disabled={mutation.isPending}>
        Cancel
      </Button>
    </Group>
  );
}
