import { Accordion, Badge, Box, Center, Group, Loader, Paper, Select, SimpleGrid, Stack, Text, Title } from '@mantine/core';
import { useMemo, useState } from 'react';
import type { ApiError, RealizationFit, StrategicFitSummary } from '../../../api/types';
import { useStrategyPillarsConfig } from '../../../hooks/useStrategyPillarsSettings';
import { useStrategicFitAnalysis } from '../hooks/useStrategicFitAnalysis';

const SCORE_RANGE = [1, 2, 3, 4, 5] as const;

type FitCategory = 'liability' | 'concern' | 'aligned';

const CATEGORY_COLOR: Record<FitCategory, string> = {
  liability: 'red.6',
  concern: 'yellow.7',
  aligned: 'green.6',
};

function getAnalysisErrorMessage(error: unknown): string {
  if (error instanceof Error && 'statusCode' in error) {
    const apiError = error as ApiError;
    switch (apiError.statusCode) {
      case 400:
        return 'Fit scoring is not enabled for this pillar.';
      case 403:
        return 'You do not have permission to view strategic fit analysis.';
      case 404:
        return 'Strategy pillar not found.';
      default:
        return apiError.message || 'Failed to load analysis';
    }
  }
  return error instanceof Error ? error.message : 'Failed to load analysis';
}

interface SummaryCardProps {
  summary: StrategicFitSummary;
}

function SummaryCard({ summary }: SummaryCardProps) {
  return (
    <Paper withBorder radius="md" p="md">
      <Stack gap="sm">
        <SimpleGrid cols={3} spacing="md">
          <SummaryStat value={summary.liabilityCount} label="Liabilities" color="red.6" />
          <SummaryStat value={summary.concernCount} label="Concerns" color="yellow.7" />
          <SummaryStat value={summary.alignedCount} label="Aligned" color="green.6" />
        </SimpleGrid>
        <Group gap="md">
          <Text size="xs" c="dimmed">
            {summary.scoredRealizations} of {summary.totalRealizations} realizations scored
          </Text>
          {summary.averageGap > 0 && (
            <Text size="xs" c="dimmed">
              Average gap: {summary.averageGap.toFixed(1)}
            </Text>
          )}
        </Group>
      </Stack>
    </Paper>
  );
}

function SummaryStat({ value, label, color }: { value: number; label: string; color: string }) {
  return (
    <Stack gap={2} align="center">
      <Text size="xl" fw={700} c={color}>
        {value}
      </Text>
      <Text size="xs" c="dimmed">
        {label}
      </Text>
    </Stack>
  );
}

interface RealizationFitCardProps {
  realization: RealizationFit;
}

function RealizationFitCard({ realization }: RealizationFitCardProps) {
  const accent = CATEGORY_COLOR[realization.category as FitCategory];
  return (
    <Paper withBorder radius="md" p="md" style={{ borderLeft: `4px solid var(--mantine-color-${accent.replace('.', '-')})` }}>
      <Stack gap="sm">
        <Group justify="space-between" wrap="nowrap">
          <Group gap="xs" wrap="nowrap">
            <Text size="sm" fw={600}>
              {realization.componentName}
            </Text>
            <Text size="sm" c="dimmed">
              →
            </Text>
            <Text size="sm" fw={600}>
              {realization.capabilityName}
            </Text>
          </Group>
          {realization.businessDomainName && (
            <Badge variant="light" color="gray" radius="sm">
              {realization.businessDomainName}
            </Badge>
          )}
        </Group>

        <SimpleGrid cols={3} spacing="md">
          <Stack gap={2}>
            <Group gap={4}>
              <Text size="xs" c="dimmed">
                Importance
              </Text>
              {realization.isImportanceInherited && realization.importanceSourceCapabilityName && (
                <Text size="xs" c="dimmed" title={`Inherited from ${realization.importanceSourceCapabilityName}`}>
                  (from {realization.importanceSourceCapabilityName})
                </Text>
              )}
            </Group>
            <ScoreStars score={realization.importance} />
          </Stack>
          <Stack gap={2}>
            <Text size="xs" c="dimmed">
              Fit
            </Text>
            <ScoreDots score={realization.fitScore} />
          </Stack>
          <Stack gap={2}>
            <Text size="xs" c="dimmed">
              Gap
            </Text>
            <Text size="lg" fw={700} c={realization.gap >= 2 ? 'red.6' : realization.gap === 1 ? 'yellow.7' : 'gray.7'}>
              {realization.gap}
            </Text>
          </Stack>
        </SimpleGrid>

        {realization.importanceRationale && (
          <Text size="xs" c="dimmed" fs="italic">
            Strategic importance: &ldquo;{realization.importanceRationale}&rdquo;
          </Text>
        )}
        {realization.fitRationale && (
          <Text size="xs" c="dimmed" fs="italic">
            &ldquo;{realization.fitRationale}&rdquo;
          </Text>
        )}
      </Stack>
    </Paper>
  );
}

function ScoreStars({ score }: { score: number }) {
  return (
    <Group gap={2}>
      {SCORE_RANGE.map((i) => (
        <Text key={i} size="sm" c={i <= score ? 'yellow.6' : 'gray.4'}>
          ★
        </Text>
      ))}
      <Text size="xs" c="dimmed" ml={4}>
        ({score})
      </Text>
    </Group>
  );
}

function ScoreDots({ score }: { score: number }) {
  return (
    <Group gap={4} align="center">
      {SCORE_RANGE.map((i) => (
        <Box
          key={i}
          w={8}
          h={8}
          style={{
            borderRadius: '50%',
            backgroundColor: i <= score ? 'var(--mantine-color-blue-6)' : 'var(--mantine-color-gray-3)',
          }}
        />
      ))}
      <Text size="xs" c="dimmed" ml={4}>
        ({score})
      </Text>
    </Group>
  );
}

interface RealizationSectionProps {
  title: string;
  realizations: RealizationFit[];
  defaultExpanded?: boolean;
  category: FitCategory;
}

function RealizationSection({ title, realizations, defaultExpanded = false, category }: RealizationSectionProps) {
  if (realizations.length === 0) return null;
  return (
    <Accordion variant="separated" radius="md" defaultValue={defaultExpanded ? category : null}>
      <Accordion.Item value={category}>
        <Accordion.Control aria-label={`${title}, ${realizations.length} items`}>
          <Text fw={600} c={CATEGORY_COLOR[category]}>
            {title} ({realizations.length})
          </Text>
        </Accordion.Control>
        <Accordion.Panel>
          <Stack gap="sm">
            {realizations.map((r) => (
              <RealizationFitCard key={r.realizationId} realization={r} />
            ))}
          </Stack>
        </Accordion.Panel>
      </Accordion.Item>
    </Accordion>
  );
}

function LoadingState({ label }: { label: string }) {
  return (
    <Center py="xl">
      <Group gap="sm">
        <Loader size="sm" />
        <Text c="dimmed">{label}</Text>
      </Group>
    </Center>
  );
}

function NoPillarsState() {
  return (
    <Paper withBorder radius="lg" p="xl">
      <Center>
        <Stack gap="sm" align="center" maw={400}>
          <Title order={4}>No Pillars with Fit Scoring</Title>
          <Text size="sm" c="dimmed" ta="center">
            Enable fit scoring for strategy pillars in Settings to analyze strategic alignment.
          </Text>
        </Stack>
      </Center>
    </Paper>
  );
}

interface FitAnalysisResultsProps {
  selectedPillarId: string | null;
  analysisLoading: boolean;
  error: Error | null;
  analysis: ReturnType<typeof useStrategicFitAnalysis>['data'];
}

function FitAnalysisResults({ selectedPillarId, analysisLoading, error, analysis }: FitAnalysisResultsProps) {
  if (!selectedPillarId) {
    return <Text c="dimmed">Select a strategy pillar to view the fit analysis</Text>;
  }
  if (analysisLoading) return <LoadingState label="Loading analysis..." />;
  if (error) {
    return (
      <Text c="red" size="sm">
        {getAnalysisErrorMessage(error)}
      </Text>
    );
  }
  if (!analysis) return null;

  return (
    <Stack gap="md">
      <SummaryCard summary={analysis.summary} />
      <RealizationSection
        title="Strategic Liabilities"
        realizations={analysis.liabilities}
        defaultExpanded
        category="liability"
      />
      <RealizationSection
        title="Concerns"
        realizations={analysis.concerns}
        defaultExpanded={analysis.liabilities.length === 0}
        category="concern"
      />
      <RealizationSection title="Well Aligned" realizations={analysis.aligned} category="aligned" />
    </Stack>
  );
}

export function StrategicFitTab() {
  const { data: pillarsConfig, isLoading: pillarsLoading } = useStrategyPillarsConfig();
  const [selectedPillarId, setSelectedPillarId] = useState<string | null>(null);
  const { data: analysis, isLoading: analysisLoading, error } = useStrategicFitAnalysis(selectedPillarId);

  const enabledPillars = useMemo(() => {
    if (!pillarsConfig?.data) return [];
    return pillarsConfig.data.filter((p) => p.active && p.fitScoringEnabled);
  }, [pillarsConfig]);

  if (pillarsLoading) return <LoadingState label="Loading pillars..." />;
  if (enabledPillars.length === 0) return <NoPillarsState />;

  return (
    <Stack gap="md">
      <Group justify="space-between" align="flex-end" wrap="wrap">
        <Stack gap={4}>
          <Title order={3}>Strategic Fit Analysis</Title>
          <Text size="sm" c="dimmed">
            Identify realizations where application fit does not match strategic importance
          </Text>
        </Stack>
        <Select
          label="Filter by pillar"
          aria-label="Select strategy pillar for fit analysis"
          placeholder="Select a pillar"
          value={selectedPillarId}
          onChange={setSelectedPillarId}
          data={enabledPillars.map((pillar) => ({ value: pillar.id, label: pillar.name }))}
          w={260}
        />
      </Group>
      <FitAnalysisResults
        selectedPillarId={selectedPillarId}
        analysisLoading={analysisLoading}
        error={error}
        analysis={analysis}
      />
    </Stack>
  );
}
