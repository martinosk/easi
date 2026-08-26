import { Box, Button, Center, Group, Stack, Text, Title } from '@mantine/core';
import { useMemo } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import type { ValueStreamDetail } from '../../../api/types';
import { toValueStreamId } from '../../../api/types';
import { useUserStore } from '../../../store/userStore';
import { hasLink } from '../../../utils/hateoas';
import { CapabilitySidebar } from '../components/CapabilitySidebar';
import { StageFlowDiagram } from '../components/StageFlowDiagram';
import { StageFormOverlay } from '../components/StageFormOverlay';
import { SummaryBar } from '../components/SummaryBar';
import { useStageOperations } from '../hooks/useStageOperations';
import { useValueStreamDetail } from '../hooks/useValueStreamStages';
import classes from './ValueStreamDetailPage.module.css';

const BACK_ICON = (
  <svg viewBox="0 0 24 24" fill="none" width="14" height="14" aria-hidden="true">
    <path
      d="M19 12H5M12 19l-7-7 7-7"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
  </svg>
);

function LoadingState() {
  return (
    <Stack gap={0} className={classes.page}>
      <Center p="xxl">
        <Text c="dimmed">Loading value stream...</Text>
      </Center>
    </Stack>
  );
}

function ErrorState({ message }: { message?: string }) {
  return (
    <Stack gap={0} className={classes.page}>
      <Center p="xxl">
        <Text c="red">{message || 'Value stream not found'}</Text>
      </Center>
    </Stack>
  );
}

interface DetailHeaderProps {
  detail: ValueStreamDetail;
  uniqueCapCount: number;
}

function DetailHeader({ detail, uniqueCapCount }: DetailHeaderProps) {
  const navigate = useNavigate();
  return (
    <Box className={classes.header}>
      <Button
        variant="subtle"
        size="sm"
        mb="md"
        className={classes.backButton}
        onClick={() => navigate('/value-streams')}
        leftSection={BACK_ICON}
      >
        Back to Value Streams
      </Button>
      <Group justify="space-between" align="flex-start" mb="md">
        <Box>
          <Title order={1} mb="xs">
            {detail.name}
          </Title>
          {detail.description && (
            <Text size="sm" c="dimmed">
              {detail.description}
            </Text>
          )}
        </Box>
      </Group>
      <SummaryBar stageCount={detail.stages.length} capabilityCount={uniqueCapCount} />
    </Box>
  );
}

interface DetailContentProps {
  detail: ValueStreamDetail;
  canWrite: boolean;
}

function DetailContent({ detail, canWrite }: DetailContentProps) {
  const ops = useStageOperations(detail);

  const mappedCapabilityIds = useMemo(
    () => new Set((detail.stageCapabilities ?? []).map((c) => c.capabilityId as string)),
    [detail.stageCapabilities],
  );

  const uniqueCapCount = new Set(detail.stageCapabilities.map((c) => c.capabilityId)).size;
  const canAddStage = canWrite && hasLink(detail, 'x-add-stage');

  return (
    <Stack gap={0} className={classes.page} data-testid="value-stream-detail-page">
      <DetailHeader detail={detail} uniqueCapCount={uniqueCapCount} />

      {ops.isFormOpen && (
        <StageFormOverlay
          isEditing={ops.editingStage !== null}
          formData={ops.formData}
          onFormDataChange={ops.setFormData}
          onSubmit={ops.submitForm}
          onCancel={ops.closeForm}
        />
      )}

      <Group gap="xl" align="stretch" wrap="nowrap" px="xxl" py="xl" className={classes.content}>
        <Box flex={1} miw={0} className={classes.main}>
          <StageFlowDiagram
            valueStreamId={detail.id}
            stages={detail.stages}
            stageCapabilities={detail.stageCapabilities}
            canWrite={canAddStage}
            onAddStage={ops.openAddForm}
            onEditStage={ops.openEditForm}
            onDeleteStage={ops.deleteStage}
            onReorder={ops.reorderStages}
            onAddCapability={canWrite ? ops.addCapability : undefined}
          />
        </Box>
        {canWrite && <CapabilitySidebar mappedCapabilityIds={mappedCapabilityIds} />}
      </Group>
    </Stack>
  );
}

export function ValueStreamDetailPage() {
  const { valueStreamId } = useParams<{ valueStreamId: string }>();
  const id = valueStreamId ? toValueStreamId(valueStreamId) : undefined;
  const { data: detail, isLoading, error } = useValueStreamDetail(id);
  const hasPermission = useUserStore((state) => state.hasPermission);
  const canWrite = hasPermission('valuestreams:write');

  if (isLoading) return <LoadingState />;
  if (error) return <ErrorState message={`Failed to load: ${error.message}`} />;
  if (!detail) return <ErrorState />;

  return <DetailContent detail={detail} canWrite={canWrite} />;
}
