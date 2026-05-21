import { Badge, Box, Button, Center, Group, Loader, Paper, Select, SimpleGrid, Stack, Text, Title } from '@mantine/core';
import { useCallback, useState } from 'react';
import { HelpTooltip } from '../../../components/shared/HelpTooltip';
import { useMaturityColorScale } from '../../../hooks/useMaturityColorScale';
import { useMaturityScale } from '../../../hooks/useMaturityScale';
import { useMaturityAnalysis } from '../hooks/useMaturityAnalysis';
import type { EnterpriseCapabilityId, MaturityAnalysisCandidate } from '../types';
import classes from './MaturityAnalysisTab.module.css';

function MaturitySectionLegend() {
  const { data: maturityScale } = useMaturityScale();
  const { getBaseSectionColor } = useMaturityColorScale();

  if (!maturityScale?.sections) return null;

  const sortedSections = [...maturityScale.sections].sort((a, b) => a.order - b.order);

  return (
    <Paper withBorder radius="md" p="sm">
      <Group gap="md" wrap="wrap">
        <Group gap="xs">
          <Text size="sm" fw={600}>
            Maturity Sections
          </Text>
          <HelpTooltip
            content="Color coding for the maturity distribution bar. Sections are configured in Settings."
            iconOnly
          />
        </Group>
        {sortedSections.map((section) => (
          <Group key={section.order} gap={6} wrap="nowrap">
            <Box className={classes.dot} style={{ backgroundColor: getBaseSectionColor(section.order) }} />
            <Text size="xs">{section.name}</Text>
            <Text size="xs" c="dimmed">
              ({section.minValue}-{section.maxValue})
            </Text>
          </Group>
        ))}
      </Group>
    </Paper>
  );
}

interface MaturityDistributionBarProps {
  distribution: { genesis: number; customBuild: number; product: number; commodity: number };
}

function MaturityDistributionBar({ distribution }: MaturityDistributionBarProps) {
  const { data: maturityScale } = useMaturityScale();
  const { getBaseSectionColor } = useMaturityColorScale();

  const total = distribution.genesis + distribution.customBuild + distribution.product + distribution.commodity;
  if (total === 0) return null;

  const sortedSections = maturityScale?.sections ? [...maturityScale.sections].sort((a, b) => a.order - b.order) : [];
  const distributionByOrder = [
    distribution.genesis,
    distribution.customBuild,
    distribution.product,
    distribution.commodity,
  ];

  const segments = sortedSections
    .map((section, index) => ({
      name: section.name,
      count: distributionByOrder[index] || 0,
      color: getBaseSectionColor(section.order),
    }))
    .filter((s) => s.count > 0);

  return (
    <Stack gap={4}>
      <Box className={classes.distributionBar}>
        {segments.map((segment) => (
          <Box
            key={segment.name}
            className={classes.distributionSegment}
            style={{ width: `${(segment.count / total) * 100}%`, backgroundColor: segment.color }}
            title={`${segment.name}: ${segment.count}`}
          />
        ))}
      </Box>
      <Group gap="sm">
        {segments.map((segment) => (
          <Group key={segment.name} gap={4} wrap="nowrap">
            <Box className={classes.dot} style={{ backgroundColor: segment.color }} />
            <Text size="xs">{segment.count}</Text>
          </Group>
        ))}
      </Group>
    </Stack>
  );
}

function GapValue({ value }: { value: number }) {
  const c = value > 40 ? 'red.6' : value >= 15 ? 'yellow.7' : 'gray.7';
  return (
    <Text size="sm" fw={600} c={c}>
      {value}
    </Text>
  );
}

interface CandidateCardProps {
  candidate: MaturityAnalysisCandidate;
  onViewDetail: (id: EnterpriseCapabilityId) => void;
}

function CandidateCard({ candidate, onViewDetail }: CandidateCardProps) {
  const { getColorForValue, getSectionNameForValue } = useMaturityColorScale();

  const targetSection = candidate.targetMaturity !== null ? getSectionNameForValue(candidate.targetMaturity) : null;
  const targetColor = candidate.targetMaturity !== null ? getColorForValue(candidate.targetMaturity) : undefined;

  return (
    <Paper withBorder radius="md" p="md" shadow="xs">
      <Stack gap="sm">
        <Group justify="space-between" align="flex-start" wrap="nowrap">
          <Stack gap={4}>
            <Title order={5}>{candidate.enterpriseCapabilityName}</Title>
            {candidate.category && (
              <Badge variant="light" color="gray" radius="sm">
                {candidate.category}
              </Badge>
            )}
          </Stack>
          <Button
            size="compact-sm"
            variant="default"
            onClick={() => onViewDetail(candidate.enterpriseCapabilityId as EnterpriseCapabilityId)}
          >
            View Details
          </Button>
        </Group>

        <SimpleGrid cols={2} spacing="sm">
          <Stack gap={2}>
            <Group gap={4}>
              <Text size="xs" c="dimmed">
                Target Maturity
              </Text>
              <HelpTooltip content="Click View Details to set the target maturity level" iconOnly />
            </Group>
            {candidate.targetMaturity !== null && targetSection ? (
              <Group gap={4}>
                <Text size="sm" fw={600}>
                  {candidate.targetMaturity}
                </Text>
                <Text size="sm" fw={500} style={{ color: targetColor }}>
                  ({targetSection})
                </Text>
              </Group>
            ) : (
              <Text size="sm" c="dimmed" fs="italic">
                Not set
              </Text>
            )}
          </Stack>

          <Stack gap={2}>
            <Group gap={4}>
              <Text size="xs" c="dimmed">
                Implementations
              </Text>
              <HelpTooltip content="Number of domain capabilities linked to this enterprise capability" iconOnly />
            </Group>
            <Text size="sm" fw={600}>
              {candidate.implementationCount}
            </Text>
          </Stack>

          <Stack gap={2}>
            <Group gap={4}>
              <Text size="xs" c="dimmed">
                Domains
              </Text>
              <HelpTooltip content="Number of distinct business domains containing implementations" iconOnly />
            </Group>
            <Text size="sm" fw={600}>
              {candidate.domainCount}
            </Text>
          </Stack>

          <Stack gap={2}>
            <Group gap={4}>
              <Text size="xs" c="dimmed">
                Max Gap
              </Text>
              <HelpTooltip content="Largest maturity difference from target among all implementations" iconOnly />
            </Group>
            <GapValue value={candidate.maxGap} />
          </Stack>
        </SimpleGrid>

        <Group gap="xs">
          <Text size="xs" c="dimmed">
            Maturity Range:
          </Text>
          <Text size="xs">
            {candidate.minMaturity} - {candidate.maxMaturity}
          </Text>
          <Text size="xs" c="dimmed">
            (avg: {candidate.averageMaturity})
          </Text>
        </Group>

        <MaturityDistributionBar distribution={candidate.maturityDistribution} />
      </Stack>
    </Paper>
  );
}

interface SummaryStatProps {
  value: number | string;
  label: string;
  tooltip: string;
}

function SummaryStat({ value, label, tooltip }: SummaryStatProps) {
  return (
    <Stack gap={2}>
      <Text size="xl" fw={700}>
        {value}
      </Text>
      <Group gap={4}>
        <Text size="xs" c="dimmed">
          {label}
        </Text>
        <HelpTooltip content={tooltip} iconOnly />
      </Group>
    </Stack>
  );
}

interface MaturityAnalysisTabProps {
  onViewDetail: (id: EnterpriseCapabilityId) => void;
}

export function MaturityAnalysisTab({ onViewDetail }: MaturityAnalysisTabProps) {
  const [sortBy, setSortBy] = useState<string>('gap');
  const { candidates, summary, isLoading, error } = useMaturityAnalysis(sortBy);

  const handleSortChange = useCallback((value: string | null) => {
    if (value) setSortBy(value);
  }, []);

  if (isLoading) {
    return (
      <Center py="xl">
        <Group gap="sm">
          <Loader size="sm" />
          <Text c="dimmed">Loading maturity analysis...</Text>
        </Group>
      </Center>
    );
  }

  if (error) {
    return (
      <Text c="red" size="sm">
        Failed to load maturity analysis: {error.message}
      </Text>
    );
  }

  return (
    <Stack gap="md">
      <Group justify="space-between" align="flex-end" wrap="wrap">
        <Group gap="xl">
          {summary && (
            <>
              <SummaryStat
                value={summary.candidateCount}
                label="Capabilities"
                tooltip="Enterprise capabilities with linked domain capabilities that can be analyzed for maturity variance"
              />
              <SummaryStat
                value={summary.totalImplementations}
                label="Implementations"
                tooltip="Total domain capabilities linked to these enterprise capabilities"
              />
              <SummaryStat
                value={summary.averageGap}
                label="Avg Gap"
                tooltip="Average difference between implementation maturity and target (or highest implementation if no target set)"
              />
            </>
          )}
        </Group>
        <Select
          label="Sort by"
          value={sortBy}
          onChange={handleSortChange}
          data={[
            { value: 'gap', label: 'Max Gap' },
            { value: 'implementations', label: 'Implementations' },
          ]}
          allowDeselect={false}
          w={200}
        />
      </Group>

      <MaturitySectionLegend />

      {candidates.length === 0 ? (
        <EmptyMaturityState />
      ) : (
        <SimpleGrid cols={{ base: 1, md: 2 }} spacing="md">
          {candidates.map((candidate) => (
            <CandidateCard key={candidate.enterpriseCapabilityId} candidate={candidate} onViewDetail={onViewDetail} />
          ))}
        </SimpleGrid>
      )}
    </Stack>
  );
}

function EmptyMaturityState() {
  return (
    <Paper withBorder radius="lg" p="xl" shadow="xs">
      <Center>
        <Stack gap="sm" align="center" maw={400}>
          <Title order={4}>No Enterprise Capabilities</Title>
          <Text size="sm" c="dimmed" ta="center">
            Create enterprise capabilities to set target maturity and analyze gaps.
          </Text>
        </Stack>
      </Center>
    </Paper>
  );
}
