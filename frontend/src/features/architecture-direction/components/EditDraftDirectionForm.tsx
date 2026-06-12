import { zodResolver } from '@hookform/resolvers/zod';
import {
  Alert,
  Badge,
  Box,
  Button,
  Group,
  Loader,
  NativeSelect,
  Paper,
  Pill,
  Select,
  Stack,
  Text,
  Textarea,
  TextInput,
} from '@mantine/core';
import { useMemo, useState } from 'react';
import { Controller, useForm } from 'react-hook-form';
import { useQueryClient } from '@tanstack/react-query';
import toast from 'react-hot-toast';
import type { EnterpriseCapabilityId } from '../../../api/types';
import { invalidateFor } from '../../../lib/invalidateFor';
import { type EditDirectionFormData, editDirectionSchema } from '../../../lib/schemas/direction';
import { useBusinessDomainsQuery } from '../../business-domains/hooks/useBusinessDomains';
import { directionApi } from '../api/directionApi';
import { useCompositionPreview, useSourceCandidates } from '../hooks/useDirection';
import { directionMutationEffects } from '../mutationEffects';
import type { Direction, Horizon, SourceCandidate } from '../types';

interface EditDraftDirectionFormProps {
  enterpriseCapabilityId: EnterpriseCapabilityId;
  direction: Direction;
  onSaved: () => void;
  onCancel: () => void;
}

interface EditableSource {
  capabilityId: string;
  name: string | null;
}

const HORIZON_OPTIONS = [
  { value: 'now', label: 'Now' },
  { value: 'next', label: 'Next' },
  { value: 'later', label: 'Later' },
] as const satisfies ReadonlyArray<{ value: Horizon; label: string }>;

function toDomainOptions(domains: { id: string; name: string }[]) {
  return [{ value: '', label: 'All domains' }, ...domains.map((d) => ({ value: d.id, label: d.name }))];
}

function useSourceSelection(initial: EditableSource[]) {
  const [sources, setSources] = useState(initial);
  const selectedIdSet = useMemo(() => new Set(sources.map((s) => s.capabilityId)), [sources]);

  const addSource = (candidate: SourceCandidate) =>
    setSources((prev) =>
      prev.some((s) => s.capabilityId === candidate.capabilityId)
        ? prev
        : [...prev, { capabilityId: candidate.capabilityId, name: candidate.name }],
    );

  const removeSource = (capabilityId: string) =>
    setSources((prev) => prev.filter((s) => s.capabilityId !== capabilityId));

  return { sources, selectedIdSet, addSource, removeSource };
}

function useEditDraftController(
  enterpriseCapabilityId: EnterpriseCapabilityId,
  direction: Direction,
  onSaved: () => void,
) {
  const queryClient = useQueryClient();
  const [isPending, setIsPending] = useState(false);

  const form = useForm<EditDirectionFormData>({
    resolver: zodResolver(editDirectionSchema),
    defaultValues: {
      horizon: direction.horizon,
      narrative: direction.narrative ?? '',
    },
  });

  const initialSources = direction.sourceCapabilities
    .filter((s) => !s.stale)
    .map((s): EditableSource => ({ capabilityId: s.id as string, name: s.name }));

  const { sources, selectedIdSet, addSource, removeSource } = useSourceSelection(initialSources);

  const [search, setSearch] = useState('');
  const [domainId, setDomainId] = useState('');

  const candidatesQuery = useSourceCandidates(enterpriseCapabilityId, { q: search, domainId: domainId || undefined });
  const previewQuery = useCompositionPreview(
    enterpriseCapabilityId,
    sources.map((s) => s.capabilityId),
  );

  const { data: domainsResponse } = useBusinessDomainsQuery();
  const domainOptions = useMemo(() => toDomainOptions(domainsResponse?.data ?? []), [domainsResponse]);

  const ineligibleSelected = previewQuery.data?.sourceEligibility.some((e) => !e.eligible) ?? false;

  const submit = form.handleSubmit(async (data) => {
    setIsPending(true);
    try {
      const originalIds = new Set(direction.sourceCapabilities.map((s) => s.id as string));
      const currentIds = new Set(sources.map((s) => s.capabilityId));
      const toRemove = direction.sourceCapabilities.filter((s) => !currentIds.has(s.id as string));
      const toAdd = sources.filter((s) => !originalIds.has(s.capabilityId));

      await Promise.all(toRemove.map((s) => directionApi.removeSource(enterpriseCapabilityId, s.id as string)));
      await Promise.all(toAdd.map((s) => directionApi.addSource(enterpriseCapabilityId, s.capabilityId)));

      await directionApi.update(enterpriseCapabilityId, {
        horizon: data.horizon,
        narrative: data.narrative.trim() || undefined,
      });

      invalidateFor(queryClient, directionMutationEffects.update(enterpriseCapabilityId));
      toast.success('Direction updated');
      onSaved();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to update direction');
    } finally {
      setIsPending(false);
    }
  });

  return {
    form,
    sources,
    selectedIdSet,
    addSource,
    removeSource,
    search,
    setSearch,
    domainId,
    setDomainId,
    domainOptions,
    candidatesQuery,
    previewQuery,
    submit,
    isPending,
    ineligibleSelected,
  };
}

export function EditDraftDirectionForm({
  enterpriseCapabilityId,
  direction,
  onSaved,
  onCancel,
}: EditDraftDirectionFormProps) {
  const {
    form,
    sources,
    selectedIdSet,
    addSource,
    removeSource,
    search,
    setSearch,
    domainId,
    setDomainId,
    domainOptions,
    candidatesQuery,
    previewQuery,
    submit,
    isPending,
    ineligibleSelected,
  } = useEditDraftController(enterpriseCapabilityId, direction, onSaved);
  const { control } = form;

  return (
    <form onSubmit={submit} data-testid="edit-draft-direction-form">
      <Stack gap="md">
        <Stack gap="xs" data-testid="source-picker">
          <Group justify="space-between" align="flex-end">
            <Text size="sm" fw={600}>
              Source capabilities
            </Text>
            <Badge variant="light" color="gray" data-testid="selected-count">
              {sources.length}
            </Badge>
          </Group>
          <Text size="xs" c="dimmed">
            Add or remove domain capabilities. Changes take effect when you save.
          </Text>

          <Group gap="xs" grow={false} wrap="nowrap">
            <Box flex={1}>
              <TextInput
                placeholder="Search capabilities…"
                value={search}
                onChange={(e) => setSearch(e.currentTarget.value)}
                data-testid="source-search-input"
                aria-label="Search capabilities"
              />
            </Box>
            <NativeSelect
              data={domainOptions}
              value={domainId}
              onChange={(e) => setDomainId(e.currentTarget.value)}
              aria-label="Filter by business domain"
              data-testid="domain-filter"
            />
          </Group>

          <CandidateResults
            query={candidatesQuery}
            selectedIds={selectedIdSet}
            onAdd={addSource}
            searchActive={search.trim().length > 0}
          />

          <EditableSourcePills sources={sources} onRemove={removeSource} />

          <CompositionPreview query={previewQuery} visible={sources.length > 0} />
        </Stack>

        <Controller
          name="horizon"
          control={control}
          render={({ field }) => (
            <Select
              label="Horizon"
              data={HORIZON_OPTIONS.map((o) => ({ value: o.value, label: o.label }))}
              value={field.value}
              onChange={(value) => value && field.onChange(value)}
              allowDeselect={false}
              withAsterisk
            />
          )}
        />

        <Controller
          name="narrative"
          control={control}
          render={({ field }) => (
            <Textarea
              label="Narrative"
              description="Required before advancing the direction to proposed."
              placeholder="One or two sentences naming what the group decided and why"
              autosize
              minRows={3}
              value={field.value}
              onChange={(e) => field.onChange(e.currentTarget.value)}
            />
          )}
        />

        <Group justify="flex-end" gap="sm">
          <Button variant="default" onClick={onCancel} disabled={isPending}>
            Cancel
          </Button>
          <Button
            type="submit"
            loading={isPending}
            disabled={isPending || ineligibleSelected}
            data-testid="edit-draft-submit"
          >
            Save changes
          </Button>
        </Group>
      </Stack>
    </form>
  );
}

interface CandidateResultsProps {
  query: ReturnType<typeof useSourceCandidates>;
  selectedIds: Set<string>;
  onAdd: (candidate: SourceCandidate) => void;
  searchActive: boolean;
}

function CandidateResults({ query, selectedIds, onAdd, searchActive }: CandidateResultsProps) {
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

function EditableSourcePills({
  sources,
  onRemove,
}: {
  sources: EditableSource[];
  onRemove: (capabilityId: string) => void;
}) {
  if (sources.length === 0) return null;
  return (
    <Box>
      <Text size="xs" fw={600} c="dimmed" tt="uppercase">
        Selected sources
      </Text>
      <Group gap="xs" mt={6}>
        {sources.map((source) => (
          <Pill
            key={source.capabilityId}
            withRemoveButton
            onRemove={() => onRemove(source.capabilityId)}
            data-testid={`selected-chip-${source.capabilityId}`}
          >
            {source.name ?? source.capabilityId}
          </Pill>
        ))}
      </Group>
    </Box>
  );
}

function CompositionPreview({ query, visible }: { query: ReturnType<typeof useCompositionPreview>; visible: boolean }) {
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
