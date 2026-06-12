import { Alert, Loader, Stack, Text } from '@mantine/core';
import type { Horizon } from '../types';
import { useCompositionPreview } from '../hooks/useDirection';

export const HORIZON_OPTIONS = [
  { value: 'now', label: 'Now' },
  { value: 'next', label: 'Next' },
  { value: 'later', label: 'Later' },
] as const satisfies ReadonlyArray<{ value: Horizon; label: string }>;

export function toDomainOptions(domains: { id: string; name: string }[]) {
  return [{ value: '', label: 'All domains' }, ...domains.map((d) => ({ value: d.id, label: d.name }))];
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
