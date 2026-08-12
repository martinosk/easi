import { zodResolver } from '@hookform/resolvers/zod';
import {
  Alert,
  Button,
  Group,
  MultiSelect,
  SegmentedControl,
  Select,
  Stack,
  Text,
  Textarea,
  TextInput,
} from '@mantine/core';
import { useEffect, useMemo } from 'react';
import { Controller, type FieldErrors, useForm } from 'react-hook-form';
import type { Capability, CapabilityRealization } from '../../../api/types';
import { type CaptureJourneyFormData, captureJourneySchema } from '../../../lib/schemas/journey';
import { useBusinessDomainsQuery } from '../../business-domains/hooks/useBusinessDomains';
import { useCapabilities } from '../../capabilities/hooks/useCapabilities';
import { useComponents } from '../../components/hooks/useComponents';
import { useCaptureJourney } from '../hooks/useJourneys';
import type { CaptureJourneyRequest, JourneyKind } from '../types';
import { QUARTER_OPTIONS, yearOptions } from './periodFields';

interface CaptureJourneyFormProps {
  capability: Capability;
  realizations: CapabilityRealization[];
  onCaptured: () => void;
  onCancel: () => void;
}

const KIND_OPTIONS = [
  { value: 'migration', label: 'Migration' },
  { value: 'consolidation', label: 'Consolidation' },
  { value: 'carve-out', label: 'Carve-out' },
  { value: 'move', label: 'Move' },
] as const satisfies ReadonlyArray<{ value: JourneyKind; label: string }>;

const KIND_DESCRIPTIONS = {
  migration: 'The realisation moves from the current application(s) to another. At least one source application.',
  consolidation:
    'Several applications merge onto one. Its current realisations, apart from the target, are the implicit sources.',
  'carve-out': 'Functionality is extracted from one application into another. Exactly one source application.',
  move: 'The capability relocates to another business domain or parent, under a new name. Its current realisations, apart from the target, are the implicit sources.',
} as const satisfies Record<JourneyKind, string>;

function defaultValues(capability: Capability, realizations: CapabilityRealization[]): CaptureJourneyFormData {
  return {
    kind: 'migration',
    fromComponentIds: realizations.map((r) => String(r.componentId)),
    toComponentId: '',
    note: '',
    targetYear: undefined,
    targetQuarter: undefined,
    targetDomainId: '',
    targetParentId: '',
    resultingName: capability.name,
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
  };
}

function useOptions(
  realizations: CapabilityRealization[],
  capabilityId: string,
  fromComponentIds: string[],
  kind: JourneyKind,
) {
  const componentsQuery = useComponents();
  const domainsQuery = useBusinessDomainsQuery();
  const capabilitiesQuery = useCapabilities();

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

  const parentOptions = useMemo(
    () =>
      (capabilitiesQuery.data ?? [])
        .filter((c) => String(c.id) !== capabilityId)
        .map((c) => ({ value: String(c.id), label: c.name })),
    [capabilitiesQuery.data, capabilityId],
  );

  return { fromAppOptions, toAppOptions, domainOptions, parentOptions };
}

function useCaptureJourneyController(
  capability: Capability,
  realizations: CapabilityRealization[],
  onCaptured: () => void,
) {
  const captureMutation = useCaptureJourney();
  const form = useForm<CaptureJourneyFormData>({
    resolver: zodResolver(captureJourneySchema),
    defaultValues: defaultValues(capability, realizations),
    mode: 'onChange',
  });
  const { watch, setValue, trigger } = form;
  const kind = watch('kind');
  const fromComponentIds = watch('fromComponentIds');
  const toComponentId = watch('toComponentId');

  useEffect(() => {
    if (hasImplicitSources(kind)) {
      setValue('fromComponentIds', implicitSources(realizations, toComponentId), { shouldValidate: true });
      void trigger('toComponentId');
    }
  }, [kind, realizations, toComponentId, setValue, trigger]);

  const options = useOptions(realizations, String(capability.id), fromComponentIds, kind);
  const consolidationBlocked = kind === 'consolidation' && realizations.length < 2;

  const submit = form.handleSubmit(async (data) => {
    try {
      await captureMutation.mutateAsync({ capabilityId: String(capability.id), request: toCaptureRequest(data) });
      onCaptured();
    } catch {}
  });

  const errorMessage = captureMutation.error instanceof Error ? captureMutation.error.message : null;

  return { form, kind, options, submit, isPending: captureMutation.isPending, errorMessage, consolidationBlocked };
}

type FormControl = ReturnType<typeof useForm<CaptureJourneyFormData>>['control'];

function KindField({ control }: { control: FormControl }) {
  return (
    <Controller
      name="kind"
      control={control}
      render={({ field }) => (
        <SegmentedControl
          data={KIND_OPTIONS as unknown as { value: string; label: string }[]}
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
}: {
  control: FormControl;
  domainOptions: { value: string; label: string }[];
  parentOptions: { value: string; label: string }[];
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
            onChange={(value) => field.onChange(value ?? '')}
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
            description="Optional"
            data={parentOptions}
            value={field.value || null}
            onChange={(value) => field.onChange(value ?? '')}
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

export function CaptureJourneyForm({ capability, realizations, onCaptured, onCancel }: CaptureJourneyFormProps) {
  const { form, kind, options, submit, isPending, errorMessage, consolidationBlocked } = useCaptureJourneyController(
    capability,
    realizations,
    onCaptured,
  );
  const { control, formState } = form;

  return (
    <form onSubmit={submit} data-testid="capture-journey-form">
      <Stack gap="md">
        <Stack gap={4}>
          <KindField control={control} />
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

        {!hasImplicitSources(kind) && <FromAppsField control={control} options={options.fromAppOptions} />}

        <ToAppField control={control} options={options.toAppOptions} />

        <TargetPeriodFields control={control} />

        <NoteField control={control} />

        {kind === 'move' && (
          <MoveFields control={control} domainOptions={options.domainOptions} parentOptions={options.parentOptions} />
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
