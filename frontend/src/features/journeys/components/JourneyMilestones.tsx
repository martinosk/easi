import { ActionIcon, Button, Group, Stack, Text, ThemeIcon, Tooltip } from '@mantine/core';
import { IconCalendarExclamation, IconGripVertical } from '@tabler/icons-react';
import { useState } from 'react';
import { hasLink } from '../../../utils/hateoas';
import { useRemoveJourneyMilestone } from '../hooks/useJourneys';
import { type MilestoneReorderControls, useMilestoneReorder } from '../hooks/useMilestoneReorder';
import type { CapabilityJourney, JourneyMilestone, TargetPeriod } from '../types';
import { milestoneWhenLabel, scheduleConflictLabel } from '../utils/journeyFormat';
import { milestoneScheduleConflicts } from '../utils/milestoneSchedule';
import classes from './JourneyMilestones.module.css';
import { MilestoneDialog } from './MilestoneDialog';

interface MilestoneRowProps {
  milestone: JourneyMilestone;
  index: number;
  reorder: MilestoneReorderControls | null;
  conflictsWith: TargetPeriod | undefined;
  onEdit: (m: JourneyMilestone) => void;
}

function ScheduleConflictMarker({
  milestone,
  latestAbove,
}: {
  milestone: JourneyMilestone;
  latestAbove: TargetPeriod;
}) {
  const label = scheduleConflictLabel(milestone, latestAbove);
  return (
    <Tooltip label={label} withArrow events={{ hover: true, focus: true, touch: false }}>
      <ThemeIcon
        variant="transparent"
        color="orange"
        size="xs"
        tabIndex={0}
        role="img"
        aria-label={label}
        data-testid={`milestone-schedule-conflict-${milestone.id}`}
      >
        <IconCalendarExclamation size={14} />
      </ThemeIcon>
    </Tooltip>
  );
}

function ReorderHandle({ milestone, index, reorder }: Pick<MilestoneRowProps, 'milestone' | 'index' | 'reorder'>) {
  if (!reorder) return null;
  return (
    <ActionIcon
      variant="subtle"
      color="gray"
      size="sm"
      className={classes.handle}
      aria-label={`Reorder ${milestone.label}`}
      onKeyDown={reorder.handleKeyDown(index)}
      {...reorder.dragHandleProps(index)}
      data-testid={`milestone-handle-${milestone.id}`}
    >
      <IconGripVertical size={14} />
    </ActionIcon>
  );
}

function MilestoneRow({ milestone, index, reorder, conflictsWith, onEdit }: MilestoneRowProps) {
  const removeMutation = useRemoveJourneyMilestone();

  return (
    <div
      className={classes.row}
      data-testid={`milestone-row-${milestone.id}`}
      data-reorderable={reorder ? 'true' : undefined}
      data-drag-over={reorder?.overIndex === index ? 'true' : undefined}
      {...reorder?.dropTargetProps(index)}
    >
      <ReorderHandle milestone={milestone} index={index} reorder={reorder} />
      <Text size="xs" c="dimmed" ff="monospace" className={classes.seq} data-testid={`milestone-seq-${milestone.id}`}>
        {index + 1}
      </Text>
      <span className={classes.dot} data-status={milestone.status} data-testid={`milestone-dot-${milestone.id}`} />
      <Text size="sm" className={classes.label}>
        {milestone.label}
      </Text>
      <Text size="xs" c="dimmed" className={classes.when} data-testid={`milestone-when-${milestone.id}`}>
        {milestoneWhenLabel(milestone)}
      </Text>
      {conflictsWith && <ScheduleConflictMarker milestone={milestone} latestAbove={conflictsWith} />}
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
  const reorder = useMilestoneReorder(journey);
  const conflicts = milestoneScheduleConflicts(journey.milestones);

  const closeDialog = () => {
    setEditing(null);
    setAdding(false);
  };

  return (
    <Stack gap={0}>
      {journey.milestones.map((milestone, index) => (
        <MilestoneRow
          key={milestone.id}
          milestone={milestone}
          index={index}
          reorder={reorder}
          conflictsWith={conflicts.get(milestone.id)}
          onEdit={setEditing}
        />
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
