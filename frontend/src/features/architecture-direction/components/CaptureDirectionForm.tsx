import {
  ActionIcon,
  Alert,
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
import { useMemo, useState } from 'react';
import type { EnterpriseCapabilityId } from '../../../api/types';
import { useBusinessDomainsQuery } from '../../business-domains/hooks/useBusinessDomains';
import {
  useEnterpriseCapability,
  useEnterpriseCapabilityLinks,
} from '../../enterprise-architecture/hooks/useEnterpriseCapabilities';
import type { EnterpriseCapabilityLink } from '../../enterprise-architecture/types';
import { useCaptureDirection } from '../hooks/useDirection';
import type { CaptureDirectionRequest, DirectionType, Horizon, PlacementInput } from '../types';

interface CaptureDirectionFormProps {
  enterpriseCapabilityId: EnterpriseCapabilityId;
  onCaptured: () => void;
  onCancel: () => void;
}

const TYPE_OPTIONS = [
  { value: 'consolidate', label: 'Consolidate', help: 'Multiple physical capabilities merge into one. Pick 2 or more sources.' },
  { value: 'decompose', label: 'Decompose', help: 'One physical capability splits into multiple. Pick exactly 1 source.' },
  { value: 'stay', label: 'Stay', help: 'Explicitly confirmed no change. Pick exactly 1 source. No target placements.' },
] as const satisfies ReadonlyArray<{ value: DirectionType; label: string; help: string }>;

const HORIZON_OPTIONS = [
  { value: 'now', label: 'Now' },
  { value: 'next', label: 'Next' },
  { value: 'later', label: 'Later' },
] as const satisfies ReadonlyArray<{ value: Horizon; label: string }>;

interface FormState {
  type: DirectionType;
  selectedSourceIds: string[];
  placements: PlacementInput[];
  horizon: Horizon;
  narrative: string;
}

const INITIAL_STATE: FormState = {
  type: 'consolidate',
  selectedSourceIds: [],
  placements: [],
  horizon: 'next',
  narrative: '',
};

function requiresExactlyOneSource(type: DirectionType): boolean {
  return type === 'decompose' || type === 'stay';
}

function canAddPlacement(type: DirectionType, currentCount: number): boolean {
  if (type === 'consolidate') return currentCount === 0;
  if (type === 'decompose') return true;
  return false;
}

function describeSourceRequirement(type: DirectionType, count: number): string | null {
  if (type === 'consolidate' && count < 2) {
    return 'Consolidate requires at least 2 source physical capabilities.';
  }
  if (requiresExactlyOneSource(type) && count !== 1) {
    return `${type === 'decompose' ? 'Decompose' : 'Stay'} requires exactly 1 source physical capability.`;
  }
  return null;
}

function describePlacementRequirement(type: DirectionType, placements: PlacementInput[]): string | null {
  if (type === 'stay' && placements.length > 0) {
    return 'Stay directions carry no placements.';
  }
  if (type === 'consolidate' && placements.length !== 1) {
    return 'Consolidate requires exactly one target placement (N physicals merge into 1).';
  }
  if (type === 'decompose' && placements.length === 0) {
    return 'Decompose requires at least one target placement.';
  }
  for (const p of placements) {
    if (!p.targetBusinessDomainId) {
      return 'Every placement needs a target business domain.';
    }
  }
  return null;
}

function buildCaptureRequest(state: FormState): CaptureDirectionRequest {
  return {
    type: state.type,
    sourceCapabilityIds: state.selectedSourceIds,
    placements: state.type === 'stay' ? [] : state.placements,
    horizon: state.horizon,
    narrative: state.narrative.trim() || undefined,
  };
}

interface PlacementOps {
  add: () => void;
  update: (index: number, patch: Partial<PlacementInput>) => void;
  remove: (index: number) => void;
}

function usePlacementOps(
  setState: React.Dispatch<React.SetStateAction<FormState>>,
  defaultResultingName: string,
): PlacementOps {
  return {
    add: () =>
      setState((s) => ({
        ...s,
        placements: [...s.placements, { targetBusinessDomainId: '', resultingName: defaultResultingName }],
      })),
    update: (index, patch) =>
      setState((s) => ({
        ...s,
        placements: s.placements.map((p, i) => (i === index ? { ...p, ...patch } : p)),
      })),
    remove: (index) =>
      setState((s) => ({ ...s, placements: s.placements.filter((_, i) => i !== index) })),
  };
}

export function CaptureDirectionForm({ enterpriseCapabilityId, onCaptured, onCancel }: CaptureDirectionFormProps) {
  const { data: parentEC } = useEnterpriseCapability(enterpriseCapabilityId);
  const defaultResultingName = parentEC?.name ?? '';
  const { data: links, isLoading: linksLoading } = useEnterpriseCapabilityLinks(enterpriseCapabilityId);
  const { data: domainsResponse, isLoading: domainsLoading } = useBusinessDomainsQuery();
  const captureMutation = useCaptureDirection();

  const [state, setState] = useState<FormState>(INITIAL_STATE);
  const linkedCapabilities = useMemo(() => links ?? [], [links]);
  const businessDomains = useMemo(() => domainsResponse?.data ?? [], [domainsResponse]);

  const sourceErr = describeSourceRequirement(state.type, state.selectedSourceIds.length);
  const placementErr = describePlacementRequirement(state.type, state.placements);
  const submittable = !sourceErr && !placementErr && !captureMutation.isPending;
  const placementOps = usePlacementOps(setState, defaultResultingName);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!submittable) return;
    try {
      await captureMutation.mutateAsync({ enterpriseCapabilityId, request: buildCaptureRequest(state) });
      onCaptured();
    } catch {
      // toast handled by hook
    }
  };

  return (
    <form onSubmit={handleSubmit} data-testid="capture-direction-form">
      <Stack gap="md">
        <TypeField
          type={state.type}
          onChange={(type) =>
            setState((s) => ({
              ...s,
              type,
              placements: type === 'stay' ? [] : s.placements,
            }))
          }
        />

        <SourcePicker
          loading={linksLoading}
          links={linkedCapabilities}
          selectedIds={state.selectedSourceIds}
          error={sourceErr}
          onChange={(selectedSourceIds) => setState((s) => ({ ...s, selectedSourceIds }))}
        />

        {state.type !== 'stay' && (
          <PlacementEditor
            loading={domainsLoading}
            placements={state.placements}
            domains={businessDomains}
            error={placementErr}
            canAdd={canAddPlacement(state.type, state.placements.length)}
            ops={placementOps}
          />
        )}

        <HorizonField value={state.horizon} onChange={(horizon) => setState({ ...state, horizon })} />

        <NarrativeField
          value={state.narrative}
          onChange={(narrative) => setState({ ...state, narrative })}
        />

        <Group justify="flex-end" gap="sm">
          <Button variant="default" onClick={onCancel} disabled={captureMutation.isPending}>
            Cancel
          </Button>
          <Button type="submit" loading={captureMutation.isPending} disabled={!submittable}>
            Capture as draft
          </Button>
        </Group>
      </Stack>
    </form>
  );
}

interface TypeFieldProps {
  type: DirectionType;
  onChange: (type: DirectionType) => void;
}

function TypeField({ type, onChange }: TypeFieldProps) {
  const help = TYPE_OPTIONS.find((o) => o.value === type)?.help;
  return (
    <Select
      label="Direction type"
      data={TYPE_OPTIONS.map((o) => ({ value: o.value, label: o.label }))}
      value={type}
      onChange={(value) => value && onChange(value as DirectionType)}
      description={help}
      allowDeselect={false}
      withAsterisk
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
  placements: PlacementInput[];
  domains: { id: string; name: string }[];
  error: string | null;
  canAdd: boolean;
  ops: PlacementOps;
}

function PlacementEditor({ loading, placements, domains, error, canAdd, ops }: PlacementEditorProps) {
  return (
    <Stack gap="xs" data-testid="placement-editor">
      <Text size="sm" fw={600}>
        Target placements
      </Text>
      {loading && <Loader size="sm" />}
      {!loading &&
        placements.map((placement, index) => (
          // biome-ignore lint/suspicious/noArrayIndexKey: placements are an editable ordered list with no stable id
          <PlacementRow key={index} index={index} placement={placement} domains={domains} ops={ops} />
        ))}
      {canAdd && (
        <Button
          variant="subtle"
          size="xs"
          onClick={ops.add}
          data-testid="add-placement"
          style={{ alignSelf: 'flex-start' }}
        >
          + Add placement
        </Button>
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
  placement: PlacementInput;
  domains: { id: string; name: string }[];
  ops: PlacementOps;
}

function PlacementRow({ index, placement, domains, ops }: PlacementRowProps) {
  const position = index + 1;
  return (
    <Group gap="xs" wrap="nowrap" align="flex-end">
      <Select
        style={{ flex: 1 }}
        placeholder="Select a business domain…"
        aria-label={`Target business domain for placement ${position}`}
        data={domains.map((d) => ({ value: d.id, label: d.name }))}
        value={placement.targetBusinessDomainId || null}
        onChange={(value) => ops.update(index, { targetBusinessDomainId: value ?? '' })}
      />
      <TextInput
        style={{ flex: 1 }}
        aria-label={`Resulting name for placement ${position}`}
        placeholder="Resulting name"
        value={placement.resultingName ?? ''}
        onChange={(e) => ops.update(index, { resultingName: e.currentTarget.value })}
      />
      <ActionIcon
        variant="subtle"
        color="red"
        aria-label={`Remove placement ${position}`}
        onClick={() => ops.remove(index)}
      >
        ×
      </ActionIcon>
    </Group>
  );
}

interface HorizonFieldProps {
  value: Horizon;
  onChange: (h: Horizon) => void;
}

function HorizonField({ value, onChange }: HorizonFieldProps) {
  return (
    <Select
      label="Horizon"
      data={HORIZON_OPTIONS.map((o) => ({ value: o.value, label: o.label }))}
      value={value}
      onChange={(v) => v && onChange(v as Horizon)}
      allowDeselect={false}
      withAsterisk
    />
  );
}

interface NarrativeFieldProps {
  value: string;
  onChange: (v: string) => void;
}

function NarrativeField({ value, onChange }: NarrativeFieldProps) {
  return (
    <Textarea
      label="Narrative"
      description="Required before advancing the direction to proposed."
      placeholder="One or two sentences naming what the group decided and why"
      autosize
      minRows={3}
      value={value}
      onChange={(e) => onChange(e.currentTarget.value)}
    />
  );
}
