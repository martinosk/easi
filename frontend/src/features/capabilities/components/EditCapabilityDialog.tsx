import React, { useState, useEffect } from 'react';
import { Modal, TextInput, Textarea, Select, NumberInput, Button, Group, Stack, Alert, SimpleGrid, Box, Badge, Text } from '@mantine/core';
import { useCapabilities, useUpdateCapability, useUpdateCapabilityMetadata } from '../hooks/useCapabilities';
import { useMaturityLevels, useStatuses, useOwnershipModels, useStrategyPillars } from '../../../hooks/useMetadata';
import type { Capability, Expert } from '../../../api/types';
import { AddExpertDialog } from './AddExpertDialog';
import { AddTagDialog } from './AddTagDialog';

interface EditCapabilityDialogProps {
  isOpen: boolean;
  onClose: () => void;
  capability: Capability | null;
}

const DEFAULT_MATURITY_LEVELS = ['Genesis', 'Custom Build', 'Product', 'Commodity'];
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
const DEFAULT_STRATEGY_PILLARS = [
  { value: 'AlwaysOn', displayName: 'Always On' },
  { value: 'Grow', displayName: 'Grow' },
  { value: 'Transform', displayName: 'Transform' },
];

interface FormState {
  name: string;
  description: string;
  status: string;
  maturityLevel: string;
  ownershipModel: string;
  primaryOwner: string;
  eaOwner: string;
  strategyPillar: string;
  pillarWeight: number;
}

interface FormErrors {
  name?: string;
  description?: string;
  pillarWeight?: string;
}

const validateForm = (form: FormState): FormErrors => {
  const errors: FormErrors = {};
  if (!form.name.trim()) {
    errors.name = 'Name is required';
  } else if (form.name.trim().length > 200) {
    errors.name = 'Name must be 200 characters or less';
  }
  if (form.description.length > 1000) {
    errors.description = 'Description must be 1000 characters or less';
  }
  if (form.pillarWeight < 0 || form.pillarWeight > 100) {
    errors.pillarWeight = 'Pillar weight must be between 0 and 100';
  }
  return errors;
};

const EMPTY_FORM: FormState = {
  name: '',
  description: '',
  status: 'Active',
  maturityLevel: '',
  ownershipModel: '',
  primaryOwner: '',
  eaOwner: '',
  strategyPillar: '',
  pillarWeight: 0,
};

const createInitialFormState = (cap?: Capability): FormState => {
  if (!cap) return EMPTY_FORM;
  return {
    name: cap.name ?? '',
    description: cap.description ?? '',
    status: cap.status ?? 'Active',
    maturityLevel: cap.maturityLevel ?? '',
    ownershipModel: cap.ownershipModel ?? '',
    primaryOwner: cap.primaryOwner ?? '',
    eaOwner: cap.eaOwner ?? '',
    strategyPillar: cap.strategyPillar ?? '',
    pillarWeight: cap.pillarWeight ?? 0,
  };
};

interface ExpertsListProps {
  experts?: Expert[];
  onAddClick: () => void;
  disabled?: boolean;
}

const ExpertsList: React.FC<ExpertsListProps> = ({ experts, onAddClick, disabled }) => (
  <Box>
    <Text size="sm" fw={500} mb="xs">Experts</Text>
    <Stack gap="xs">
      {experts && experts.length > 0 ? (
        experts.map((expert, index) => (
          <Text key={index} size="sm" c="dimmed">
            {expert.name} ({expert.role}) - {expert.contact}
          </Text>
        ))
      ) : (
        <Text size="sm" c="dimmed">No experts added</Text>
      )}
    </Stack>
    <Button
      variant="subtle"
      size="compact-sm"
      onClick={onAddClick}
      disabled={disabled}
      mt="xs"
      data-testid="add-expert-button"
    >
      + Add Expert
    </Button>
  </Box>
);

interface TagsListProps {
  tags?: string[];
  onAddClick: () => void;
  disabled?: boolean;
}

const TagsList: React.FC<TagsListProps> = ({ tags, onAddClick, disabled }) => (
  <Box>
    <Text size="sm" fw={500} mb="xs">Tags</Text>
    <Group gap="xs">
      {tags && tags.length > 0 ? (
        tags.map((tag, index) => (
          <Badge key={index} variant="light">{tag}</Badge>
        ))
      ) : (
        <Text size="sm" c="dimmed">No tags added</Text>
      )}
    </Group>
    <Button
      variant="subtle"
      size="compact-sm"
      onClick={onAddClick}
      disabled={disabled}
      mt="xs"
      data-testid="add-tag-button"
    >
      + Add Tag
    </Button>
  </Box>
);

export const EditCapabilityDialog: React.FC<EditCapabilityDialogProps> = ({ isOpen, onClose, capability }) => {
  const [form, setForm] = useState<FormState>(createInitialFormState());
  const [errors, setErrors] = useState<FormErrors>({});
  const [backendError, setBackendError] = useState<string | null>(null);
  const [isAddExpertOpen, setIsAddExpertOpen] = useState(false);
  const [isAddTagOpen, setIsAddTagOpen] = useState(false);
  const [currentCapability, setCurrentCapability] = useState<Capability | null>(null);

  const updateCapabilityMutation = useUpdateCapability();
  const updateCapabilityMetadataMutation = useUpdateCapabilityMetadata();
  const { data: capabilities = [] } = useCapabilities();

  const { data: maturityLevelsData, isLoading: isLoadingMaturityLevels } = useMaturityLevels();
  const { data: statusesData, isLoading: isLoadingStatuses } = useStatuses();
  const { data: ownershipModelsData, isLoading: isLoadingOwnershipModels } = useOwnershipModels();
  const { data: strategyPillarsData, isLoading: isLoadingStrategyPillars } = useStrategyPillars();

  const maturityLevels = maturityLevelsData ?? DEFAULT_MATURITY_LEVELS;
  const statuses = statusesData ?? DEFAULT_STATUSES;
  const ownershipModels = ownershipModelsData ?? DEFAULT_OWNERSHIP_MODELS;
  const strategyPillars = strategyPillarsData ?? DEFAULT_STRATEGY_PILLARS;

  useEffect(() => {
    if (capability) {
      const updated = capabilities.find((c) => c.id === capability.id);
      setCurrentCapability(updated || capability);
    }
  }, [capability, capabilities]);

  useEffect(() => {
    if (isOpen && capability) {
      setForm(createInitialFormState(capability));
      setErrors({});
      setBackendError(null);
    }
  }, [isOpen, capability]);

  const handleClose = () => {
    setErrors({});
    setBackendError(null);
    onClose();
  };

  const handleFieldChange = (field: keyof FormState, value: string | number) => {
    setForm((prev) => ({ ...prev, [field]: value }));
    if (errors[field as keyof FormErrors]) {
      setErrors((prev) => ({ ...prev, [field]: undefined }));
    }
    if (backendError) {
      setBackendError(null);
    }
  };

  const buildMetadataRequest = () => {
    const defaultMaturity = maturityLevels[0] ?? DEFAULT_MATURITY_LEVELS[0];
    return {
      status: form.status,
      maturityLevel: form.maturityLevel || defaultMaturity,
      ownershipModel: form.ownershipModel || undefined,
      primaryOwner: form.primaryOwner.trim() || undefined,
      eaOwner: form.eaOwner.trim() || undefined,
      strategyPillar: form.strategyPillar || undefined,
      pillarWeight: form.pillarWeight || undefined,
    };
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!capability) return;
    setBackendError(null);
    const validationErrors = validateForm(form);
    if (Object.keys(validationErrors).length > 0) {
      setErrors(validationErrors);
      return;
    }
    try {
      const description = form.description.trim() || undefined;
      await updateCapabilityMutation.mutateAsync({ id: capability.id, request: { name: form.name.trim(), description } });
      await updateCapabilityMetadataMutation.mutateAsync({ id: capability.id, request: buildMetadataRequest() });
      handleClose();
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to update capability';
      setBackendError(message);
    }
  };

  if (!capability) return null;
  const displayCapability = currentCapability || capability;

  const statusOptions = [...statuses]
    .sort((a, b) => a.sortOrder - b.sortOrder)
    .map((s) => ({
      value: s.value,
      label: s.displayName,
    }));

  const ownershipModelOptions = ownershipModels.map((om) => ({
    value: om.value,
    label: om.displayName,
  }));

  const strategyPillarOptions = strategyPillars.map((sp) => ({
    value: sp.value,
    label: sp.displayName,
  }));

  const isSaving = updateCapabilityMutation.isPending || updateCapabilityMetadataMutation.isPending;
  const isLoadingMetadata = isLoadingMaturityLevels || isLoadingStatuses || isLoadingOwnershipModels || isLoadingStrategyPillars;

  return (
    <>
      <Modal
        opened={isOpen}
        onClose={handleClose}
        title="Edit Capability"
        centered
        size="lg"
        data-testid="edit-capability-dialog"
      >
        <form onSubmit={handleSubmit}>
          <Stack gap="md">
            <TextInput
              label="Name"
              placeholder="Enter capability name"
              value={form.name}
              onChange={(e) => handleFieldChange('name', e.currentTarget.value)}
              required
              withAsterisk
              autoFocus
              disabled={isSaving}
              error={errors.name}
              data-testid="edit-capability-name-input"
            />

            <Textarea
              label="Description"
              placeholder="Enter capability description (optional)"
              value={form.description}
              onChange={(e) => handleFieldChange('description', e.currentTarget.value)}
              rows={3}
              disabled={isSaving}
              error={errors.description}
              data-testid="edit-capability-description-input"
            />

            <SimpleGrid cols={2}>
              <Select
                label="Status"
                value={form.status}
                onChange={(value) => handleFieldChange('status', value || 'Active')}
                data={isLoadingStatuses ? [] : statusOptions}
                disabled={isSaving || isLoadingStatuses}
                data-testid="edit-capability-status-select"
              />

              <Select
                label="Maturity Level"
                value={form.maturityLevel}
                onChange={(value) => handleFieldChange('maturityLevel', value || '')}
                data={isLoadingMaturityLevels ? [] : maturityLevels}
                disabled={isSaving || isLoadingMaturityLevels}
                data-testid="edit-capability-maturity-select"
              />
            </SimpleGrid>

            <SimpleGrid cols={2}>
              <Select
                label="Ownership Model"
                placeholder="Select ownership model"
                value={form.ownershipModel}
                onChange={(value) => handleFieldChange('ownershipModel', value || '')}
                data={isLoadingOwnershipModels ? [] : ownershipModelOptions}
                disabled={isSaving || isLoadingOwnershipModels}
                clearable
                data-testid="edit-capability-ownership-select"
              />

              <TextInput
                label="Primary Owner"
                placeholder="Enter primary owner"
                value={form.primaryOwner}
                onChange={(e) => handleFieldChange('primaryOwner', e.currentTarget.value)}
                disabled={isSaving}
                data-testid="edit-capability-primary-owner-input"
              />
            </SimpleGrid>

            <SimpleGrid cols={2}>
              <TextInput
                label="EA Owner"
                placeholder="Enter EA owner"
                value={form.eaOwner}
                onChange={(e) => handleFieldChange('eaOwner', e.currentTarget.value)}
                disabled={isSaving}
                data-testid="edit-capability-ea-owner-input"
              />

              <Select
                label="Strategy Pillar"
                placeholder="Select strategy pillar"
                value={form.strategyPillar}
                onChange={(value) => handleFieldChange('strategyPillar', value || '')}
                data={isLoadingStrategyPillars ? [] : strategyPillarOptions}
                disabled={isSaving || isLoadingStrategyPillars}
                clearable
                data-testid="edit-capability-strategy-pillar-select"
              />
            </SimpleGrid>

            <NumberInput
              label="Pillar Weight (0-100)"
              value={form.pillarWeight}
              onChange={(value) => handleFieldChange('pillarWeight', typeof value === 'number' ? value : 0)}
              min={0}
              max={100}
              disabled={isSaving}
              error={errors.pillarWeight}
              data-testid="edit-capability-pillar-weight-input"
            />

            <ExpertsList
              experts={displayCapability.experts}
              onAddClick={() => setIsAddExpertOpen(true)}
              disabled={isSaving}
            />

            <TagsList
              tags={displayCapability.tags}
              onAddClick={() => setIsAddTagOpen(true)}
              disabled={isSaving}
            />

            {backendError && (
              <Alert color="red" data-testid="edit-capability-error">
                {backendError}
              </Alert>
            )}

            <Group justify="flex-end" gap="sm">
              <Button
                variant="default"
                onClick={handleClose}
                disabled={isSaving}
                data-testid="edit-capability-cancel"
              >
                Cancel
              </Button>
              <Button
                type="submit"
                loading={isSaving}
                disabled={isLoadingMetadata || !form.name.trim()}
                data-testid="edit-capability-submit"
              >
                Save
              </Button>
            </Group>
          </Stack>
        </form>
      </Modal>

      <AddExpertDialog isOpen={isAddExpertOpen} onClose={() => setIsAddExpertOpen(false)} capabilityId={capability.id} />
      <AddTagDialog isOpen={isAddTagOpen} onClose={() => setIsAddTagOpen(false)} capabilityId={capability.id} />
    </>
  );
};
