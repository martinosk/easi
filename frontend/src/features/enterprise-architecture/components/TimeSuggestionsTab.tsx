import { Badge, Box, Center, Group, Loader, Paper, Select, SimpleGrid, Stack, Table, Text, Title } from '@mantine/core';
import { useCallback, useMemo, useState } from 'react';
import { HelpTooltip } from '../../../components/shared/HelpTooltip';
import { useTimeSuggestions } from '../hooks/useTimeSuggestions';
import type { TimeClassification, TimeSuggestion } from '../types';

const TIME_CLASSIFICATIONS: { value: TimeClassification; color: string; description: string }[] = [
  {
    value: 'Tolerate',
    color: 'green',
    description: 'Low Value, High Technical Quality - Keep but limit investment',
  },
  {
    value: 'Invest',
    color: 'blue',
    description: 'High Value, High Technical Quality - Prioritize for funding',
  },
  {
    value: 'Migrate',
    color: 'yellow',
    description: 'High Value, Low Technical Quality - Upgrade or move to new platform',
  },
  {
    value: 'Eliminate',
    color: 'red',
    description: 'Low Value, Low Technical Quality - Retire or remove',
  },
];

function colorForTime(time: TimeClassification): string {
  const entry = TIME_CLASSIFICATIONS.find((item) => item.value === time);
  return entry?.color ?? 'gray';
}

function TimeLegend() {
  return (
    <Paper withBorder radius="md" p="sm">
      <Group gap="lg" wrap="wrap">
        <Text size="sm" fw={600}>
          TIME Classifications
        </Text>
        {TIME_CLASSIFICATIONS.map((item) => (
          <Group key={item.value} gap={6}>
            <Badge color={item.color} variant="light" radius="sm">
              {item.value}
            </Badge>
            <HelpTooltip content={item.description} iconOnly />
          </Group>
        ))}
      </Group>
    </Paper>
  );
}

function TimeBadge({ time }: { time: TimeClassification | null }) {
  if (!time) {
    return (
      <Badge color="gray" variant="light" radius="sm">
        N/A
      </Badge>
    );
  }
  return (
    <Badge color={colorForTime(time)} variant="light" radius="sm">
      {time}
    </Badge>
  );
}

function GapCell({ gap, label }: { gap: number | null; label: string }) {
  if (gap === null) {
    return (
      <Table.Td ta="center" title={label}>
        <Text size="sm" c="dimmed">
          -
        </Text>
      </Table.Td>
    );
  }
  const sign = gap > 0 ? '+' : '';
  const color = gap > 0 ? 'red.6' : gap < 0 ? 'green.6' : 'gray.7';
  return (
    <Table.Td ta="center" title={`${label}: ${sign}${gap.toFixed(1)}`}>
      <Text size="sm" fw={600} c={color}>
        {sign}
        {gap.toFixed(1)}
      </Text>
    </Table.Td>
  );
}

function SuggestionRow({ suggestion }: { suggestion: TimeSuggestion }) {
  return (
    <Table.Tr>
      <Table.Td>
        <Text size="sm">{suggestion.capabilityName}</Text>
      </Table.Td>
      <Table.Td>
        <Text size="sm">{suggestion.componentName}</Text>
      </Table.Td>
      <GapCell gap={suggestion.technicalGap} label="Technical Gap" />
      <GapCell gap={suggestion.functionalGap} label="Functional Gap" />
      <Table.Td ta="center">
        <TimeBadge time={suggestion.suggestedTime} />
      </Table.Td>
    </Table.Tr>
  );
}

interface SummaryStats {
  total: number;
  byClassification: Record<TimeClassification | 'Unknown', number>;
}

function calculateSummary(suggestions: TimeSuggestion[]): SummaryStats {
  const byClassification: Record<TimeClassification | 'Unknown', number> = {
    Tolerate: 0,
    Invest: 0,
    Migrate: 0,
    Eliminate: 0,
    Unknown: 0,
  };
  for (const s of suggestions) {
    if (s.suggestedTime) byClassification[s.suggestedTime]++;
    else byClassification.Unknown++;
  }
  return { total: suggestions.length, byClassification };
}

type GroupBy = 'none' | 'capability' | 'component';

function SummaryStat({ value, label, tooltip }: { value: number; label: string; tooltip: string }) {
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

interface HeaderProps {
  summary: SummaryStats;
  groupBy: GroupBy;
  onGroupByChange: (value: GroupBy) => void;
}

function TimeSuggestionsHeader({ summary, groupBy, onGroupByChange }: HeaderProps) {
  return (
    <Group justify="space-between" align="flex-end" wrap="wrap">
      <SimpleGrid cols={2} spacing="xl">
        <SummaryStat
          value={summary.total}
          label="Total Realizations"
          tooltip="Component-capability combinations with both strategic importance and fit scores"
        />
        <SummaryStat
          value={summary.byClassification.Eliminate}
          label="Eliminate"
          tooltip="Components suggested for phase-out due to both technical and functional gaps"
        />
      </SimpleGrid>
      <Select
        label="Group by"
        value={groupBy}
        onChange={(value) => value && onGroupByChange(value as GroupBy)}
        data={[
          { value: 'none', label: 'No grouping' },
          { value: 'capability', label: 'Enterprise Capability' },
          { value: 'component', label: 'Component' },
        ]}
        allowDeselect={false}
        w={220}
      />
    </Group>
  );
}

function TimeSuggestionsEmptyState() {
  return (
    <Paper withBorder radius="lg" p="xl">
      <Center>
        <Stack gap="sm" align="center" maw={420}>
          <Title order={4}>No TIME Suggestions Available</Title>
          <Box>
            <Text size="sm" c="dimmed">
              TIME suggestions require:
            </Text>
            <Text size="sm" c="dimmed">
              • Enterprise capabilities with strategic importance configured
            </Text>
            <Text size="sm" c="dimmed">
              • Components with fit scores
            </Text>
            <Text size="sm" c="dimmed">
              • Strategy pillars with fit types (Technical/Functional) enabled
            </Text>
          </Box>
        </Stack>
      </Center>
    </Paper>
  );
}

function SuggestionsTable({ suggestions }: { suggestions: TimeSuggestion[] }) {
  return (
    <Paper withBorder radius="md" p={0}>
      <Table striped highlightOnHover>
        <Table.Thead>
          <Table.Tr>
            <Table.Th>Capability</Table.Th>
            <Table.Th>Component</Table.Th>
            <Table.Th ta="center">
              <Group gap={4} justify="center">
                <Text size="sm" fw={600}>
                  Technical Gap
                </Text>
                <HelpTooltip
                  content="Difference between strategic importance and fit score for technical pillars. Positive = underperforming"
                  iconOnly
                />
              </Group>
            </Table.Th>
            <Table.Th ta="center">
              <Group gap={4} justify="center">
                <Text size="sm" fw={600}>
                  Functional Gap
                </Text>
                <HelpTooltip
                  content="Difference between strategic importance and fit score for functional pillars. Positive = underperforming"
                  iconOnly
                />
              </Group>
            </Table.Th>
            <Table.Th ta="center">
              <Group gap={4} justify="center">
                <Text size="sm" fw={600}>
                  Suggested TIME
                </Text>
                <HelpTooltip content="Recommended action based on technical and functional gap analysis" iconOnly />
              </Group>
            </Table.Th>
          </Table.Tr>
        </Table.Thead>
        <Table.Tbody>
          {suggestions.map((s) => (
            <SuggestionRow key={`${s.capabilityId}-${s.componentId}`} suggestion={s} />
          ))}
        </Table.Tbody>
      </Table>
    </Paper>
  );
}

export function TimeSuggestionsTab() {
  const [groupBy, setGroupBy] = useState<GroupBy>('none');
  const { suggestions, isLoading, error } = useTimeSuggestions();

  const summary = useMemo(() => calculateSummary(suggestions), [suggestions]);

  const sortedSuggestions = useMemo(() => {
    const sorted = [...suggestions];
    if (groupBy === 'capability') {
      sorted.sort((a, b) => a.capabilityName.localeCompare(b.capabilityName));
    } else if (groupBy === 'component') {
      sorted.sort((a, b) => a.componentName.localeCompare(b.componentName));
    }
    return sorted;
  }, [suggestions, groupBy]);

  const handleGroupByChange = useCallback((value: GroupBy) => {
    setGroupBy(value);
  }, []);

  if (isLoading) {
    return (
      <Center py="xl">
        <Group gap="sm">
          <Loader size="sm" />
          <Text c="dimmed">Loading TIME suggestions...</Text>
        </Group>
      </Center>
    );
  }

  if (error) {
    return (
      <Text c="red" size="sm">
        Failed to load TIME suggestions: {error.message}
      </Text>
    );
  }

  return (
    <Stack gap="md">
      <TimeSuggestionsHeader summary={summary} groupBy={groupBy} onGroupByChange={handleGroupByChange} />
      <TimeLegend />
      {suggestions.length === 0 ? <TimeSuggestionsEmptyState /> : <SuggestionsTable suggestions={sortedSuggestions} />}
    </Stack>
  );
}
