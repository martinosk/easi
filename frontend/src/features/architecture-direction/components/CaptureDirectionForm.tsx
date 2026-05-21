import { zodResolver } from '@hookform/resolvers/zod';
import {
  ActionIcon,
  Alert,
  Box,
  Button,
  Checkbox,
  Group,
  Loader,
  Select,
  Stack,
  Text,
  Textarea,
  TextInput,
} from '@mantine/core';
import { useEffect, useMemo } from 'react';
import { Controller, useFieldArray, useForm } from 'react-hook-form';
import type { EnterpriseCapabilityId } from '../../../api/types';
import { type CaptureDirectionFormData, captureDirectionSchema } from '../../../lib/schemas/direction';
import { useBusinessDomainsQuery } from '../../business-domains/hooks/useBusinessDomains';
import {
  useEnterpriseCapability,
  useEnterpriseCapabilityLinks,
} from '../../enterprise-architecture/hooks/useEnterpriseCapabilities';
import type { EnterpriseCapabilityLink } from '../../enterprise-architecture/types';
import { useCaptureDirection } from '../hooks/useDirection';
import type { DirectionType, Horizon } from '../types';

interface CaptureDirectionFormProps {
  enterpriseCapabilityId: EnterpriseCapabilityId;
  onCaptured: () => void;
  onCancel: () => void;
}

const TYPE_OPTIONS = [
  {
    value: 'consolidate',
    label: 'Consolidate',
    help: 'Multiple physical capabilities merge into one. Pick 2 or more sources.',
  },
  {
    value: 'decompose',
    label: 'Decompose',
    help: 'One physical capability splits into multiple. Pick exactly 1 source.',
  },
  {
    value: 'stay',
    label: 'Stay',
    help: 'Explicitly confirmed no change. Pick exactly 1 source. No target placements.',
  },
] as const satisfies ReadonlyArray<{ value: DirectionType; label: string; help: string }>;

const HORIZON_OPTIONS = [
  { value: 'now', label: 'Now' },
  { value: 'next', label: 'Next' },
  { value: 'later', label: 'Later' },
] as const satisfies ReadonlyArray<{ value: Horizon; label: string }>;

const DEFAULT_VALUES: CaptureDirectionFormData = {
  type: 'consolidate',
  sourceCapabilityIds: [],
  placements: [],
  horizon: 'next',
  narrative: '',
};

function canAddPlacement(type: DirectionType, currentCount: number): boolean {
  if (type === 'consolidate') return currentCount === 0;
  if (type === 'decompose') return true;
  return false;
}

function useCaptureDirectionFormState(enterpriseCapabilityId: EnterpriseCapabilityId, onCaptured: () => void) {
  const captureMutation = useCaptureDirection();
  const form = useForm<CaptureDirectionFormData>({
    resolver: zodResolver(captureDirectionSchema),
    defaultValues: DEFAULT_VALUES,
    mode: 'onChange',
  });
  const placementsArray = useFieldArray({ control: form.control, name: 'placements' });
  const type = form.watch('type');
  const fieldCount = placementsArray.fields.length;

  useEffect(() => {
    if (type === 'stay' && fieldCount > 0) {
      form.setValue('placements', []);
    }
  }, [type, fieldCount, form]);

  const onSubmit = async (data: CaptureDirectionFormData) => {
    try {
      await captureMutation.mutateAsync({
        enterpriseCapabilityId,
        request: {
          type: data.type,
          sourceCapabilityIds: data.sourceCapabilityIds,
          placements: data.type === 'stay' ? [] : data.placements,
          horizon: data.horizon,
          narrative: data.narrative.trim() || undefined,
        },
      });
      onCaptured();
    } catch {
      // toast handled by hook
    }
  };

  return { form, placementsArray, onSubmit, isPending: captureMutation.isPending };
}

export function CaptureDirectionForm({ enterpriseCapabilityId, onCaptured, onCancel }: CaptureDirectionFormProps) {
  const { data: parentEC } = useEnterpriseCapability(enterpriseCapabilityId);
  const defaultResultingName = parentEC?.name ?? '';
  const { data: links, isLoading: linksLoading } = useEnterpriseCapabilityLinks(enterpriseCapabilityId);
  const { data: domainsResponse, isLoading: domainsLoading } = useBusinessDomainsQuery();
  const linkedCapabilities = useMemo(() => links ?? [], [links]);
  const businessDomains = useMemo(() => domainsResponse?.data ?? [], [domainsResponse]);

  const { form, placementsArray, onSubmit, isPending } = useCaptureDirectionFormState(
    enterpriseCapabilityId,
    onCaptured,
  );
  const { control, handleSubmit, watch, formState } = form;
  const type = watch('type');
  const placements = watch('placements');

  return (
    <form onSubmit={handleSubmit(onSubmit)} data-testid="capture-direction-form">
      <Stack gap="md">
        <TypeField control={control} />

        <Controller
          name="sourceCapabilityIds"
          control={control}
          render={({ field }) => (
            <SourcePicker
              loading={linksLoading}
              links={linkedCapabilities}
              selectedIds={field.value}
              error={formState.errors.sourceCapabilityIds?.message ?? null}
              onChange={field.onChange}
            />
          )}
        />

        {type !== 'stay' && (
          <PlacementEditor
            loading={domainsLoading}
            placements={placements}
            error={formState.errors.placements?.message ?? null}
            canAdd={canAddPlacement(type, placementsArray.fields.length)}
            onAdd={() =>
              placementsArray.append({ targetBusinessDomainId: '', resultingName: defaultResultingName })
            }
            renderRow={(index) => (
              <PlacementRow
                key={placementsArray.fields[index].id}
                index={index}
                control={control}
                domains={businessDomains}
                onRemove={() => placementsArray.remove(index)}
              />
            )}
          />
        )}

        <HorizonField control={control} />
        <NarrativeField control={control} />

        <Group justify="flex-end" gap="sm">
          <Button variant="default" onClick={onCancel} disabled={isPending}>
            Cancel
          </Button>
          <Button type="submit" loading={isPending} disabled={!formState.isValid || isPending}>
            Capture as draft
          </Button>
        </Group>
      </Stack>
    </form>
  );
}

type FormControl = ReturnType<typeof useForm<CaptureDirectionFormData>>['control'];

type EnumFieldName = 'type' | 'horizon';

interface EnumFieldProps<TValue extends string> {
  name: EnumFieldName;
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

function TypeField({ control }: { control: FormControl }) {
  return <EnumField name="type" label="Direction type" control={control} options={TYPE_OPTIONS} />;
}

function HorizonField({ control }: { control: FormControl }) {
  return <EnumField name="horizon" label="Horizon" control={control} options={HORIZON_OPTIONS} />;
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

interface SourcePickerProps {
  loading: boolean;
  links: EnterpriseCapabilityLink[];
  selectedIds: string[];
  error: string | null;
  onChange: (selectedIds: string[]) => void;
}

function SourcePicker({ loading, links, selectedIds, error, onChange }: SourcePickerProps) {
  return (
    <Stack gap="xs" data-testid="source-picker">
      <Text size="sm" fw={600}>
        Source physical capabilities
      </Text>
      {loading && <Loader size="sm" />}
      {!loading && links.length === 0 && (
        <Alert color="yellow" variant="light">
          This Enterprise Capability has no linked physical capabilities yet. Link some on the Manage Links page first.
        </Alert>
      )}
      {!loading && links.length > 0 && (
        <Checkbox.Group value={selectedIds} onChange={onChange}>
          <Stack gap={4}>
            {links.map((link) => (
              <Checkbox key={link.id} value={link.domainCapabilityId} label={<SourceLabel link={link} />} />
            ))}
          </Stack>
        </Checkbox.Group>
      )}
      {error && (
        <Text c="red" size="xs" data-testid="source-error">
          {error}
        </Text>
      )}
    </Stack>
  );
}

function SourceLabel({ link }: { link: EnterpriseCapabilityLink }) {
  return (
    <>
      {link.domainCapabilityName || link.domainCapabilityId}
      {link.businessDomainName && (
        <Text component="span" c="dimmed" size="xs">
          {' · '}
          {link.businessDomainName}
        </Text>
      )}
    </>
  );
}

interface PlacementEditorProps {
  loading: boolean;
  placements: CaptureDirectionFormData['placements'];
  error: string | null;
  canAdd: boolean;
  onAdd: () => void;
  renderRow: (index: number) => React.ReactNode;
}

function PlacementEditor({ loading, placements, error, canAdd, onAdd, renderRow }: PlacementEditorProps) {
  return (
    <Stack gap="xs" data-testid="placement-editor">
      <Text size="sm" fw={600}>
        Target placements
      </Text>
      {loading && <Loader size="sm" />}
      {!loading && placements.map((_p, index) => renderRow(index))}
      {canAdd && (
        <Group justify="flex-start">
          <Button variant="subtle" size="xs" onClick={onAdd} data-testid="add-placement">
            + Add placement
          </Button>
        </Group>
      )}
      {error && (
        <Text c="red" size="xs" data-testid="placement-error">
          {error}
        </Text>
      )}
    </Stack>
  );
}

interface PlacementRowProps {
  index: number;
  control: FormControl;
  domains: { id: string; name: string }[];
  onRemove: () => void;
}

function PlacementRow({ index, control, domains, onRemove }: PlacementRowProps) {
  const position = index + 1;
  return (
    <Group gap="xs" wrap="nowrap" align="flex-end">
      <Box flex={1}>
        <Controller
          name={`placements.${index}.targetBusinessDomainId`}
          control={control}
          render={({ field }) => (
            <Select
              placeholder="Select a business domain…"
              aria-label={`Target business domain for placement ${position}`}
              data={domains.map((d) => ({ value: d.id, label: d.name }))}
              value={field.value || null}
              onChange={(value) => field.onChange(value ?? '')}
            />
          )}
        />
      </Box>
      <Box flex={1}>
        <Controller
          name={`placements.${index}.resultingName`}
          control={control}
          render={({ field }) => (
            <TextInput
              aria-label={`Resulting name for placement ${position}`}
              placeholder="Resulting name"
              value={field.value ?? ''}
              onChange={(e) => field.onChange(e.currentTarget.value)}
            />
          )}
        />
      </Box>
      <ActionIcon variant="subtle" color="red" aria-label={`Remove placement ${position}`} onClick={onRemove}>
        ×
      </ActionIcon>
    </Group>
  );
}
