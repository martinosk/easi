import { zodResolver } from '@hookform/resolvers/zod';
import {
  Badge,
  Box,
  Button,
  Group,
  NativeSelect,
  Pill,
  Select,
  Stack,
  Text,
  Textarea,
  TextInput,
} from '@mantine/core';
import { useQueryClient } from '@tanstack/react-query';
import { useMemo, useState } from 'react';
import { Controller, useForm } from 'react-hook-form';
import toast from 'react-hot-toast';
import type { EnterpriseCapabilityId } from '../../../api/types';
import { invalidateFor } from '../../../lib/invalidateFor';
import { type EditDirectionFormData, editDirectionSchema } from '../../../lib/schemas/direction';
import { useBusinessDomainsQuery } from '../../business-domains/hooks/useBusinessDomains';
import { directionApi } from '../api/directionApi';
import { useCompositionPreview, useSourceCandidates } from '../hooks/useDirection';
import { directionMutationEffects } from '../mutationEffects';
import type { Direction, SourceCandidate } from '../types';
import { CandidateResults, CompositionPreview, HORIZON_OPTIONS, toDomainOptions } from './sourcePickerPrimitives';

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
      const currentCapabilityIds = new Set(sources.map((s) => s.capabilityId));
      const toRemove = direction.sourceCapabilities.filter((s) => !currentCapabilityIds.has(s.id as string));
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
