import { zodResolver } from '@hookform/resolvers/zod';
import {
  Alert,
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
import { useEffect, useMemo, useState } from 'react';
import { Controller, useForm } from 'react-hook-form';
import type { EnterpriseCapabilityId } from '../../../api/types';
import { type CaptureDirectionFormData, captureDirectionSchema } from '../../../lib/schemas/direction';
import { useBusinessDomainsQuery } from '../../business-domains/hooks/useBusinessDomains';
import { useCaptureDirection, useCompositionPreview, useSourceCandidates } from '../hooks/useDirection';
import type { DirectionType, SourceCandidate } from '../types';
import { CandidateResults, CompositionPreview, HORIZON_OPTIONS, toDomainOptions } from './sourcePickerPrimitives';

interface CaptureDirectionFormProps {
  enterpriseCapabilityId: EnterpriseCapabilityId;
  onCaptured: () => void;
  onCancel: () => void;
}

const TYPE_OPTIONS = [
  { value: 'consolidate', label: 'Consolidate', help: 'Multiple capabilities merge into one.' },
  { value: 'decompose', label: 'Decompose', help: 'One capability splits into multiple.' },
  { value: 'stay', label: 'Stay', help: 'Explicitly confirmed: no change.' },
] as const satisfies ReadonlyArray<{ value: DirectionType; label: string; help: string }>;

const DEFAULT_VALUES: CaptureDirectionFormData = {
  type: 'consolidate',
  sourceCapabilityIds: [],
  horizon: 'next',
  narrative: '',
};

function useSourceSelection() {
  const [selected, setSelected] = useState<SourceCandidate[]>([]);
  const selectedIds = useMemo(() => selected.map((s) => s.capabilityId), [selected]);
  const selectedIdSet = useMemo(() => new Set(selectedIds), [selectedIds]);

  const add = (candidate: SourceCandidate) =>
    setSelected((prev) => (prev.some((s) => s.capabilityId === candidate.capabilityId) ? prev : [...prev, candidate]));
  const remove = (capabilityId: string) => setSelected((prev) => prev.filter((s) => s.capabilityId !== capabilityId));

  return { selected, selectedIds, selectedIdSet, add, remove };
}

function useCaptureDirectionController(enterpriseCapabilityId: EnterpriseCapabilityId, onCaptured: () => void) {
  const captureMutation = useCaptureDirection();
  const form = useForm<CaptureDirectionFormData>({
    resolver: zodResolver(captureDirectionSchema),
    defaultValues: DEFAULT_VALUES,
    mode: 'onChange',
  });
  const { setValue } = form;

  const [search, setSearch] = useState('');
  const [domainId, setDomainId] = useState('');
  const selection = useSourceSelection();

  useEffect(() => {
    setValue('sourceCapabilityIds', selection.selectedIds, { shouldValidate: true });
  }, [selection.selectedIds, setValue]);

  const candidatesQuery = useSourceCandidates(enterpriseCapabilityId, { q: search, domainId: domainId || undefined });
  const previewQuery = useCompositionPreview(enterpriseCapabilityId, selection.selectedIds);
  const { data: domainsResponse } = useBusinessDomainsQuery();
  const domainOptions = useMemo(() => toDomainOptions(domainsResponse?.data ?? []), [domainsResponse]);

  const submit = form.handleSubmit(async (data) => {
    try {
      await captureMutation.mutateAsync({
        enterpriseCapabilityId,
        request: {
          type: data.type,
          sourceCapabilityIds: data.sourceCapabilityIds,
          horizon: data.horizon,
          narrative: data.narrative.trim() || undefined,
        },
      });
      onCaptured();
    } catch {}
  });

  return {
    form,
    search,
    setSearch,
    domainId,
    setDomainId,
    domainOptions,
    selection,
    candidatesQuery,
    previewQuery,
    submit,
    isPending: captureMutation.isPending,
    ineligibleSelected: previewQuery.data?.sourceEligibility.some((e) => !e.eligible) ?? false,
  };
}

export function CaptureDirectionForm({ enterpriseCapabilityId, onCaptured, onCancel }: CaptureDirectionFormProps) {
  const {
    form,
    search,
    setSearch,
    domainId,
    setDomainId,
    domainOptions,
    selection,
    candidatesQuery,
    previewQuery,
    submit,
    isPending,
    ineligibleSelected,
  } = useCaptureDirectionController(enterpriseCapabilityId, onCaptured);
  const { control, watch, formState } = form;
  const type = watch('type');
  const selected = selection.selected;

  return (
    <form onSubmit={submit} data-testid="capture-direction-form">
      <Stack gap="md">
        <EnumField name="type" label="Direction type" control={control} options={TYPE_OPTIONS} />

        <Stack gap="xs" data-testid="source-picker">
          <Group justify="space-between" align="flex-end">
            <Text size="sm" fw={600}>
              Source capabilities
            </Text>
            <Badge variant="light" color="gray" data-testid="selected-count">
              {selected.length}
            </Badge>
          </Group>
          <Text size="xs" c="dimmed">
            Search any domain capability by name (L1–L4, across all business domains). The chosen sources — and their
            subtrees — become this EC&apos;s composition.
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
            selectedIds={selection.selectedIdSet}
            onAdd={selection.add}
            searchActive={search.trim().length > 0}
          />

          <SelectedSources selected={selected} onRemove={selection.remove} />

          {selected.length > 0 && <DraftCardinalityHint type={type} count={selected.length} />}

          <CompositionPreview query={previewQuery} visible={selected.length > 0} />
        </Stack>

        <EnumField name="horizon" label="Horizon" control={control} options={HORIZON_OPTIONS} />
        <NarrativeField control={control} />

        <Group justify="flex-end" gap="sm">
          <Button variant="default" onClick={onCancel} disabled={isPending}>
            Cancel
          </Button>
          <Button
            type="submit"
            loading={isPending}
            disabled={!formState.isValid || isPending || ineligibleSelected}
            data-testid="capture-submit"
          >
            Capture as draft
          </Button>
        </Group>
      </Stack>
    </form>
  );
}

type FormControl = ReturnType<typeof useForm<CaptureDirectionFormData>>['control'];

interface EnumFieldProps<TValue extends string> {
  name: 'type' | 'horizon';
  label: string;
  control: FormControl;
  options: ReadonlyArray<{ value: TValue; label: string; help?: string }>;
}

function EnumField<TValue extends string>({ name, label, control, options }: EnumFieldProps<TValue>) {
  return (
    <Controller
      name={name}
      control={control}
      render={({ field }) => (
        <Select
          label={label}
          data={options.map((o) => ({ value: o.value, label: o.label }))}
          value={field.value}
          onChange={(value) => value && field.onChange(value)}
          description={options.find((o) => o.value === field.value)?.help}
          allowDeselect={false}
          withAsterisk
        />
      )}
    />
  );
}

function NarrativeField({ control }: { control: FormControl }) {
  return (
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
  );
}

function SelectedSources({ selected, onRemove }: { selected: SourceCandidate[]; onRemove: (id: string) => void }) {
  if (selected.length === 0) return null;
  return (
    <Box>
      <Text size="xs" fw={600} c="dimmed" tt="uppercase">
        Selected sources
      </Text>
      <Group gap="xs" mt={6}>
        {selected.map((candidate) => (
          <Pill
            key={candidate.capabilityId}
            withRemoveButton
            onRemove={() => onRemove(candidate.capabilityId)}
            data-testid={`selected-chip-${candidate.capabilityId}`}
          >
            {candidate.level} · {candidate.name}
          </Pill>
        ))}
      </Group>
    </Box>
  );
}

function DraftCardinalityHint({ type, count }: { type: DirectionType; count: number }) {
  const advanceRule =
    type === 'consolidate'
      ? 'Advancing to proposed requires at least 2 sources for a Consolidate direction.'
      : 'Advancing to proposed enforces this type’s source cardinality.';
  return (
    <Alert color="yellow" variant="light" data-testid="draft-cardinality-hint">
      {count} source{count === 1 ? '' : 's'} selected. A draft is accepted with any number of sources.{' '}
      {advanceRule}
    </Alert>
  );
}
