import React, { useState, useEffect, useMemo } from 'react';
import {
  Modal,
  TextInput,
  Textarea,
  Select,
  Button,
  Group,
  Stack,
  Alert,
  SimpleGrid,
  Box,
  Badge,
  Text,
} from '@mantine/core';
import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import {
  useCapabilities,
  useUpdateCapability,
  useUpdateCapabilityMetadata,
} from '../hooks/useCapabilities';
import { useStatuses, useOwnershipModels } from '../../../hooks/useMetadata';
import { useMaturityScale } from '../../../hooks/useMaturityScale';
import { useActiveUsers } from '../../users/hooks/useUsers';
import { editCapabilitySchema, type EditCapabilityFormData } from '../../../lib/schemas';
import type { Capability, Expert } from '../../../api/types';
import { AddExpertDialog } from './AddExpertDialog';
import { AddTagDialog } from './AddTagDialog';
import { MaturitySlider } from '../../../components/shared/MaturitySlider';
import { deriveLegacyMaturityValue, getDefaultSections, getMaturityBounds } from '../../../utils/maturity';

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

interface ExpertsListProps {
  experts?: Expert[];
  onAddClick: () => void;
  disabled?: boolean;
}

const ExpertsList: React.FC<ExpertsListProps> = ({ experts, onAddClick, disabled }) => (
  <Box>
    <Text size="sm" fw={500} mb="xs">
      Experts
    </Text>
    {experts?.length ? (
      <Stack gap="xs">
        {experts.map((expert, i) => (
          <Text key={i} size="sm" c="dimmed">
            {expert.name} ({expert.role}) - {expert.contact}
          </Text>
        ))}
      </Stack>
    ) : (
      <Text size="sm" c="dimmed">
        No experts added
      </Text>
    )}
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
    <Text size="sm" fw={500} mb="xs">
      Tags
    </Text>
    {tags?.length ? (
      <Group gap="xs">
        {tags.map((tag, i) => (
          <Badge key={i} variant="light">
            {tag}
          </Badge>
        ))}
      </Group>
    ) : (
      <Text size="sm" c="dimmed">
        No tags added
      </Text>
    )}
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

export const EditCapabilityDialog: React.FC<EditCapabilityDialogProps> = ({
  isOpen,
  onClose,
  capability,
}) => {
  const [isAddExpertOpen, setIsAddExpertOpen] = useState(false);
  const [isAddTagOpen, setIsAddTagOpen] = useState(false);
  const [backendError, setBackendError] = useState<string | null>(null);
  const [currentCapability, setCurrentCapability] = useState<Capability | null>(null);

  const { data: maturityScale } = useMaturityScale();
  const { data: statusesData, isLoading: isLoadingStatuses } = useStatuses();
  const { data: ownershipModelsData, isLoading: isLoadingOwnershipModels } = useOwnershipModels();
  const { data: usersData, isLoading: isLoadingUsers } = useActiveUsers();
  const { data: capabilities = [] } = useCapabilities();

  const updateCapabilityMutation = useUpdateCapability();
  const updateMetadataMutation = useUpdateCapabilityMetadata();

  const sections = maturityScale?.sections ?? getDefaultSections();
  const statuses = statusesData ?? DEFAULT_STATUSES;
  const ownershipModels = ownershipModelsData ?? DEFAULT_OWNERSHIP_MODELS;

  const maturityBounds = useMemo(() => getMaturityBounds(sections), [sections]);
  const schema = useMemo(() => editCapabilitySchema(maturityBounds), [maturityBounds]);

  const {
    register,
    handleSubmit,
    control,
    reset,
    formState: { errors, isValid },
  } = useForm<EditCapabilityFormData>({
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
      reset(createDefaultValues(capability, sections));
      setBackendError(null);
    }
  }, [isOpen, capability, sections, reset]);

  const isSaving = updateCapabilityMutation.isPending || updateMetadataMutation.isPending;
  const isLoadingMetadata = isLoadingStatuses || isLoadingOwnershipModels || isLoadingUsers;

  const handleClose = () => {
    setBackendError(null);
    onClose();
  };

  const onSubmit = async (data: EditCapabilityFormData) => {
    if (!capability) return;
    setBackendError(null);
    try {
      await updateCapabilityMutation.mutateAsync({
        id: capability.id,
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
      handleClose();
    } catch (err) {
      setBackendError(err instanceof Error ? err.message : 'Failed to update capability');
    }
  };

  if (!capability) return null;

  const statusOptions = [...statuses]
    .sort((a, b) => a.sortOrder - b.sortOrder)
    .map((s) => ({ value: s.value, label: s.displayName }));

  const ownershipOptions = ownershipModels.map((om) => ({
    value: om.value,
    label: om.displayName,
  }));

  const userOptions = (usersData ?? []).map((u) => ({
    value: u.id,
    label: u.name || u.email,
  }));

  const displayCapability = currentCapability || capability;

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
        <form onSubmit={handleSubmit(onSubmit)}>
          <Stack gap="md">
            <TextInput
              label="Name"
              placeholder="Enter capability name"
              {...register('name')}
              required
              withAsterisk
              autoFocus
              disabled={isSaving}
              error={errors.name?.message}
              data-testid="edit-capability-name-input"
            />

            <Textarea
              label="Description"
              placeholder="Enter capability description (optional)"
              {...register('description')}
              rows={3}
              disabled={isSaving}
              error={errors.description?.message}
              data-testid="edit-capability-description-input"
            />

            <Controller
              name="status"
              control={control}
              render={({ field }) => (
                <Select
                  label="Status"
                  data={isLoadingStatuses ? [] : statusOptions}
                  disabled={isSaving || isLoadingStatuses}
                  data-testid="edit-capability-status-select"
                  {...field}
                />
              )}
            />

            <Controller
              name="maturityValue"
              control={control}
              render={({ field }) => (
                <MaturitySlider
                  value={field.value}
                  onChange={field.onChange}
                  disabled={isSaving}
                />
              )}
            />

            <SimpleGrid cols={2}>
              <Controller
                name="ownershipModel"
                control={control}
                render={({ field }) => (
                  <Select
                    label="Ownership Model"
                    placeholder="Select ownership model"
                    data={isLoadingOwnershipModels ? [] : ownershipOptions}
                    disabled={isSaving || isLoadingOwnershipModels}
                    clearable
                    data-testid="edit-capability-ownership-select"
                    {...field}
                    value={field.value || null}
                  />
                )}
              />

              <TextInput
                label="Primary Owner"
                placeholder="Enter primary owner"
                {...register('primaryOwner')}
                disabled={isSaving}
                data-testid="edit-capability-primary-owner-input"
              />
            </SimpleGrid>

            <Controller
              name="eaOwner"
              control={control}
              render={({ field }) => (
                <Select
                  label="EA Owner"
                  placeholder="Select EA owner"
                  data={isLoadingUsers ? [] : userOptions}
                  disabled={isSaving || isLoadingUsers}
                  clearable
                  searchable
                  data-testid="edit-capability-ea-owner-select"
                  {...field}
                  value={field.value || null}
                />
              )}
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
                disabled={isLoadingMetadata || !isValid}
                data-testid="edit-capability-submit"
              >
                Save
              </Button>
            </Group>
          </Stack>
        </form>
      </Modal>

      <AddExpertDialog
        isOpen={isAddExpertOpen}
        onClose={() => setIsAddExpertOpen(false)}
        capabilityId={capability.id}
      />

      <AddTagDialog
        isOpen={isAddTagOpen}
        onClose={() => setIsAddTagOpen(false)}
        capabilityId={capability.id}
      />
    </>
  );
};
