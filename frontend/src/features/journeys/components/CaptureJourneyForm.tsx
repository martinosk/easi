import { zodResolver } from '@hookform/resolvers/zod';
import {
  Alert,
  Button,
  Group,
  MultiSelect,
  NumberInput,
  SegmentedControl,
  Select,
  Stack,
  Text,
  Textarea,
  TextInput,
} from '@mantine/core';
import { useEffect, useMemo } from 'react';
import { Controller, type FieldErrors, useForm } from 'react-hook-form';
import type { BusinessDomainId, Capability, CapabilityRealization } from '../../../api/types';
import {
  carriesApplications,
  type CaptureJourneyFormData,
  captureJourneySchema,
  MAX_TARGET_MATURITY,
  MIN_TARGET_MATURITY,
} from '../../../lib/schemas/journey';
import { useBusinessDomainsQuery } from '../../business-domains/hooks/useBusinessDomains';
import { useCapabilitiesInDomainQuery } from '../../business-domains/hooks/useDomainCapabilities';
import { useCapabilities } from '../../capabilities/hooks/useCapabilities';
import { useComponents } from '../../components/hooks/useComponents';
import { getDescendantCapabilityIds } from '../../navigation/utils/filterByDomain';
import { useCaptureJourney } from '../hooks/useJourneys';
import type { CaptureJourneyRequest, JourneyKind } from '../types';
import { QUARTER_OPTIONS, yearOptions } from './periodFields';

interface CaptureJourneyFormProps {
  capability: Capability;
  realizations: CapabilityRealization[];
  availableKinds: JourneyKind[];
  onCaptured: () => void;
  onCancel: () => void;
}

const KIND_OPTIONS = [
  { value: 'migration', label: 'Migration' },
  { value: 'consolidation', label: 'Consolidation' },
  { value: 'carve-out', label: 'Carve-out' },
  { value: 'move', label: 'Move' },
  { value: 'maturity', label: 'Maturity' },
] as const satisfies ReadonlyArray<{ value: JourneyKind; label: string }>;

const KIND_DESCRIPTIONS = {
  migration: 'The realisation moves from the current application(s) to another. At least one source application.',
  consolidation:
    'Several applications merge onto one. Its current realisations, apart from the target, are the implicit sources.',
  'carve-out': 'Functionality is extracted from one application into another. Exactly one source application.',
  move: 'The capability relocates to another business domain or parent, under a new name. Its current realisations, apart from the target, are the implicit sources.',
  maturity:
    'The capability is deliberately matured to a higher level. No applications — the milestones carry the work, which need not be technical.',
} as const satisfies Record<JourneyKind, string>;

function defaultValues(
  capability: Capability,
  realizations: CapabilityRealization[],
  kind: JourneyKind,
): CaptureJourneyFormData {
  return {
    kind,
    fromComponentIds: carriesApplications(kind) ? realizations.map((r) => String(r.componentId)) : [],
    toComponentId: '',
    note: '',
    targetYear: undefined,
    targetQuarter: undefined,
    targetDomainId: '',
    targetParentId: '',
    resultingName: capability.name,
    targetMaturity: undefined,
  };
}

function hasImplicitSources(kind: JourneyKind): boolean {
  return kind === 'move' || kind === 'consolidation';
}

function implicitSources(realizations: CapabilityRealization[], toComponentId: string): string[] {
  return realizations.map((r) => String(r.componentId)).filter((id) => id !== toComponentId);
}

function toCaptureRequest(data: CaptureJourneyFormData): CaptureJourneyRequest {
  const hasPeriod = data.targetYear !== undefined && data.targetQuarter !== undefined;
  return {
    kind: data.kind,
    fromComponentIds: data.fromComponentIds,
    toComponentId: data.toComponentId,
    note: data.note.trim() || undefined,
    targetPeriod: hasPeriod ? { year: data.targetYear!, quarter: data.targetQuarter! } : null,
    ...(data.kind === 'move'
      ? {
          targetDomainId: data.targetDomainId,
          targetParentId: data.targetParentId || null,
          resultingName: data.resultingName.trim(),
        }
      : {}),
    ...(data.kind === 'maturity' ? { targetMaturity: data.targetMaturity } : {}),
  };
}

function useOptions(
  realizations: CapabilityRealization[],
  fromComponentIds: string[],
  kind: JourneyKind,
) {
  const componentsQuery = useComponents();
  const domainsQuery = useBusinessDomainsQuery();

  const fromAppOptions = useMemo(
    () => realizations.map((r) => ({ value: String(r.componentId), label: r.componentName ?? String(r.componentId) })),
    [realizations],
  );

  const toAppOptions = useMemo(
    () =>
      (componentsQuery.data ?? [])
        .filter((c) => hasImplicitSources(kind) || !fromComponentIds.includes(String(c.id)))
        .map((c) => ({ value: String(c.id), label: c.name })),
    [componentsQuery.data, fromComponentIds, kind],
  );

  const domainOptions = useMemo(
    () => (domainsQuery.data?.data ?? []).map((d) => ({ value: String(d.id), label: d.name })),
    [domainsQuery.data],
  );

  return { fromAppOptions, toAppOptions, domainOptions };
}

function useTargetParentOptions(capabilityId: string, targetDomainId: string) {
  const capabilitiesQuery = useCapabilities();
  const domainCapabilitiesQuery = useCapabilitiesInDomainQuery(
    targetDomainId ? (targetDomainId as BusinessDomainId) : undefined,
  );

  return useMemo(() => {
    if (!targetDomainId) return [];
    const allCapabilities = capabilitiesQuery.data ?? [];
    const directIds = new Set((domainCapabilitiesQuery.data ?? []).map((c) => String(c.id)));
    const effectiveIds = getDescendantCapabilityIds(directIds, allCapabilities);
    return allCapabilities
      .filter((c) => effectiveIds.has(String(c.id)) && String(c.id) !== capabilityId)
      .map((c) => ({ value: String(c.id), label: c.name }));
  }, [capabilitiesQuery.data, domainCapabilitiesQuery.data, capabilityId, targetDomainId]);
}

function useCaptureJourneyController(
  capability: Capability,
  realizations: CapabilityRealization[],
  availableKinds: JourneyKind[],
  onCaptured: () => void,
) {
  const captureMutation = useCaptureJourney();
  const form = useForm<CaptureJourneyFormData>({
    resolver: zodResolver(captureJourneySchema),
    defaultValues: defaultValues(capability, realizations, availableKinds[0]),
    mode: 'onChange',
  });
  const { watch, setValue, trigger } = form;
  const kind = watch('kind');
  const fromComponentIds = watch('fromComponentIds');
  const toComponentId = watch('toComponentId');
  const targetDomainId = watch('targetDomainId');

  useEffect(() => {
    if (!carriesApplications(kind)) {
      if (fromComponentIds.length > 0) setValue('fromComponentIds', [], { shouldValidate: true });
      if (toComponentId !== '') setValue('toComponentId', '', { shouldValidate: true });
      return;
    }
    if (hasImplicitSources(kind)) {
      setValue('fromComponentIds', implicitSources(realizations, toComponentId), { shouldValidate: true });
      void trigger('toComponentId');
    }
  }, [kind, realizations, fromComponentIds, toComponentId, setValue, trigger]);

  const options = {
    ...useOptions(realizations, fromComponentIds, kind),
    parentOptions: useTargetParentOptions(String(capability.id), targetDomainId),
  };
  const consolidationBlocked = kind === 'consolidation' && realizations.length < 2;

  const submit = form.handleSubmit(async (data) => {
    try {
      await captureMutation.mutateAsync({ capabilityId: String(capability.id), request: toCaptureRequest(data) });
      onCaptured();
    } catch {}
  });

  const errorMessage = captureMutation.error instanceof Error ? captureMutation.error.message : null;

  const clearTargetParent = () => setValue('targetParentId', '');

  return {
    form,
    kind,
    options,
    parentDisabled: !targetDomainId,
    clearTargetParent,
    submit,
    isPending: captureMutation.isPending,
    errorMessage,
    consolidationBlocked,
  };
}

type FormControl = ReturnType<typeof useForm<CaptureJourneyFormData>>['control'];

function KindField({ control, availableKinds }: { control: FormControl; availableKinds: JourneyKind[] }) {
  const options = KIND_OPTIONS.filter((option) => availableKinds.includes(option.value));
  return (
    <Controller
      name="kind"
      control={control}
      render={({ field }) => (
        <SegmentedControl
          data={options as unknown as { value: string; label: string }[]}
          value={field.value}
          onChange={field.onChange}
          fullWidth
          data-testid="journey-kind"
        />
      )}
    />
  );
}

function FromAppsField({
  control,
  options,
}: {
  control: FormControl;
  options: { value: string; label: string }[];
}) {
  return (
    <Controller
      name="fromComponentIds"
      control={control}
      render={({ field }) => (
        <MultiSelect
          label="From applications"
          description="The capability's current realising applications"
          data={options}
          value={field.value}
          onChange={field.onChange}
          searchable
          data-testid="journey-from-apps"
        />
      )}
    />
  );
}

function ToAppField({ control, options }: { control: FormControl; options: { value: string; label: string }[] }) {
  return (
    <Controller
      name="toComponentId"
      control={control}
      render={({ field }) => (
        <Select
          label="Target application"
          withAsterisk
          data={options}
          value={field.value || null}
          onChange={(value) => field.onChange(value ?? '')}
          searchable
          data-testid="journey-to-app"
        />
      )}
    />
  );
}

function TargetPeriodFields({ control }: { control: FormControl }) {
  const years = useMemo(yearOptions, []);
  return (
    <Group grow align="flex-start">
      <Controller
        name="targetYear"
        control={control}
        render={({ field }) => (
          <Select
            label="Target year"
            data={years}
            value={field.value !== undefined ? String(field.value) : ''}
            onChange={(value) => field.onChange(value ? Number(value) : undefined)}
            data-testid="journey-target-year"
          />
        )}
      />
      <Controller
        name="targetQuarter"
        control={control}
        render={({ field }) => (
          <Stack gap={4}>
            <Text size="sm" fw={500}>
              Target quarter
            </Text>
            <SegmentedControl
              data={QUARTER_OPTIONS}
              value={field.value !== undefined ? String(field.value) : ''}
              onChange={(value) => field.onChange(value ? Number(value) : undefined)}
              data-testid="journey-target-quarter"
            />
          </Stack>
        )}
      />
    </Group>
  );
}

function MoveFields({
  control,
  domainOptions,
  parentOptions,
  parentDisabled,
  onDomainChange,
}: {
  control: FormControl;
  domainOptions: { value: string; label: string }[];
  parentOptions: { value: string; label: string }[];
  parentDisabled: boolean;
  onDomainChange: () => void;
}) {
  return (
    <>
      <Controller
        name="targetDomainId"
        control={control}
        render={({ field }) => (
          <Select
            label="Target business domain"
            withAsterisk
            data={domainOptions}
            value={field.value || null}
            onChange={(value) => {
              field.onChange(value ?? '');
              onDomainChange();
            }}
            searchable
            data-testid="journey-target-domain"
          />
        )}
      />
      <Controller
        name="targetParentId"
        control={control}
        render={({ field }) => (
          <Select
            label="Target parent capability"
            description="Optional — a capability in the target business domain"
            placeholder={parentDisabled ? 'Select a target business domain first' : undefined}
            data={parentOptions}
            value={field.value || null}
            onChange={(value) => field.onChange(value ?? '')}
            disabled={parentDisabled}
            searchable
            clearable
            data-testid="journey-target-parent"
          />
        )}
      />
      <Controller
        name="resultingName"
        control={control}
        render={({ field }) => (
          <TextInput
            label="Resulting name"
            withAsterisk
            value={field.value}
            onChange={(e) => field.onChange(e.currentTarget.value)}
            data-testid="journey-resulting-name"
          />
        )}
      />
    </>
  );
}

function MaturityFields({ control }: { control: FormControl }) {
  return (
    <Controller
      name="targetMaturity"
      control={control}
      render={({ field, fieldState }) => (
        <NumberInput
          label="Target maturity"
          description={`The level this journey will reach, ${MIN_TARGET_MATURITY}–${MAX_TARGET_MATURITY}. It must be above the capability's current maturity.`}
          withAsterisk
          min={MIN_TARGET_MATURITY}
          max={MAX_TARGET_MATURITY}
          value={field.value ?? ''}
          onChange={(value) => field.onChange(value === '' ? undefined : Number(value))}
          error={fieldState.error?.message}
          data-testid="journey-target-maturity"
        />
      )}
    />
  );
}

function NoteField({ control }: { control: FormControl }) {
  return (
    <Controller
      name="note"
      control={control}
      render={({ field }) => (
        <Textarea
          label="Note"
          placeholder="Plan summary — what is moving, and why"
          maxLength={2000}
          autosize
          minRows={2}
          value={field.value}
          onChange={(e) => field.onChange(e.currentTarget.value)}
          data-testid="journey-note"
        />
      )}
    />
  );
}

function SubmitAlerts({
  errors,
  errorMessage,
}: {
  errors: FieldErrors<CaptureJourneyFormData>;
  errorMessage: string | null;
}) {
  const validationMessage = errors.fromComponentIds?.message ?? errors.toComponentId?.message;
  return (
    <>
      {validationMessage && (
        <Alert color="red" data-testid="capture-submit-error">
          {validationMessage}
        </Alert>
      )}
      {errorMessage && (
        <Alert color="red" data-testid="capture-journey-error">
          {errorMessage}
        </Alert>
      )}
    </>
  );
}

export function CaptureJourneyForm({
  capability,
  realizations,
  availableKinds,
  onCaptured,
  onCancel,
}: CaptureJourneyFormProps) {
  const { form, kind, options, parentDisabled, clearTargetParent, submit, isPending, errorMessage, consolidationBlocked } =
    useCaptureJourneyController(capability, realizations, availableKinds, onCaptured);
  const { control, formState } = form;

  return (
    <form onSubmit={submit} data-testid="capture-journey-form">
      <Stack gap="md">
        <Stack gap={4}>
          <KindField control={control} availableKinds={availableKinds} />
          <Text size="sm" c="dimmed" data-testid="journey-kind-description">
            {KIND_DESCRIPTIONS[kind]}
          </Text>
        </Stack>

        {consolidationBlocked && (
          <Alert color="yellow" data-testid="consolidation-gate">
            A consolidation needs at least two current realising applications. Use a migration to move a single
            realisation.
          </Alert>
        )}

        {carriesApplications(kind) && (
          <>
            {!hasImplicitSources(kind) && <FromAppsField control={control} options={options.fromAppOptions} />}
            <ToAppField control={control} options={options.toAppOptions} />
          </>
        )}

        {kind === 'maturity' && <MaturityFields control={control} />}

        <TargetPeriodFields control={control} />

        <NoteField control={control} />

        {kind === 'move' && (
          <MoveFields
            control={control}
            domainOptions={options.domainOptions}
            parentOptions={options.parentOptions}
            parentDisabled={parentDisabled}
            onDomainChange={clearTargetParent}
          />
        )}

        <SubmitAlerts errors={formState.errors} errorMessage={errorMessage} />

        <Group justify="flex-end" gap="sm">
          <Button variant="default" onClick={onCancel} disabled={isPending}>
            Cancel
          </Button>
          <Button
            type="submit"
            loading={isPending}
            disabled={!formState.isValid || isPending || consolidationBlocked}
            data-testid="capture-journey-submit"
          >
            Plan journey
          </Button>
        </Group>
      </Stack>
    </form>
  );
}
