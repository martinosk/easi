import React, { useState, useEffect } from 'react';
import { Modal, TextInput, Textarea, Select, Button, Group, Stack, Alert, SimpleGrid, Box, Badge, Text } from '@mantine/core';
import { useCapabilities, useUpdateCapability, useUpdateCapabilityMetadata } from '../hooks/useCapabilities';
import { useStatuses, useOwnershipModels } from '../../../hooks/useMetadata';
import { useMaturityScale } from '../../../hooks/useMaturityScale';
import { useActiveUsers } from '../../users/hooks/useUsers';
import type { Capability, Expert } from '../../../api/types';
import { AddExpertDialog } from './AddExpertDialog';
import { AddTagDialog } from './AddTagDialog';
import { MaturitySlider } from '../../../components/shared/MaturitySlider';
import { deriveLegacyMaturityValue, getDefaultSections } from '../../../utils/maturity';

interface EditCapabilityDialogProps {
  isOpen: boolean;
  onClose: () => void;
  capability: Capability | null;
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

interface FormState {
  name: string;
  description: string;
  status: string;
  maturityValue: number;
  ownershipModel: string;
  primaryOwner: string;
  eaOwner: string;
}

interface FormErrors {
  name?: string;
  description?: string;
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
  return errors;
};

const EMPTY_FORM: FormState = {
  name: '',
  description: '',
  status: 'Active',
  maturityValue: 12,
  ownershipModel: '',
  primaryOwner: '',
  eaOwner: '',
};

const createInitialFormState = (cap?: Capability, sections = getDefaultSections()): FormState => {
  if (!cap) return EMPTY_FORM;

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

interface ItemListSectionProps {
  label: string;
  emptyMessage: string;
  buttonLabel: string;
  testId: string;
  onAddClick: () => void;
  disabled?: boolean;
  children: React.ReactNode;
  hasItems: boolean;
}

const ItemListSection: React.FC<ItemListSectionProps> = ({
  label, emptyMessage, buttonLabel, testId, onAddClick, disabled, children, hasItems,
}) => (
  <Box>
    <Text size="sm" fw={500} mb="xs">{label}</Text>
    {hasItems ? children : <Text size="sm" c="dimmed">{emptyMessage}</Text>}
    <Button variant="subtle" size="compact-sm" onClick={onAddClick} disabled={disabled} mt="xs" data-testid={testId}>
      {buttonLabel}
    </Button>
  </Box>
);

const ExpertsList: React.FC<{ experts?: Expert[]; onAddClick: () => void; disabled?: boolean }> = ({ experts, onAddClick, disabled }) => (
  <ItemListSection
    label="Experts"
    emptyMessage="No experts added"
    buttonLabel="+ Add Expert"
    testId="add-expert-button"
    onAddClick={onAddClick}
    disabled={disabled}
    hasItems={!!experts?.length}
  >
    <Stack gap="xs">
      {experts?.map((expert, index) => (
        <Text key={index} size="sm" c="dimmed">{expert.name} ({expert.role}) - {expert.contact}</Text>
      ))}
    </Stack>
  </ItemListSection>
);

const TagsList: React.FC<{ tags?: string[]; onAddClick: () => void; disabled?: boolean }> = ({ tags, onAddClick, disabled }) => (
  <ItemListSection
    label="Tags"
    emptyMessage="No tags added"
    buttonLabel="+ Add Tag"
    testId="add-tag-button"
    onAddClick={onAddClick}
    disabled={disabled}
    hasItems={!!tags?.length}
  >
    <Group gap="xs">{tags?.map((tag, index) => <Badge key={index} variant="light">{tag}</Badge>)}</Group>
  </ItemListSection>
);

interface SelectOption {
  value: string;
  label: string;
}

interface BasicInfoFieldsProps {
  form: FormState;
  errors: FormErrors;
  isSaving: boolean;
  onFieldChange: (field: keyof FormState, value: string | number) => void;
}

const BasicInfoFields: React.FC<BasicInfoFieldsProps> = ({ form, errors, isSaving, onFieldChange }) => (
  <>
    <TextInput
      label="Name"
      placeholder="Enter capability name"
      value={form.name}
      onChange={(e) => onFieldChange('name', e.currentTarget.value)}
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
      onChange={(e) => onFieldChange('description', e.currentTarget.value)}
      rows={3}
      disabled={isSaving}
      error={errors.description}
      data-testid="edit-capability-description-input"
    />
  </>
);

interface StatusFieldsProps {
  form: FormState;
  statusOptions: SelectOption[];
  isLoadingStatuses: boolean;
  isSaving: boolean;
  onFieldChange: (field: keyof FormState, value: string | number) => void;
}

const StatusFields: React.FC<StatusFieldsProps> = ({
  form, statusOptions, isLoadingStatuses, isSaving, onFieldChange,
}) => (
  <Select
    label="Status"
    value={form.status}
    onChange={(value) => onFieldChange('status', value || 'Active')}
    data={isLoadingStatuses ? [] : statusOptions}
    disabled={isSaving || isLoadingStatuses}
    data-testid="edit-capability-status-select"
  />
);

interface OwnershipFieldsProps {
  form: FormState;
  ownershipModelOptions: SelectOption[];
  isLoadingOwnershipModels: boolean;
  isSaving: boolean;
  onFieldChange: (field: keyof FormState, value: string | number) => void;
}

const OwnershipFields: React.FC<OwnershipFieldsProps> = ({
  form, ownershipModelOptions, isLoadingOwnershipModels, isSaving, onFieldChange,
}) => (
  <SimpleGrid cols={2}>
    <Select
      label="Ownership Model"
      placeholder="Select ownership model"
      value={form.ownershipModel}
      onChange={(value) => onFieldChange('ownershipModel', value || '')}
      data={isLoadingOwnershipModels ? [] : ownershipModelOptions}
      disabled={isSaving || isLoadingOwnershipModels}
      clearable
      data-testid="edit-capability-ownership-select"
    />
    <TextInput
      label="Primary Owner"
      placeholder="Enter primary owner"
      value={form.primaryOwner}
      onChange={(e) => onFieldChange('primaryOwner', e.currentTarget.value)}
      disabled={isSaving}
      data-testid="edit-capability-primary-owner-input"
    />
  </SimpleGrid>
);

interface EAOwnerFieldProps {
  form: FormState;
  userOptions: SelectOption[];
  isLoadingUsers: boolean;
  isSaving: boolean;
  onFieldChange: (field: keyof FormState, value: string | number) => void;
}

const EAOwnerField: React.FC<EAOwnerFieldProps> = ({
  form, userOptions, isLoadingUsers, isSaving, onFieldChange,
}) => (
  <Select
    label="EA Owner"
    placeholder="Select EA owner"
    value={form.eaOwner}
    onChange={(value) => onFieldChange('eaOwner', value || '')}
    data={isLoadingUsers ? [] : userOptions}
    disabled={isSaving || isLoadingUsers}
    clearable
    searchable
    data-testid="edit-capability-ea-owner-select"
  />
);

function useEditCapabilityForm(capability: Capability | null, isOpen: boolean, onClose: () => void) {
  const [form, setForm] = useState<FormState>(createInitialFormState());
  const [errors, setErrors] = useState<FormErrors>({});
  const [backendError, setBackendError] = useState<string | null>(null);
  const [currentCapability, setCurrentCapability] = useState<Capability | null>(null);

  const updateCapabilityMutation = useUpdateCapability();
  const updateCapabilityMetadataMutation = useUpdateCapabilityMetadata();
  const { data: capabilities = [] } = useCapabilities();

  const { data: maturityScale } = useMaturityScale();
  const { data: statusesData, isLoading: isLoadingStatuses } = useStatuses();
  const { data: ownershipModelsData, isLoading: isLoadingOwnershipModels } = useOwnershipModels();
  const { data: usersData, isLoading: isLoadingUsers } = useActiveUsers();

  const sections = maturityScale?.sections ?? getDefaultSections();
  const statuses = statusesData ?? DEFAULT_STATUSES;
  const ownershipModels = ownershipModelsData ?? DEFAULT_OWNERSHIP_MODELS;

  useEffect(() => {
    if (capability) {
      const updated = capabilities.find((c) => c.id === capability.id);
      setCurrentCapability(updated || capability);
    }
  }, [capability, capabilities]);

  useEffect(() => {
    if (isOpen && capability) {
      setForm(createInitialFormState(capability, sections));
      setErrors({});
      setBackendError(null);
    }
  }, [isOpen, capability, sections]);

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
      const metadataRequest = {
        status: form.status,
        maturityValue: form.maturityValue,
        ownershipModel: form.ownershipModel || undefined,
        primaryOwner: form.primaryOwner.trim() || undefined,
        eaOwner: form.eaOwner || undefined,
      };
      await updateCapabilityMutation.mutateAsync({ id: capability.id, request: { name: form.name.trim(), description } });
      await updateCapabilityMetadataMutation.mutateAsync({ id: capability.id, request: metadataRequest });
      handleClose();
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to update capability';
      setBackendError(message);
    }
  };

  const statusOptions = [...statuses].sort((a, b) => a.sortOrder - b.sortOrder).map((s) => ({ value: s.value, label: s.displayName }));
  const ownershipModelOptions = ownershipModels.map((om) => ({ value: om.value, label: om.displayName }));
  const userOptions = (usersData ?? []).map((u) => ({ value: u.id, label: u.name || u.email }));

  return {
    form,
    errors,
    backendError,
    currentCapability,
    statusOptions,
    ownershipModelOptions,
    userOptions,
    isSaving: updateCapabilityMutation.isPending || updateCapabilityMetadataMutation.isPending,
    isLoadingMetadata: isLoadingStatuses || isLoadingOwnershipModels || isLoadingUsers,
    isLoadingStatuses,
    isLoadingOwnershipModels,
    isLoadingUsers,
    handleClose,
    handleFieldChange,
    handleSubmit,
  };
}

export const EditCapabilityDialog: React.FC<EditCapabilityDialogProps> = ({ isOpen, onClose, capability }) => {
  const [isAddExpertOpen, setIsAddExpertOpen] = useState(false);
  const [isAddTagOpen, setIsAddTagOpen] = useState(false);

  const formProps = useEditCapabilityForm(capability, isOpen, onClose);
  const { form, errors, backendError, currentCapability, isSaving, isLoadingMetadata, handleClose, handleFieldChange, handleSubmit } = formProps;

  if (!capability) return null;
  const displayCapability = currentCapability || capability;

  return (
    <>
      <Modal opened={isOpen} onClose={handleClose} title="Edit Capability" centered size="lg" data-testid="edit-capability-dialog">
        <form onSubmit={handleSubmit}>
          <Stack gap="md">
            <BasicInfoFields form={form} errors={errors} isSaving={isSaving} onFieldChange={handleFieldChange} />
            <StatusFields
              form={form}
              statusOptions={formProps.statusOptions}
              isLoadingStatuses={formProps.isLoadingStatuses}
              isSaving={isSaving}
              onFieldChange={handleFieldChange}
            />
            <MaturitySlider
              value={form.maturityValue}
              onChange={(value) => handleFieldChange('maturityValue', value)}
              disabled={isSaving}
            />
            <OwnershipFields
              form={form}
              ownershipModelOptions={formProps.ownershipModelOptions}
              isLoadingOwnershipModels={formProps.isLoadingOwnershipModels}
              isSaving={isSaving}
              onFieldChange={handleFieldChange}
            />
            <EAOwnerField
              form={form}
              userOptions={formProps.userOptions}
              isLoadingUsers={formProps.isLoadingUsers}
              isSaving={isSaving}
              onFieldChange={handleFieldChange}
            />
            <ExpertsList experts={displayCapability.experts} onAddClick={() => setIsAddExpertOpen(true)} disabled={isSaving} />
            <TagsList tags={displayCapability.tags} onAddClick={() => setIsAddTagOpen(true)} disabled={isSaving} />
            {backendError && <Alert color="red" data-testid="edit-capability-error">{backendError}</Alert>}
            <Group justify="flex-end" gap="sm">
              <Button variant="default" onClick={handleClose} disabled={isSaving} data-testid="edit-capability-cancel">Cancel</Button>
              <Button type="submit" loading={isSaving} disabled={isLoadingMetadata || !form.name.trim()} data-testid="edit-capability-submit">Save</Button>
            </Group>
          </Stack>
        </form>
      </Modal>
      <AddExpertDialog isOpen={isAddExpertOpen} onClose={() => setIsAddExpertOpen(false)} capabilityId={capability.id} />
      <AddTagDialog isOpen={isAddTagOpen} onClose={() => setIsAddTagOpen(false)} capabilityId={capability.id} />
    </>
  );
};
