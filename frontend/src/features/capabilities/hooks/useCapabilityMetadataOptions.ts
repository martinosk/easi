import { useMemo } from 'react';
import { useOwnershipModels, useStatuses } from '../../../hooks/useMetadata';
import { useEAOwnerCandidates } from '../../users/hooks/useUsers';

export interface MetadataOption {
  value: string;
  label: string;
}

const DEFAULT_STATUSES = [
  { value: 'Active', displayName: 'Active', sortOrder: 1 },
  { value: 'Planned', displayName: 'Planned', sortOrder: 2 },
  { value: 'Deprecated', displayName: 'Deprecated', sortOrder: 3 },
];

const DEFAULT_OWNERSHIP_MODELS = [
  { value: 'TribeOwned', displayName: 'Tribe Owned' },
  { value: 'TeamOwned', displayName: 'Team Owned' },
  { value: 'Shared', displayName: 'Shared' },
  { value: 'EnterpriseService', displayName: 'Enterprise Service' },
];

export interface CapabilityMetadataOptions {
  statusOptions: MetadataOption[];
  ownershipOptions: MetadataOption[];
  eaOwnerOptions: MetadataOption[];
}

export function useCapabilityMetadataOptions(): CapabilityMetadataOptions {
  const { data: statuses } = useStatuses();
  const { data: ownershipModels } = useOwnershipModels();
  const { data: users } = useEAOwnerCandidates();

  const statusOptions = useMemo(
    () =>
      [...(statuses?.length ? statuses : DEFAULT_STATUSES)]
        .sort((a, b) => a.sortOrder - b.sortOrder)
        .map((s) => ({ value: s.value, label: s.displayName })),
    [statuses],
  );
  const ownershipOptions = useMemo(
    () =>
      (ownershipModels?.length ? ownershipModels : DEFAULT_OWNERSHIP_MODELS).map((m) => ({
        value: m.value,
        label: m.displayName,
      })),
    [ownershipModels],
  );
  const eaOwnerOptions = useMemo(() => (users ?? []).map((u) => ({ value: u.id, label: u.name || u.email })), [users]);

  return { statusOptions, ownershipOptions, eaOwnerOptions };
}
