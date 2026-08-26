import { ActionIcon, Box, Group, Paper, Stack, Text, Title } from '@mantine/core';
import { useCallback, useState } from 'react';
import type { StageCapabilityMapping, ValueStreamStage } from '../../../api/types';
import { useRemoveStageCapability } from '../hooks/useValueStreamStages';
import { AddStageButton } from './AddStageButton';
import { StageColumn } from './StageColumn';
import classes from './StageFlowDiagram.module.css';

interface StageFlowDiagramProps {
  valueStreamId: string;
  stages: ValueStreamStage[];
  stageCapabilities: StageCapabilityMapping[];
  canWrite: boolean;
  onAddStage: (position?: number) => void;
  onEditStage: (stage: ValueStreamStage) => void;
  onDeleteStage: (stage: ValueStreamStage) => void;
  onReorder: (orderedStageIds: string[]) => void;
  onAddCapability?: (stageId: string, capabilityId: string) => void;
}

function tryHandleCapabilityDrop(
  e: React.DragEvent,
  targetStageId: string,
  onAddCapability?: (stageId: string, capabilityId: string) => void,
): boolean {
  const json = e.dataTransfer.getData('application/json');
  if (!json || !onAddCapability) return false;
  try {
    const capability = JSON.parse(json);
    if (capability?.id) {
      onAddCapability(targetStageId, capability.id);
      return true;
    }
  } catch {
    /* not a capability drop */
  }
  return false;
}

function reorderStageIds(sortedStages: ValueStreamStage[], draggedId: string, targetId: string): string[] | null {
  const ordered: string[] = sortedStages.map((s) => s.id);
  const fromIndex = ordered.indexOf(draggedId);
  const toIndex = ordered.indexOf(targetId);
  if (fromIndex < 0 || toIndex < 0) return null;
  ordered.splice(fromIndex, 1);
  ordered.splice(toIndex, 0, draggedId);
  return ordered;
}

function applyStageReorder(
  draggedStageId: string | null,
  targetStageId: string,
  sortedStages: ValueStreamStage[],
  onReorder: (orderedStageIds: string[]) => void,
): void {
  if (!draggedStageId || draggedStageId === targetStageId) return;
  const ordered = reorderStageIds(sortedStages, draggedStageId, targetStageId);
  if (ordered) onReorder(ordered);
}

function groupCapabilitiesByStage(stageCapabilities: StageCapabilityMapping[]) {
  const capsByStage = new Map<string, StageCapabilityMapping[]>();
  for (const cap of stageCapabilities) {
    const list = capsByStage.get(cap.stageId) || [];
    list.push(cap);
    capsByStage.set(cap.stageId, list);
  }
  return capsByStage;
}

interface StageConnectorProps {
  canWrite: boolean;
  position: number;
  onInsert: (position: number) => void;
  index: number;
}

function StageConnector({ canWrite, position, onInsert, index }: StageConnectorProps) {
  return (
    <Stack gap={0} align="center" className={classes.connectorGroup}>
      <Box className={classes.connector} />
      {canWrite && (
        <ActionIcon
          variant="filled"
          size="sm"
          className={classes.insertButton}
          data-testid={`insert-stage-btn-${index}`}
          onClick={() => onInsert(position)}
          title="Insert stage here"
          aria-label="Insert stage"
        >
          <svg viewBox="0 0 24 24" fill="none" width="14" height="14">
            <path
              d="M12 5V19M5 12H19"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
          </svg>
        </ActionIcon>
      )}
    </Stack>
  );
}

function EmptyStages({ canWrite, onAddStage }: { canWrite: boolean; onAddStage: (position?: number) => void }) {
  return (
    <Paper radius="lg" shadow="sm" px="xl" py="xxl" data-testid="empty-stages">
      <Stack align="center" gap="md">
        <Box c="gray.4">
          <svg viewBox="0 0 24 24" fill="none" width="48" height="48">
            <path
              d="M22 12H18L15 21L9 3L6 12H2"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
          </svg>
        </Box>
        <Title order={3}>No stages yet</Title>
        <Text size="sm" c="dimmed" ta="center" className={classes.emptyHint}>
          Add stages to model the flow of this value stream.
        </Text>
        {canWrite && <AddStageButton onClick={() => onAddStage()} />}
      </Stack>
    </Paper>
  );
}

export function StageFlowDiagram({
  valueStreamId,
  stages,
  stageCapabilities,
  canWrite,
  onAddStage,
  onEditStage,
  onDeleteStage,
  onReorder,
  onAddCapability,
}: StageFlowDiagramProps) {
  const removeCapMutation = useRemoveStageCapability();
  const [draggedStageId, setDraggedStageId] = useState<string | null>(null);

  const sortedStages = [...stages].sort((a, b) => a.position - b.position);
  const capsByStage = groupCapabilitiesByStage(stageCapabilities);

  const handleDragStart = useCallback((e: React.DragEvent, stageId: string) => {
    setDraggedStageId(stageId);
    e.dataTransfer.effectAllowed = 'move';
  }, []);

  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.dataTransfer.dropEffect = e.dataTransfer.effectAllowed === 'copy' ? 'copy' : 'move';
  }, []);

  const handleDrop = useCallback(
    (e: React.DragEvent, targetStageId: string) => {
      e.preventDefault();
      if (!tryHandleCapabilityDrop(e, targetStageId, onAddCapability)) {
        applyStageReorder(draggedStageId, targetStageId, sortedStages, onReorder);
      }
      setDraggedStageId(null);
    },
    [draggedStageId, sortedStages, onReorder, onAddCapability],
  );

  const handleRemoveCapability = useCallback(
    async (mapping: StageCapabilityMapping) => {
      await removeCapMutation.mutateAsync({ mapping, valueStreamId });
    },
    [removeCapMutation, valueStreamId],
  );

  if (stages.length === 0) {
    return <EmptyStages canWrite={canWrite} onAddStage={onAddStage} />;
  }

  return (
    <Group gap={0} align="flex-start" wrap="nowrap" py="md" className={classes.scroll} data-testid="stage-flow-diagram">
      {sortedStages.map((stage, i) => (
        <Group key={stage.id} gap={0} align="flex-start" wrap="nowrap" className={classes.item}>
          {i > 0 && <StageConnector canWrite={canWrite} position={stage.position} onInsert={onAddStage} index={i} />}
          <StageColumn
            stage={stage}
            capabilities={capsByStage.get(stage.id) || []}
            canWrite={canWrite}
            onEdit={onEditStage}
            onDelete={onDeleteStage}
            onRemoveCapability={handleRemoveCapability}
            onDragStart={handleDragStart}
            onDragOver={handleDragOver}
            onDrop={handleDrop}
          />
        </Group>
      ))}
      {canWrite && (
        <Group gap={0} align="flex-start" wrap="nowrap" className={classes.item}>
          <Box className={classes.connector} />
          <AddStageButton onClick={() => onAddStage()} />
        </Group>
      )}
    </Group>
  );
}
