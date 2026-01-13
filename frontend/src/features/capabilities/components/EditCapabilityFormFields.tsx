import React from 'react';
import {
  TextInput,
  Textarea,
  Select,
  SimpleGrid,
  Box,
  Badge,
  Text,
  Stack,
  Group,
  Button,
} from '@mantine/core';
import { Controller, type Control, type UseFormRegister, type FieldErrors } from 'react-hook-form';
import type { EditCapabilityFormData } from '../../../lib/schemas';
import type { Expert } from '../../../api/types';
import { MaturitySlider } from '../../../components/shared/MaturitySlider';

interface SelectOption {
  value: string;
  label: string;
}

interface BasicFieldsProps {
  register: UseFormRegister<EditCapabilityFormData>;
  errors: FieldErrors<EditCapabilityFormData>;
  disabled: boolean;
}

export const BasicFields: React.FC<BasicFieldsProps> = ({ register, errors, disabled }) => (
  <>
    <TextInput
      label="Name"
      placeholder="Enter capability name"
      {...register('name')}
      required
      withAsterisk
      autoFocus
      disabled={disabled}
      error={errors.name?.message}
      data-testid="edit-capability-name-input"
    />

    <Textarea
      label="Description"
      placeholder="Enter capability description (optional)"
      {...register('description')}
      rows={3}
      disabled={disabled}
      error={errors.description?.message}
      data-testid="edit-capability-description-input"
    />
  </>
);

interface StatusFieldProps {
  control: Control<EditCapabilityFormData>;
  options: SelectOption[];
  isLoading: boolean;
  disabled: boolean;
}

export const StatusField: React.FC<StatusFieldProps> = ({
  control,
  options,
  isLoading,
  disabled,
}) => (
  <Controller
    name="status"
    control={control}
    render={({ field }) => (
      <Select
        label="Status"
        data={isLoading ? [] : options}
        disabled={disabled || isLoading}
        data-testid="edit-capability-status-select"
        {...field}
      />
    )}
  />
);

interface MaturityFieldProps {
  control: Control<EditCapabilityFormData>;
  disabled: boolean;
}

export const MaturityField: React.FC<MaturityFieldProps> = ({ control, disabled }) => (
  <Controller
    name="maturityValue"
    control={control}
    render={({ field }) => (
      <MaturitySlider value={field.value} onChange={field.onChange} disabled={disabled} />
    )}
  />
);

interface OwnershipFieldsProps {
  control: Control<EditCapabilityFormData>;
  register: UseFormRegister<EditCapabilityFormData>;
  ownershipOptions: SelectOption[];
  userOptions: SelectOption[];
  isLoadingOwnership: boolean;
  isLoadingUsers: boolean;
  disabled: boolean;
}

export const OwnershipFields: React.FC<OwnershipFieldsProps> = ({
  control,
  register,
  ownershipOptions,
  userOptions,
  isLoadingOwnership,
  isLoadingUsers,
  disabled,
}) => (
  <>
    <SimpleGrid cols={2}>
      <Controller
        name="ownershipModel"
        control={control}
        render={({ field }) => (
          <Select
            label="Ownership Model"
            placeholder="Select ownership model"
            data={isLoadingOwnership ? [] : ownershipOptions}
            disabled={disabled || isLoadingOwnership}
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
        disabled={disabled}
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
          disabled={disabled || isLoadingUsers}
          clearable
          searchable
          data-testid="edit-capability-ea-owner-select"
          {...field}
          value={field.value || null}
        />
      )}
    />
  </>
);

interface ExpertsListProps {
  experts?: Expert[];
  onAddClick: () => void;
  disabled?: boolean;
}

export const ExpertsList: React.FC<ExpertsListProps> = ({ experts, onAddClick, disabled }) => (
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

export const TagsList: React.FC<TagsListProps> = ({ tags, onAddClick, disabled }) => (
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
