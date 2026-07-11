import { zodResolver } from '@hookform/resolvers/zod';
import {
  Badge,
  Box,
  Button,
  type ComboboxItem,
  Group,
  MultiSelect,
  NativeSelect,
  type OptionsFilter,
  Select,
  Stack,
  Text,
  Textarea,
} from '@mantine/core';
import { IconBan } from '@tabler/icons-react';
import { useQueryClient } from '@tanstack/react-query';
import { useCallback, useMemo, useState } from 'react';
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
import { CompositionPreview, HORIZON_OPTIONS, toDomainOptions } from './sourcePickerPrimitives';

interface EditDraftDirectionFormProps {
  enterpriseCapabilityId: EnterpriseCapabilityId;
  direction: Direction;
  onSaved: () => void;
  onCancel: () => void;
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

  const initialSelectedIds = useMemo(
    () => direction.sourceCapabilities.filter((s) => !s.stale).map((s) => s.id as string),
    [direction.sourceCapabilities],
  );

  const [selectedIds, setSelectedIds] = useState<string[]>(initialSelectedIds);
  const [domainId, setDomainId] = useState('');

  const candidatesQuery = useSourceCandidates(enterpriseCapabilityId);
  const previewQuery = useCompositionPreview(enterpriseCapabilityId, selectedIds);
  const { data: domainsResponse } = useBusinessDomainsQuery();
  const domainOptions = useMemo(() => toDomainOptions(domainsResponse?.data ?? []), [domainsResponse]);

  const seedOptions = useMemo(
    () =>
      direction.sourceCapabilities
        .filter((s) => !s.stale)
        .map((s) => ({ value: s.id as string, label: s.name ?? (s.id as string), disabled: false })),
    [direction.sourceCapabilities],
  );

  const candidateOptions = useMemo(
    () =>
      (candidatesQuery.data?.data ?? []).map((c) => ({
        value: c.capabilityId,
        label: c.name,
        disabled: !c.eligible,
      })),
    [candidatesQuery.data],
  );

  const allOptions = useMemo(
    () => (candidateOptions.length > 0 ? candidateOptions : seedOptions),
    [candidateOptions, seedOptions],
  );

  const candidateMap = useMemo(() => {
    const map = new Map<string, SourceCandidate>();
    for (const c of candidatesQuery.data?.data ?? []) {
      map.set(c.capabilityId, c);
    }
    return map;
  }, [candidatesQuery.data]);

  const filterOptions: OptionsFilter = useCallback(
    ({ options, search }) => {
      const term = search.toLowerCase();
      return (options as ComboboxItem[]).filter((o) => {
        const c = candidateMap.get(o.value);
        if (!c) return !term && (!domainId || seedOptions.some((s) => s.value === o.value));
        if (domainId && c.businessDomainId !== domainId) return false;
        return !term || c.name.toLowerCase().includes(term);
      });
    },
    [candidateMap, domainId, seedOptions],
  );

  const ineligibleSelected = previewQuery.data?.sourceEligibility.some((e) => !e.eligible) ?? false;

  const submit = form.handleSubmit(async (data) => {
    setIsPending(true);
    try {
      const originalIds = new Set(direction.sourceCapabilities.map((s) => s.id as string));
      const currentIds = new Set(selectedIds);
      const toRemove = direction.sourceCapabilities.filter((s) => !currentIds.has(s.id as string));
      const toAdd = selectedIds.filter((id) => !originalIds.has(id));

      await Promise.all(toRemove.map((s) => directionApi.removeSource(enterpriseCapabilityId, s.id as string)));
      await Promise.all(toAdd.map((id) => directionApi.addSource(enterpriseCapabilityId, id)));

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
    selectedIds,
    onSelectionChange: setSelectedIds,
    allOptions,
    filterOptions,
    domainId,
    setDomainId,
    domainOptions,
    candidateMap,
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
    selectedIds,
    onSelectionChange,
    allOptions,
    filterOptions,
    domainId,
    setDomainId,
    domainOptions,
    candidateMap,
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
              {selectedIds.length}
            </Badge>
          </Group>
          <Text size="xs" c="dimmed">
            Add or remove domain capabilities. Changes take effect when you save.
          </Text>

          <Group gap="xs" align="flex-start" wrap="nowrap">
            <Box flex={1}>
              <MultiSelect
                placeholder={
                  candidatesQuery.isLoading ? 'Loading capabilities…' : 'Search or scroll to add capabilities…'
                }
                data={allOptions}
                value={selectedIds}
                onChange={onSelectionChange}
                searchable
                filter={filterOptions}
                renderOption={({ option }) => <CandidateOptionContent option={option} candidateMap={candidateMap} />}
                maxDropdownHeight={280}
                nothingFoundMessage="No matching capabilities"
                disabled={candidatesQuery.isLoading}
                data-testid="source-search-input"
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

          <CompositionPreview query={previewQuery} visible={selectedIds.length > 0} />
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

function CandidateOptionContent({
  option,
  candidateMap,
}: {
  option: ComboboxItem;
  candidateMap: Map<string, SourceCandidate>;
}) {
  const c = candidateMap.get(option.value);
  if (!c) return <Text size="sm">{option.label}</Text>;
  return (
    <Stack gap={0}>
      <Text size="sm">{c.name}</Text>
      {c.eligible ? (
        <Text size="xs" c="dimmed">
          {c.businessDomainName ?? 'Unassigned'} · {c.level}
        </Text>
      ) : (
        <Text size="xs" c="red">
          <IconBan size={16} stroke={1.75} /> {c.ineligibilityReason}
        </Text>
      )}
    </Stack>
  );
}
