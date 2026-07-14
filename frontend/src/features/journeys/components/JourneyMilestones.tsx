import { Button, Group, Stack, Text } from '@mantine/core';
import { useState } from 'react';
import { hasLink } from '../../../utils/hateoas';
import { useRemoveJourneyMilestone } from '../hooks/useJourneys';
import type { CapabilityJourney, JourneyMilestone } from '../types';
import { milestoneWhenLabel } from '../utils/journeyFormat';
import classes from './JourneyMilestones.module.css';
import { MilestoneDialog } from './MilestoneDialog';

function MilestoneRow({ milestone, onEdit }: { milestone: JourneyMilestone; onEdit: (m: JourneyMilestone) => void }) {
  const removeMutation = useRemoveJourneyMilestone();

  return (
    <div className={classes.row} data-testid={`milestone-row-${milestone.id}`}>
      <span className={classes.dot} data-status={milestone.status} data-testid={`milestone-dot-${milestone.id}`} />
      <Text size="sm" className={classes.label}>
        {milestone.label}
      </Text>
      <Text size="xs" c="dimmed" className={classes.when} data-testid={`milestone-when-${milestone.id}`}>
        {milestoneWhenLabel(milestone)}
      </Text>
      {hasLink(milestone, 'edit') && (
        <Button
          variant="subtle"
          size="compact-xs"
          onClick={() => onEdit(milestone)}
          data-testid={`edit-milestone-btn-${milestone.id}`}
        >
          Edit
        </Button>
      )}
      {hasLink(milestone, 'delete') && (
        <Button
          variant="subtle"
          color="red"
          size="compact-xs"
          loading={removeMutation.isPending}
          onClick={() => removeMutation.mutateAsync(milestone)}
          data-testid={`remove-milestone-btn-${milestone.id}`}
        >
          Remove
        </Button>
      )}
    </div>
  );
}

export function JourneyMilestones({ journey }: { journey: CapabilityJourney }) {
  const [editing, setEditing] = useState<JourneyMilestone | null>(null);
  const [adding, setAdding] = useState(false);

  const closeDialog = () => {
    setEditing(null);
    setAdding(false);
  };

  return (
    <Stack gap={0}>
      {journey.milestones.map((milestone) => (
        <MilestoneRow key={milestone.id} milestone={milestone} onEdit={setEditing} />
      ))}
      {hasLink(journey, 'x-add-milestone') && (
        <Group mt="xs">
          <Button variant="default" size="xs" onClick={() => setAdding(true)} data-testid="add-milestone-btn">
            Add milestone
          </Button>
        </Group>
      )}
      {(adding || editing) && <MilestoneDialog journey={journey} milestone={editing} onClose={closeDialog} />}
    </Stack>
  );
}
