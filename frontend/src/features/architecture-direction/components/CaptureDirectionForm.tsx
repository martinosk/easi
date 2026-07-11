import { zodResolver } from '@hookform/resolvers/zod';
import {
  Alert,
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
import { useEffect, useMemo, useState } from 'react';
import { Controller, useForm } from 'react-hook-form';
import type { EnterpriseCapabilityId } from '../../../api/types';
import { type CaptureDirectionFormData, captureDirectionSchema } from '../../../lib/schemas/direction';
import { useBusinessDomainsQuery } from '../../business-domains/hooks/useBusinessDomains';
import { useCaptureDirection, useCompositionPreview, useSourceCandidates } from '../hooks/useDirection';
import type { DirectionType, SourceCandidate } from '../types';
import { CompositionPreview, HORIZON_OPTIONS, toDomainOptions } from './sourcePickerPrimitives';

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

function buildCandidateFilter(candidateMap: Map<string, SourceCandidate>, domainId: string): OptionsFilter {
  return ({ options, search }) => {
    const term = search.toLowerCase();
    return (options as ComboboxItem[]).filter((o) => {
      const c = candidateMap.get(o.value);
      if (!c) return false;
      if (domainId && c.businessDomainId !== domainId) return false;
      return !term || c.name.toLowerCase().includes(term);
    });
  };
}

function useCaptureDirectionController(enterpriseCapabilityId: EnterpriseCapabilityId, onCaptured: () => void) {
  const captureMutation = useCaptureDirection();
  const form = useForm<CaptureDirectionFormData>({
    resolver: zodResolver(captureDirectionSchema),
    defaultValues: DEFAULT_VALUES,
    mode: 'onChange',
  });
  const { setValue } = form;

  const [selectedIds, setSelectedIds] = useState<string[]>([]);
  const [domainId, setDomainId] = useState('');

  useEffect(() => {
    setValue('sourceCapabilityIds', selectedIds, { shouldValidate: true });
  }, [selectedIds, setValue]);

  const candidatesQuery = useSourceCandidates(enterpriseCapabilityId);
  const previewQuery = useCompositionPreview(enterpriseCapabilityId, selectedIds);
  const { data: domainsResponse } = useBusinessDomainsQuery();
  const domainOptions = useMemo(() => toDomainOptions(domainsResponse?.data ?? []), [domainsResponse]);

  const candidateMap = useMemo(() => {
    const map = new Map<string, SourceCandidate>();
    for (const c of candidatesQuery.data?.data ?? []) {
      map.set(c.capabilityId, c);
    }
    return map;
  }, [candidatesQuery.data]);

  const allOptions = useMemo(
    () =>
      (candidatesQuery.data?.data ?? []).map((c) => ({
        value: c.capabilityId,
        label: c.name,
        disabled: !c.eligible,
      })),
    [candidatesQuery.data],
  );

  const filterOptions = useMemo(() => buildCandidateFilter(candidateMap, domainId), [candidateMap, domainId]);

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
    isPending: captureMutation.isPending,
    ineligibleSelected: previewQuery.data?.sourceEligibility.some((e) => !e.eligible) ?? false,
  };
}

export function CaptureDirectionForm({ enterpriseCapabilityId, onCaptured, onCancel }: CaptureDirectionFormProps) {
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
  } = useCaptureDirectionController(enterpriseCapabilityId, onCaptured);
  const { control, watch, formState } = form;
  const type = watch('type');

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
              {selectedIds.length}
            </Badge>
          </Group>
          <Text size="xs" c="dimmed">
            Search any domain capability by name (L1–L4, across all business domains). The chosen sources — and their
            subtrees — become this EC&apos;s composition.
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

          {selectedIds.length > 0 && <DraftCardinalityHint type={type} count={selectedIds.length} />}

          <CompositionPreview query={previewQuery} visible={selectedIds.length > 0} />
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

function DraftCardinalityHint({ type, count }: { type: DirectionType; count: number }) {
  const advanceRule =
    type === 'consolidate'
      ? 'Advancing to proposed requires at least 2 sources for a Consolidate direction.'
      : "Advancing to proposed enforces this type's source cardinality.";
  return (
    <Alert color="yellow" variant="light" data-testid="draft-cardinality-hint">
      {count} source{count === 1 ? '' : 's'} selected. A draft is accepted with any number of sources. {advanceRule}
    </Alert>
  );
}
