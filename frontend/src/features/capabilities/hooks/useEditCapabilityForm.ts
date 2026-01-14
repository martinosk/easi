import { useState, useEffect, useMemo } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import {
  useCapabilities,
  useUpdateCapability,
  useUpdateCapabilityMetadata,
} from './useCapabilities';
import { useStatuses, useOwnershipModels } from '../../../hooks/useMetadata';
import { useMaturityScale } from '../../../hooks/useMaturityScale';
import { useEAOwnerCandidates } from '../../users/hooks/useUsers';
import { editCapabilitySchema, type EditCapabilityFormData } from '../../../lib/schemas';
import type { Capability } from '../../../api/types';
import { deriveLegacyMaturityValue, getDefaultSections, getMaturityBounds } from '../../../utils/maturity';

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

interface SelectOption {
  value: string;
  label: string;
}

const createDefaultValues = (
  cap?: Capability,
  sections = getDefaultSections()
): EditCapabilityFormData => {
  if (!cap) {
    return {
      name: '',
      description: '',
      status: 'Active',
      maturityValue: 12,
      ownershipModel: '',
      primaryOwner: '',
      eaOwner: '',
    };
  }

  let maturityValue = cap.maturityValue ?? 12;
  if (maturityValue === undefined && cap.maturityLevel) {
    maturityValue = deriveLegacyMaturityValue(cap.maturityLevel, sections);
  }

  return {
    name: cap.name ?? '',
    description: cap.description ?? '',
    status: cap.status ?? 'Active',
    maturityValue,
    ownershipModel: cap.ownershipModel ?? '',
    primaryOwner: cap.primaryOwner ?? '',
    eaOwner: cap.eaOwner ?? '',
  };
};

export interface UseEditCapabilityFormResult {
  form: ReturnType<typeof useForm<EditCapabilityFormData>>;
  currentCapability: Capability | null;
  statusOptions: SelectOption[];
  ownershipOptions: SelectOption[];
  userOptions: SelectOption[];
  isSaving: boolean;
  isLoadingMetadata: boolean;
  backendError: string | null;
  handleSubmit: (data: EditCapabilityFormData) => Promise<void>;
  clearError: () => void;
}

export function useEditCapabilityForm(
  capability: Capability | null,
  isOpen: boolean,
  onSuccess: () => void
): UseEditCapabilityFormResult {
  const [backendError, setBackendError] = useState<string | null>(null);
  const [currentCapability, setCurrentCapability] = useState<Capability | null>(null);

  const { data: maturityScale } = useMaturityScale();
  const { data: statusesData, isLoading: isLoadingStatuses } = useStatuses();
  const { data: ownershipModelsData, isLoading: isLoadingOwnershipModels } = useOwnershipModels();
  const { data: usersData, isLoading: isLoadingUsers } = useEAOwnerCandidates();
  const { data: capabilities = [] } = useCapabilities();

  const updateCapabilityMutation = useUpdateCapability();
  const updateMetadataMutation = useUpdateCapabilityMetadata();

  const sections = maturityScale?.sections ?? getDefaultSections();
  const statuses = statusesData ?? DEFAULT_STATUSES;
  const ownershipModels = ownershipModelsData ?? DEFAULT_OWNERSHIP_MODELS;

  const maturityBounds = useMemo(() => getMaturityBounds(sections), [sections]);
  const schema = useMemo(() => editCapabilitySchema(maturityBounds), [maturityBounds]);

  const form = useForm<EditCapabilityFormData>({
    resolver: zodResolver(schema),
    defaultValues: createDefaultValues(),
    mode: 'onChange',
  });

  useEffect(() => {
    if (capability) {
      setCurrentCapability(capabilities.find((c) => c.id === capability.id) || capability);
    }
  }, [capability, capabilities]);

  useEffect(() => {
    if (isOpen && capability) {
      form.reset(createDefaultValues(capability, sections));
      setBackendError(null);
    }
  }, [isOpen, capability, sections, form]);

  const statusOptions = useMemo(
    () =>
      [...statuses]
        .sort((a, b) => a.sortOrder - b.sortOrder)
        .map((s) => ({ value: s.value, label: s.displayName })),
    [statuses]
  );

  const ownershipOptions = useMemo(
    () => ownershipModels.map((om) => ({ value: om.value, label: om.displayName })),
    [ownershipModels]
  );

  const userOptions = useMemo(
    () => (usersData ?? []).map((u) => ({ value: u.id, label: u.name || u.email })),
    [usersData]
  );

  const handleSubmit = async (data: EditCapabilityFormData) => {
    if (!capability) return;
    setBackendError(null);
    try {
      await updateCapabilityMutation.mutateAsync({
        capability,
        request: {
          name: data.name,
          description: data.description || undefined,
        },
      });
      await updateMetadataMutation.mutateAsync({
        id: capability.id,
        request: {
          status: data.status,
          maturityValue: data.maturityValue,
          ownershipModel: data.ownershipModel || undefined,
          primaryOwner: data.primaryOwner || undefined,
          eaOwner: data.eaOwner || undefined,
        },
      });
      onSuccess();
    } catch (err) {
      setBackendError(err instanceof Error ? err.message : 'Failed to update capability');
    }
  };

  return {
    form,
    currentCapability,
    statusOptions,
    ownershipOptions,
    userOptions,
    isSaving: updateCapabilityMutation.isPending || updateMetadataMutation.isPending,
    isLoadingMetadata: isLoadingStatuses || isLoadingOwnershipModels || isLoadingUsers,
    backendError,
    handleSubmit,
    clearError: () => setBackendError(null),
  };
}
