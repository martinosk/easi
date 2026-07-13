import { Button, Group, Modal, SegmentedControl, Stack, TextInput } from '@mantine/core';
import { useState } from 'react';
import { useAddJourneyMilestone, useUpdateJourneyMilestone } from '../hooks/useJourneys';
import type { CapabilityJourney, JourneyMilestone, MilestoneStatus, UpdateJourneyMilestoneRequest } from '../types';
import { isPeriodPaired, PeriodFields, type PeriodValue, toTargetPeriod } from './periodFields';

const STATUS_OPTIONS = [
  { value: 'planned', label: 'Planned' },
  { value: 'in-flight', label: 'In flight' },
  { value: 'done', label: 'Done' },
];

function initialPeriod(milestone: JourneyMilestone | null): PeriodValue {
  return { year: milestone?.targetPeriod?.year, quarter: milestone?.targetPeriod?.quarter };
}

function useMilestoneSave(journey: CapabilityJourney, milestone: JourneyMilestone | null, onClose: () => void) {
  const addMutation = useAddJourneyMilestone();
  const updateMutation = useUpdateJourneyMilestone();

  const save = async (request: UpdateJourneyMilestoneRequest) => {
    if (milestone) await updateMutation.mutateAsync({ milestone, request });
    else await addMutation.mutateAsync({ journey, request });
    onClose();
  };

  return { save, isPending: addMutation.isPending || updateMutation.isPending };
}

interface MilestoneDialogProps {
  journey: CapabilityJourney;
  milestone: JourneyMilestone | null;
  onClose: () => void;
}

export function MilestoneDialog({ journey, milestone, onClose }: MilestoneDialogProps) {
  const [label, setLabel] = useState(milestone?.label ?? '');
  const [period, setPeriod] = useState(initialPeriod(milestone));
  const [status, setStatus] = useState<MilestoneStatus>(milestone?.status ?? 'planned');
  const { save, isPending } = useMilestoneSave(journey, milestone, onClose);

  const valid = label.trim().length > 0 && isPeriodPaired(period);
  const submit = () => save({ label: label.trim(), targetPeriod: toTargetPeriod(period), status });

  return (
    <Modal opened onClose={onClose} title={milestone ? 'Edit milestone' : 'Add milestone'} data-testid="milestone-dialog">
      <Stack gap="md">
        <TextInput
          label="Label"
          withAsterisk
          maxLength={200}
          value={label}
          onChange={(e) => setLabel(e.currentTarget.value)}
          data-testid="milestone-label-input"
        />
        <PeriodFields value={period} onChange={setPeriod} />
        <SegmentedControl
          data={STATUS_OPTIONS}
          value={status}
          onChange={(value) => setStatus(value as MilestoneStatus)}
          data-testid="milestone-status"
        />
        <Group justify="flex-end" gap="sm">
          <Button variant="default" onClick={onClose} disabled={isPending}>
            Cancel
          </Button>
          <Button onClick={submit} disabled={!valid || isPending} loading={isPending} data-testid="save-milestone-btn">
            Save
          </Button>
        </Group>
      </Stack>
    </Modal>
  );
}
