import { Alert, Button, Group, Loader, Paper, Stack, Text } from '@mantine/core';
import type { Horizon, SourceCandidate } from '../types';
import { useCompositionPreview, useSourceCandidates } from '../hooks/useDirection';

export const HORIZON_OPTIONS = [
  { value: 'now', label: 'Now' },
  { value: 'next', label: 'Next' },
  { value: 'later', label: 'Later' },
] as const satisfies ReadonlyArray<{ value: Horizon; label: string }>;

export function toDomainOptions(domains: { id: string; name: string }[]) {
  return [{ value: '', label: 'All domains' }, ...domains.map((d) => ({ value: d.id, label: d.name }))];
}

interface CandidateResultsProps {
  query: ReturnType<typeof useSourceCandidates>;
  selectedIds: Set<string>;
  onAdd: (candidate: SourceCandidate) => void;
  searchActive: boolean;
}

export function CandidateResults({ query, selectedIds, onAdd, searchActive }: CandidateResultsProps) {
  if (!searchActive) return null;
  if (query.isLoading) return <Loader size="sm" data-testid="candidates-loading" />;
  const candidates = query.data?.data ?? [];
  if (candidates.length === 0) {
    return (
      <Text size="xs" c="dimmed" data-testid="candidates-empty">
        No matching capabilities.
      </Text>
    );
  }
  return (
    <Paper withBorder radius="md" data-testid="candidate-results">
      <Stack gap={0}>
        {candidates.map((candidate) => (
          <CandidateRow
            key={candidate.capabilityId}
            candidate={candidate}
            selected={selectedIds.has(candidate.capabilityId)}
            onAdd={onAdd}
          />
        ))}
      </Stack>
    </Paper>
  );
}

function CandidateRow({
  candidate,
  selected,
  onAdd,
}: {
  candidate: SourceCandidate;
  selected: boolean;
  onAdd: (candidate: SourceCandidate) => void;
}) {
  return (
    <Group justify="space-between" wrap="nowrap" px="md" py="xs" data-testid={`candidate-${candidate.capabilityId}`}>
      <Stack gap={0}>
        <Text size="sm">{candidate.name}</Text>
        {candidate.eligible ? (
          <Text size="xs" c="dimmed">
            {candidate.businessDomainName ?? 'Unassigned'} · {candidate.level}
          </Text>
        ) : (
          <Text size="xs" c="red">
            ⛔ {candidate.ineligibilityReason}
          </Text>
        )}
      </Stack>
      <Button
        size="compact-xs"
        variant="filled"
        disabled={!candidate.eligible || selected}
        onClick={() => onAdd(candidate)}
        data-testid={`add-candidate-${candidate.capabilityId}`}
      >
        {selected ? 'Added' : '+ Add'}
      </Button>
    </Group>
  );
}

export function CompositionPreview({
  query,
  visible,
}: {
  query: ReturnType<typeof useCompositionPreview>;
  visible: boolean;
}) {
  if (!visible) return null;
  if (query.isLoading || !query.data) {
    return <Loader size="sm" data-testid="composition-preview-loading" />;
  }
  const included = query.data.includedCapabilities.filter((c) => c.role !== 'carved-out');
  const carved = query.data.includedCapabilities.filter((c) => c.role === 'carved-out');
  return (
    <Alert
      color="blue"
      variant="light"
      data-testid="composition-preview"
      title="This source implicitly includes its descendants"
    >
      <Stack gap={4}>
        <Text size="xs">Included here: {included.length > 0 ? included.map((c) => c.name).join(', ') : '—'}</Text>
        {carved.length > 0 && (
          <Text size="xs">
            Carved out:{' '}
            {carved.map((c) => `${c.name} (owned by ${c.carvedOutBy?.enterpriseCapabilityName})`).join(', ')}
          </Text>
        )}
      </Stack>
    </Alert>
  );
}
