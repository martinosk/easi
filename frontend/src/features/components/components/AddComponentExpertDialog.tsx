import React, { useEffect, useState } from 'react';
import { Modal, TextInput, Button, Group, Stack, Alert, Autocomplete } from '@mantine/core';
import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { useAddComponentExpert, useComponentExpertRoles } from '../hooks/useComponents';
import {
  addComponentExpertSchema,
  type AddComponentExpertFormData,
} from '../../../lib/schemas/component';
import { toComponentId } from '../../../api/types';

interface AddComponentExpertDialogProps {
  isOpen: boolean;
  onClose: () => void;
  componentId: string;
}

const DEFAULT_VALUES: AddComponentExpertFormData = { name: '', role: '', contact: '' };

export const AddComponentExpertDialog: React.FC<AddComponentExpertDialogProps> = ({
  isOpen,
  onClose,
  componentId,
}) => {
  const [backendError, setBackendError] = useState<string | null>(null);
  const addExpertMutation = useAddComponentExpert();
  const { data: expertRoles = [] } = useComponentExpertRoles();

  const {
    register,
    handleSubmit,
    reset,
    control,
    formState: { errors, isValid },
  } = useForm<AddComponentExpertFormData>({
    resolver: zodResolver(addComponentExpertSchema),
    defaultValues: DEFAULT_VALUES,
    mode: 'onChange',
  });

  useEffect(() => {
    if (isOpen) {
      reset(DEFAULT_VALUES);
      setBackendError(null);
    }
  }, [isOpen, reset]);

  const onSubmit = async (data: AddComponentExpertFormData) => {
    setBackendError(null);
    try {
      await addExpertMutation.mutateAsync({
        id: toComponentId(componentId),
        request: { expertName: data.name, expertRole: data.role, contactInfo: data.contact },
      });
      onClose();
    } catch (err) {
      setBackendError(err instanceof Error ? err.message : 'Failed to add expert');
    }
  };

  return (
    <Modal
      opened={isOpen}
      onClose={onClose}
      title="Add Expert"
      centered
      data-testid="add-component-expert-dialog"
    >
      <form onSubmit={handleSubmit(onSubmit)}>
        <Stack gap="md">
          <TextInput
            label="Name"
            placeholder="Enter expert name"
            {...register('name')}
            required
            withAsterisk
            autoFocus
            disabled={addExpertMutation.isPending}
            error={errors.name?.message}
            data-testid="component-expert-name-input"
          />

          <Controller
            name="role"
            control={control}
            render={({ field }) => (
              <Autocomplete
                label="Role"
                placeholder="Enter or select expert role"
                data={expertRoles}
                value={field.value}
                onChange={field.onChange}
                onBlur={field.onBlur}
                required
                withAsterisk
                disabled={addExpertMutation.isPending}
                error={errors.role?.message}
                data-testid="component-expert-role-input"
              />
            )}
          />

          <TextInput
            label="Contact"
            placeholder="Enter contact information"
            {...register('contact')}
            required
            withAsterisk
            disabled={addExpertMutation.isPending}
            error={errors.contact?.message}
            data-testid="component-expert-contact-input"
          />

          {backendError && (
            <Alert color="red" data-testid="add-component-expert-error">
              {backendError}
            </Alert>
          )}

          <Group justify="flex-end" gap="sm">
            <Button
              variant="default"
              onClick={onClose}
              disabled={addExpertMutation.isPending}
              data-testid="add-component-expert-cancel"
            >
              Cancel
            </Button>
            <Button
              type="submit"
              loading={addExpertMutation.isPending}
              disabled={!isValid}
              data-testid="add-component-expert-submit"
            >
              Add
            </Button>
          </Group>
        </Stack>
      </form>
    </Modal>
  );
};
