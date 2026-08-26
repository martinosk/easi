import { ActionIcon, Box, Center, Group, Paper, Text, Title } from '@mantine/core';
import type { StageCapabilityMapping, ValueStreamStage } from '../../../api/types';
import { hasLink } from '../../../utils/hateoas';
import { CapabilityChip } from './CapabilityChip';
import classes from './StageColumn.module.css';

interface StageColumnProps {
  stage: ValueStreamStage;
  capabilities: StageCapabilityMapping[];
  canWrite: boolean;
  onEdit: (stage: ValueStreamStage) => void;
  onDelete: (stage: ValueStreamStage) => void;
  onRemoveCapability: (mapping: StageCapabilityMapping) => void;
  onDragStart: (e: React.DragEvent, stageId: string) => void;
  onDragOver: (e: React.DragEvent) => void;
  onDrop: (e: React.DragEvent, stageId: string) => void;
}

interface StageActionsProps {
  stage: ValueStreamStage;
  onEdit: (stage: ValueStreamStage) => void;
  onDelete: (stage: ValueStreamStage) => void;
}

function StageActions({ stage, onEdit, onDelete }: StageActionsProps) {
  return (
    <Box className={classes.actions}>
      {hasLink(stage, 'edit') && (
        <ActionIcon
          variant="subtle"
          color="gray"
          size="sm"
          onClick={() => onEdit(stage)}
          title="Edit stage"
          aria-label="Edit stage"
        >
          <svg viewBox="0 0 24 24" fill="none" width="14" height="14">
            <path
              d="M11 4H4a2 2 0 00-2 2v14a2 2 0 002 2h14a2 2 0 002-2v-7"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
            <path
              d="M18.5 2.5a2.121 2.121 0 013 3L12 15l-4 1 1-4 9.5-9.5z"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
          </svg>
        </ActionIcon>
      )}
      {hasLink(stage, 'delete') && (
        <ActionIcon
          variant="subtle"
          color="red"
          size="sm"
          onClick={() => onDelete(stage)}
          title="Delete stage"
          aria-label="Delete stage"
        >
          <svg viewBox="0 0 24 24" fill="none" width="14" height="14">
            <path
              d="M3 6h18M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a2 2 0 012-2h4a2 2 0 012 2v2"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
          </svg>
        </ActionIcon>
      )}
    </Box>
  );
}

interface StageCapabilitiesProps {
  capabilities: StageCapabilityMapping[];
  canWrite: boolean;
  onRemoveCapability: (mapping: StageCapabilityMapping) => void;
}

function StageCapabilities({ capabilities, canWrite, onRemoveCapability }: StageCapabilitiesProps) {
  return (
    <Group gap="xs" align="flex-start" pt="sm" className={classes.capabilities}>
      {capabilities.length > 0 ? (
        capabilities.map((cap) => (
          <CapabilityChip
            key={`${cap.stageId}-${cap.capabilityId}`}
            mapping={cap}
            canRemove={canWrite}
            onRemove={onRemoveCapability}
          />
        ))
      ) : (
        <Text size="xs" fs="italic" c="dimmed">
          No capabilities mapped
        </Text>
      )}
    </Group>
  );
}

export function StageColumn({
  stage,
  capabilities,
  canWrite,
  onEdit,
  onDelete,
  onRemoveCapability,
  onDragStart,
  onDragOver,
  onDrop,
}: StageColumnProps) {
  return (
    <Paper
      component="section"
      radius="lg"
      shadow="sm"
      p="md"
      className={classes.column}
      data-testid={`stage-${stage.id}`}
      aria-label={stage.name}
      draggable={canWrite}
      onDragStart={(e) => onDragStart(e, stage.id)}
      onDragOver={onDragOver}
      onDrop={(e) => onDrop(e, stage.id)}
    >
      <Group gap="sm" align="flex-start" wrap="nowrap" mb="sm">
        <Center fz="xs" fw={700} className={classes.position}>
          {stage.position}
        </Center>
        <Title order={3} fz="md" flex={1} className={classes.name}>
          {stage.name}
        </Title>
        {canWrite && <StageActions stage={stage} onEdit={onEdit} onDelete={onDelete} />}
      </Group>
      {stage.description && (
        <Text size="xs" c="dimmed" mb="sm">
          {stage.description}
        </Text>
      )}
      <StageCapabilities capabilities={capabilities} canWrite={canWrite} onRemoveCapability={onRemoveCapability} />
    </Paper>
  );
}
