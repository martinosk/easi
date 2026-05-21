import { Alert, Badge, Box, Button, Center, Group, Loader, Modal, Paper, Stack, Text, Title } from '@mantine/core';
import { useCallback, useState } from 'react';
import { HelpTooltip } from '../../../components/shared/HelpTooltip';
import { useMaturityColorScale } from '../../../hooks/useMaturityColorScale';
import { useMaturityGapDetailHook, useSetTargetMaturity } from '../hooks/useMaturityAnalysis';
import type { EnterpriseCapabilityId, ImplementationDetail, MaturityGapDetail } from '../types';
import { SetTargetMaturityModal } from './SetTargetMaturityModal';
import classes from './MaturityGapDetailPanel.module.css';

type Priority = 'High' | 'Medium' | 'Low' | 'None';

function getPriorityColor(priority: Priority): string {
  switch (priority) {
    case 'High':
      return 'var(--color-error, #ef4444)';
    case 'Medium':
      return 'var(--color-warning, #f59e0b)';
    case 'Low':
      return 'var(--color-blue-500, #3b82f6)';
    default:
      return 'var(--color-green-500, #22c55e)';
  }
}

interface ImplementationBarProps {
  implementation: ImplementationDetail;
  targetMaturity: number;
  getColorForValue: (value: number) => string;
}

function ImplementationBar({ implementation, targetMaturity, getColorForValue }: ImplementationBarProps) {
  return (
    <Group gap="md" wrap="nowrap" align="center" py="xs">
      <Stack gap={0} miw={180} maw={180}>
        <Text size="sm" fw={500} truncate>
          {implementation.domainCapabilityName}
        </Text>
        {implementation.businessDomainName && (
          <Text size="xs" c="dimmed" truncate>
            {implementation.businessDomainName}
          </Text>
        )}
      </Stack>
      <Box flex={1}>
        <Box className={classes.bar}>
          <Box
            className={classes.barFill}
            style={{
              width: `${implementation.maturityValue}%`,
              backgroundColor: getColorForValue(implementation.maturityValue),
            }}
          />
          <Box className={classes.targetMarker} style={{ left: `${targetMaturity}%` }} title="Target maturity level" />
        </Box>
        <Group justify="space-between" gap="xs" mt={4}>
          <Text size="xs" c="dimmed">
            {implementation.maturityValue}
          </Text>
          <Text size="xs" fw={600} style={{ color: getPriorityColor(implementation.priority as Priority) }}>
            {implementation.gap > 0 ? `-${implementation.gap}` : 'On Target'}
          </Text>
        </Group>
      </Box>
    </Group>
  );
}

interface PrioritySectionProps {
  title: string;
  priority: Priority;
  implementations: ImplementationDetail[];
  targetMaturity: number;
  tooltip: string;
  getColorForValue: (value: number) => string;
}

function PrioritySection({
  title,
  priority,
  implementations,
  targetMaturity,
  tooltip,
  getColorForValue,
}: PrioritySectionProps) {
  if (implementations.length === 0) return null;

  return (
    <Paper withBorder radius="md" p="md" style={{ borderLeft: `4px solid ${getPriorityColor(priority)}` }}>
      <Group justify="space-between" mb="sm">
        <Group gap="xs">
          <Title order={5}>{title}</Title>
          <HelpTooltip content={tooltip} iconOnly />
        </Group>
        <Badge variant="light" color="gray">
          {implementations.length}
        </Badge>
      </Group>
      <Stack gap="xs">
        {implementations.map((impl) => (
          <ImplementationBar
            key={impl.domainCapabilityId}
            implementation={impl}
            targetMaturity={targetMaturity}
            getColorForValue={getColorForValue}
          />
        ))}
      </Stack>
    </Paper>
  );
}

interface TargetMaturityDisplayProps {
  detail: MaturityGapDetail;
  targetMaturity: number;
  getColorForValue: (value: number) => string;
  getSectionNameForValue: (value: number) => string;
  onOpenModal: () => void;
}

function TargetMaturityDisplay({
  detail,
  targetMaturity,
  getColorForValue,
  getSectionNameForValue,
  onOpenModal,
}: TargetMaturityDisplayProps) {
  const targetSection = detail.targetMaturity !== null ? getSectionNameForValue(detail.targetMaturity) : null;
  return (
    <Paper withBorder radius="md" p="md">
      <Group justify="space-between" wrap="nowrap">
        <Group gap="xs">
          <Text size="sm" c="dimmed">
            Target Maturity:
          </Text>
          {detail.targetMaturity !== null && targetSection ? (
            <Group gap="xs">
              <Text size="sm" fw={600}>
                {detail.targetMaturity}
              </Text>
              <Text size="sm" fw={600} style={{ color: getColorForValue(detail.targetMaturity) }}>
                ({targetSection})
              </Text>
            </Group>
          ) : (
            <Text size="sm" c="dimmed" fs="italic">
              Not set (using max: {targetMaturity})
            </Text>
          )}
        </Group>
        {detail._links?.['x-set-target-maturity'] && (
          <Button size="compact-sm" variant="default" onClick={onOpenModal}>
            {detail.targetMaturity !== null ? 'Edit Target' : 'Set Target'}
          </Button>
        )}
      </Group>
    </Paper>
  );
}

const PRIORITY_SECTIONS = [
  {
    title: 'High Priority (Gap > 40)',
    priority: 'High',
    key: 'high',
    tooltip: 'Implementations that need significant work to reach the target',
  },
  {
    title: 'Medium Priority (Gap 15-40)',
    priority: 'Medium',
    key: 'medium',
    tooltip: 'Implementations that need moderate work to reach the target',
  },
  {
    title: 'Low Priority (Gap 1-14)',
    priority: 'Low',
    key: 'low',
    tooltip: 'Implementations that need minor work to reach the target',
  },
  {
    title: 'On Target',
    priority: 'None',
    key: 'onTarget',
    tooltip: 'Implementations that meet or exceed the target maturity level',
  },
] as const satisfies ReadonlyArray<{
  title: string;
  priority: Priority;
  key: 'high' | 'medium' | 'low' | 'onTarget';
  tooltip: string;
}>;

interface MaturityGapDetailPanelProps {
  enterpriseCapabilityId: EnterpriseCapabilityId;
  onBack: () => void;
}

export function MaturityGapDetailPanel({ enterpriseCapabilityId, onBack }: MaturityGapDetailPanelProps) {
  const [isModalOpen, setIsModalOpen] = useState(false);
  const { detail, isLoading, error } = useMaturityGapDetailHook(enterpriseCapabilityId);
  const setTargetMaturityMutation = useSetTargetMaturity();
  const { getColorForValue, getSectionNameForValue, bounds } = useMaturityColorScale();

  const handleOpenModal = useCallback(() => setIsModalOpen(true), []);
  const handleCloseModal = useCallback(() => setIsModalOpen(false), []);

  const handleSaveTargetMaturity = useCallback(
    async (value: number) => {
      await setTargetMaturityMutation.mutateAsync({ enterpriseCapabilityId, targetMaturity: value });
      setIsModalOpen(false);
    },
    [enterpriseCapabilityId, setTargetMaturityMutation],
  );

  if (isLoading) {
    return (
      <Center py="xl">
        <Group gap="sm">
          <Loader size="sm" />
          <Text c="dimmed">Loading details...</Text>
        </Group>
      </Center>
    );
  }

  if (error || !detail) {
    return (
      <Stack gap="md">
        <Group>
          <Button variant="subtle" onClick={onBack}>
            ← Back to Analysis
          </Button>
        </Group>
        <Alert color="red">{error ? `Failed to load details: ${error.message}` : 'Capability not found'}</Alert>
      </Stack>
    );
  }

  const targetMaturity = detail.targetMaturity ?? Math.max(...detail.implementations.map((i) => i.maturityValue));

  return (
    <Stack gap="md">
      <Group>
        <Button variant="subtle" onClick={onBack}>
          ← Back to Analysis
        </Button>
      </Group>

      <Group justify="space-between" align="flex-start">
        <Stack gap="xs">
          <Title order={2}>{detail.enterpriseCapabilityName}</Title>
          {detail.category && (
            <Badge variant="light" color="gray" radius="sm">
              {detail.category}
            </Badge>
          )}
        </Stack>
      </Group>

      <TargetMaturityDisplay
        detail={detail}
        targetMaturity={targetMaturity}
        getColorForValue={getColorForValue}
        getSectionNameForValue={getSectionNameForValue}
        onOpenModal={handleOpenModal}
      />

      <Stack gap="sm">
        <Group gap="xs">
          <Title order={4}>Implementations ({detail.implementations.length})</Title>
          <HelpTooltip
            content="Each bar shows current maturity level. The vertical line marks the target. Gap is the difference between current and target maturity."
            iconOnly
          />
        </Group>
        {PRIORITY_SECTIONS.map(({ title, priority, key, tooltip }) => (
          <PrioritySection
            key={key}
            title={title}
            priority={priority}
            implementations={detail.investmentPriorities[key]}
            targetMaturity={targetMaturity}
            tooltip={tooltip}
            getColorForValue={getColorForValue}
          />
        ))}
      </Stack>

      <Modal opened={isModalOpen} onClose={handleCloseModal} title="Set Target Maturity" centered>
        <SetTargetMaturityModal
          currentValue={detail.targetMaturity}
          onClose={handleCloseModal}
          onSave={handleSaveTargetMaturity}
          isSaving={setTargetMaturityMutation.isPending}
          getColorForValue={getColorForValue}
          getSectionNameForValue={getSectionNameForValue}
          bounds={bounds}
        />
      </Modal>
    </Stack>
  );
}
