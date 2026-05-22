import { Box, Divider, Group, Paper, Stack, Text } from '@mantine/core';
import type { BusinessDomainId, CapabilityId, ComponentId, FitCategory, FitComparison } from '../../../api/types';
import { useFitComparisons } from '../../components/hooks/useFitScores';

const CATEGORY_BG: Record<FitCategory, string> = {
  liability: 'red.0',
  concern: 'yellow.0',
  aligned: 'green.0',
};

const CATEGORY_COLOR: Record<FitCategory, string> = {
  liability: 'red',
  concern: 'yellow',
  aligned: 'green',
};

function getCategoryLabel(category: FitCategory, gap: number): string {
  switch (category) {
    case 'liability':
      return 'LIABILITY';
    case 'concern':
      return 'Minor';
    case 'aligned':
      return 'OK';
    default:
      return `Gap ${gap}`;
  }
}

interface FitComparisonDisplayProps {
  comparison: FitComparison;
}

function FitComparisonDisplay({ comparison }: FitComparisonDisplayProps) {
  return (
    <Paper px="xs" py="xs" radius="sm" bg={CATEGORY_BG[comparison.category]}>
      <Group gap="xs" wrap="wrap">
        <Text size="xs" fw={500} c="gray.7">
          {comparison.pillarName}:
        </Text>
        <Text size="xs" c="dimmed">
          Fit{' '}
          <Text component="span" fw={600}>
            {comparison.fitScore}
          </Text>{' '}
          vs Imp{' '}
          <Text component="span" fw={600}>
            {comparison.importance}
          </Text>
        </Text>
        <Text size="xs" c="dimmed">
          → Gap {comparison.gap}{' '}
          <Text component="span" fw={600} c={CATEGORY_COLOR[comparison.category]}>
            ({getCategoryLabel(comparison.category, comparison.gap)})
          </Text>
        </Text>
      </Group>
    </Paper>
  );
}

interface RealizationFitContextProps {
  componentId: ComponentId;
  capabilityId: CapabilityId;
  businessDomainId: BusinessDomainId;
}

export function RealizationFitContext({ componentId, capabilityId, businessDomainId }: RealizationFitContextProps) {
  const { data: comparisons = [] } = useFitComparisons(componentId, capabilityId, businessDomainId);

  const validComparisons = comparisons.filter((c) => c.importance > 0 && c.fitScore > 0);

  if (validComparisons.length === 0) return null;

  return (
    <Box mt="xs">
      <Divider variant="dashed" mb="xs" />
      <Text size="xs" c="dimmed" tt="uppercase" mb="xs">
        Fit vs Importance:
      </Text>
      <Stack gap="xs">
        {validComparisons.map((c) => (
          <FitComparisonDisplay key={c.pillarId} comparison={c} />
        ))}
      </Stack>
    </Box>
  );
}
