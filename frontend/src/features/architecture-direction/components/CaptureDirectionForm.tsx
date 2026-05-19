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

const TYPE_OPTIONS: { value: DirectionType; label: string; help: string }[] = [
  { value: 'consolidate', label: 'Consolidate', help: 'Multiple physical capabilities merge into one. Pick 2 or more sources.' },
  { value: 'decompose', label: 'Decompose', help: 'One physical capability splits into multiple. Pick exactly 1 source.' },
  { value: 'stay', label: 'Stay', help: 'Explicitly confirmed no change. Pick exactly 1 source. No target placements.' },
];

const HORIZON_OPTIONS: { value: Horizon; label: string }[] = [
  { value: 'now', label: 'Now' },
  { value: 'next', label: 'Next' },
  { value: 'later', label: 'Later' },
];

interface FormState {
  type: DirectionType;
  selectedSourceIds: string[];
  placements: PlacementInput[];
  horizon: Horizon;
  narrative: string;
}

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

function useEnterpriseCapabilityName(id: EnterpriseCapabilityId): string {
  const { data } = useEnterpriseCapability(id);
  return data?.name ?? '';
}

function isReadyToSubmit(sourceErr: string | null, placementErr: string | null, isPending: boolean): boolean {
  return !sourceErr && !placementErr && !isPending;
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

export function CaptureDirectionForm({ enterpriseCapabilityId, onCaptured, onCancel }: CaptureDirectionFormProps) {
  const defaultResultingName = useEnterpriseCapabilityName(enterpriseCapabilityId);
  const { data: links, isLoading: linksLoading } = useEnterpriseCapabilityLinks(enterpriseCapabilityId);
  const { data: domainsResponse, isLoading: domainsLoading } = useBusinessDomainsQuery();
  const captureMutation = useCaptureDirection();

  const [state, setState] = useState<FormState>({
    type: 'consolidate',
    selectedSourceIds: [],
    placements: [],
    horizon: 'next',
    narrative: '',
  });

  const linkedCapabilities = useMemo(() => links ?? [], [links]);
  const businessDomains = useMemo(() => domainsResponse?.data ?? [], [domainsResponse]);

  const sourceErr = describeSourceRequirement(state.type, state.selectedSourceIds.length);
  const placementErr = describePlacementRequirement(state.type, state.placements);
  const submittable = isReadyToSubmit(sourceErr, placementErr, captureMutation.isPending);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!submittable) return;
    const request = buildCaptureRequest(state);
    try {
      await captureMutation.mutateAsync({ enterpriseCapabilityId, request });
      onCaptured();
    } catch {
      // toast handled by hook
    }
  };

  return (
    <form onSubmit={handleSubmit} data-testid="capture-direction-form" style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
      <TypeField type={state.type} onChange={(type) => setState({ ...state, type })} />

      <SourcePicker
        links={linkedCapabilities}
        loading={linksLoading}
        selectedIds={state.selectedSourceIds}
        error={sourceErr}
        onToggle={(capabilityId) =>
          setState((s) => ({
            ...s,
            selectedSourceIds: s.selectedSourceIds.includes(capabilityId)
              ? s.selectedSourceIds.filter((id) => id !== capabilityId)
              : [...s.selectedSourceIds, capabilityId],
          }))
        }
      />

      {state.type !== 'stay' && (
        <PlacementEditor
          placements={state.placements}
          domains={businessDomains}
          loading={domainsLoading}
          error={placementErr}
          defaultResultingName={defaultResultingName}
          canAddMore={canAddPlacement(state.type, state.placements.length)}
          onChange={(placements) => setState({ ...state, placements })}
        />
      )}

      <HorizonField value={state.horizon} onChange={(horizon) => setState({ ...state, horizon })} />
      <NarrativeField value={state.narrative} onChange={(narrative) => setState({ ...state, narrative })} />

      <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end' }}>
        <button type="button" onClick={onCancel}>
          Cancel
        </button>
        <button type="submit" disabled={!submittable}>
          {captureMutation.isPending ? 'Capturing…' : 'Capture as draft'}
        </button>
      </div>
    </form>
  );
}

function TypeField({ type, onChange }: { type: DirectionType; onChange: (t: DirectionType) => void }) {
  const help = TYPE_OPTIONS.find((o) => o.value === type)?.help;
  return (
    <div>
      <FieldLabel htmlFor="direction-type">Direction type</FieldLabel>
      <select
        id="direction-type"
        value={type}
        onChange={(e) => onChange(e.target.value as DirectionType)}
        style={{ width: '100%', padding: 6 }}
      >
        {TYPE_OPTIONS.map((opt) => (
          <option key={opt.value} value={opt.value}>
            {opt.label}
          </option>
        ))}
      </select>
      <p style={{ fontSize: 12, color: '#6B7280', margin: '4px 0 0' }}>{help}</p>
    </div>
  );
}

interface SourcePickerProps {
  links: EnterpriseCapabilityLink[];
  loading: boolean;
  selectedIds: string[];
  error: string | null;
  onToggle: (capabilityId: string) => void;
}

function SourcePicker({ links, loading, selectedIds, error, onToggle }: SourcePickerProps) {
  return (
    <fieldset style={{ border: '1px solid #E5E7EB', padding: '8px 12px', borderRadius: 4 }}>
      <legend style={{ fontWeight: 600 }}>Source physical capabilities</legend>
      {loading && <p style={{ color: '#6B7280', margin: '4px 0' }}>Loading linked capabilities…</p>}
      {!loading && links.length === 0 && (
        <p style={{ color: '#B45309', margin: '4px 0', fontSize: 13 }}>
          This Enterprise Capability has no linked physical capabilities yet. Link some on the Manage Links page first.
        </p>
      )}
      {!loading && links.length > 0 && (
        <ul data-testid="source-picker" style={{ listStyle: 'none', margin: 0, padding: 0, display: 'flex', flexDirection: 'column', gap: 4 }}>
          {links.map((link) => (
            <li key={link.id}>
              <label style={{ display: 'flex', alignItems: 'center', gap: 8, cursor: 'pointer' }}>
                <input
                  type="checkbox"
                  checked={selectedIds.includes(link.domainCapabilityId)}
                  onChange={() => onToggle(link.domainCapabilityId)}
                />
                <span>{link.domainCapabilityName || link.domainCapabilityId}</span>
                {link.businessDomainName && (
                  <span style={{ color: '#6B7280', fontSize: 12 }}>· {link.businessDomainName}</span>
                )}
              </label>
            </li>
          ))}
        </ul>
      )}
      {error && (
        <p data-testid="source-error" style={{ color: '#B91C1C', fontSize: 12, margin: '6px 0 0' }}>
          {error}
        </p>
      )}
    </fieldset>
  );
}

interface PlacementEditorProps {
  placements: PlacementInput[];
  domains: { id: string; name: string }[];
  loading: boolean;
  error: string | null;
  defaultResultingName: string;
  canAddMore: boolean;
  onChange: (placements: PlacementInput[]) => void;
}

function PlacementEditor({ placements, domains, loading, error, defaultResultingName, canAddMore, onChange }: PlacementEditorProps) {
  const updateAt = (index: number, patch: Partial<PlacementInput>) => {
    onChange(placements.map((p, i) => (i === index ? { ...p, ...patch } : p)));
  };
  const removeAt = (index: number) => onChange(placements.filter((_, i) => i !== index));
  const add = () => onChange([...placements, { targetBusinessDomainId: '', resultingName: defaultResultingName }]);

  return (
    <fieldset style={{ border: '1px solid #E5E7EB', padding: '8px 12px', borderRadius: 4 }}>
      <legend style={{ fontWeight: 600 }}>Target placements</legend>
      {loading && <p style={{ color: '#6B7280', margin: '4px 0' }}>Loading business domains…</p>}
      {!loading && (
        <div data-testid="placement-editor" style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
          {placements.map((placement, index) => (
            <div key={index} style={{ display: 'flex', gap: 6, alignItems: 'center' }}>
              <select
                value={placement.targetBusinessDomainId}
                onChange={(e) => updateAt(index, { targetBusinessDomainId: e.target.value })}
                style={{ flex: 1, padding: 6 }}
                aria-label={`Target business domain for placement ${index + 1}`}
              >
                <option value="">Select a business domain…</option>
                {domains.map((d) => (
                  <option key={d.id} value={d.id}>
                    {d.name}
                  </option>
                ))}
              </select>
              <input
                type="text"
                value={placement.resultingName ?? ''}
                onChange={(e) => updateAt(index, { resultingName: e.target.value })}
                placeholder="Resulting name"
                style={{ flex: 1, padding: 6 }}
                aria-label={`Resulting name for placement ${index + 1}`}
              />
              <button type="button" onClick={() => removeAt(index)} aria-label={`Remove placement ${index + 1}`}>
                ×
              </button>
            </div>
          ))}
          {canAddMore && (
            <button type="button" onClick={add} style={{ alignSelf: 'flex-start' }} data-testid="add-placement">
              + Add placement
            </button>
          )}
        </div>
      )}
      {error && (
        <p data-testid="placement-error" style={{ color: '#B91C1C', fontSize: 12, margin: '6px 0 0' }}>
          {error}
        </p>
      )}
    </fieldset>
  );
}

function HorizonField({ value, onChange }: { value: Horizon; onChange: (h: Horizon) => void }) {
  return (
    <div>
      <FieldLabel htmlFor="direction-horizon">Horizon</FieldLabel>
      <select
        id="direction-horizon"
        value={value}
        onChange={(e) => onChange(e.target.value as Horizon)}
        style={{ width: '100%', padding: 6 }}
      >
        {HORIZON_OPTIONS.map((opt) => (
          <option key={opt.value} value={opt.value}>
            {opt.label}
          </option>
        ))}
      </select>
    </div>
  );
}

function NarrativeField({ value, onChange }: { value: string; onChange: (v: string) => void }) {
  return (
    <div>
      <FieldLabel htmlFor="direction-narrative">Narrative</FieldLabel>
      <textarea
        id="direction-narrative"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder="One or two sentences naming what the group decided and why"
        rows={3}
        style={{ width: '100%', padding: 6 }}
      />
      <p style={{ fontSize: 12, color: '#6B7280', margin: '4px 0 0' }}>
        Required before advancing the direction to <em>proposed</em>.
      </p>
    </div>
  );
}

function FieldLabel({ htmlFor, children }: { htmlFor: string; children: React.ReactNode }) {
  return (
    <label htmlFor={htmlFor} style={{ display: 'block', fontWeight: 600, marginBottom: 4 }}>
      {children}
    </label>
  );
}
