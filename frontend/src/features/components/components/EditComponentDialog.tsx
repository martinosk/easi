import React, { useEffect, useState, useCallback } from 'react';
import { Modal, TextInput, Textarea, Button, Group, Stack, Alert } from '@mantine/core';
import { useForm } from 'react-hook-form';
import type { UseFormRegister, FieldErrors, UseFormReset } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { useUpdateComponent, useComponent } from '../hooks/useComponents';
import { editComponentSchema, type EditComponentFormData } from '../../../lib/schemas';
import { hasLink } from '../../../utils/hateoas';
import type { Component } from '../../../api/types';
import { ComponentExpertsList } from './ComponentExpertsList';
import { AddComponentExpertDialog } from './AddComponentExpertDialog';
import { toComponentId } from '../../../api/types';

interface EditComponentDialogProps {
  isOpen: boolean;
  onClose: () => void;
  component: Component | null;
}

const DEFAULT_VALUES: EditComponentFormData = { name: '', description: '' };

interface FormFieldsProps {
  register: UseFormRegister<EditComponentFormData>;
  errors: FieldErrors<EditComponentFormData>;
  isPending: boolean;
}

const FormFields: React.FC<FormFieldsProps> = ({ register, errors, isPending }) => (
  <>
    <TextInput
      label="Name"
      placeholder="Enter application name"
      {...register('name')}
      required
      withAsterisk
      autoFocus
      disabled={isPending}
      error={errors.name?.message}
    />
    <Textarea
      label="Description"
      placeholder="Enter application description (optional)"
      {...register('description')}
      rows={3}
      disabled={isPending}
      error={errors.description?.message}
    />
  </>
);

interface FormActionsProps {
  isPending: boolean;
  isValid: boolean;
  onCancel: () => void;
}

const FormActions: React.FC<FormActionsProps> = ({ isPending, isValid, onCancel }) => (
  <Group justify="flex-end" gap="sm">
    <Button variant="default" onClick={onCancel} disabled={isPending}>
      Cancel
    </Button>
    <Button type="submit" loading={isPending} disabled={!isValid}>
      Save Changes
    </Button>
  </Group>
);

function useFormReset(
  isOpen: boolean,
  component: Component | null,
  reset: UseFormReset<EditComponentFormData>,
  setBackendError: (error: string | null) => void,
) {
  useEffect(() => {
    if (!isOpen || !component) return;
    reset({ name: component.name, description: component.description || '' });
    setBackendError(null);
  }, [isOpen, component, reset, setBackendError]);
}

export const EditComponentDialog: React.FC<EditComponentDialogProps> = ({
  isOpen,
  onClose,
  component: componentProp,
}) => {
  const [backendError, setBackendError] = useState<string | null>(null);
  const [isAddExpertOpen, setIsAddExpertOpen] = useState(false);
  const updateComponentMutation = useUpdateComponent();

  const componentId = componentProp?.id ? toComponentId(componentProp.id) : undefined;
  const { data: freshComponent } = useComponent(componentId);
  const component = freshComponent ?? componentProp;

  const { register, handleSubmit, reset, formState: { errors, isValid } } = useForm<EditComponentFormData>({
    resolver: zodResolver(editComponentSchema),
    defaultValues: DEFAULT_VALUES,
    mode: 'onChange',
  });

  useFormReset(isOpen, componentProp, reset, setBackendError);

  const onSubmit = useCallback(async (data: EditComponentFormData) => {
    if (!component) return setBackendError('No application selected');
    setBackendError(null);
    try {
      await updateComponentMutation.mutateAsync({
        component,
        request: { name: data.name, description: data.description || undefined },
      });
      onClose();
    } catch (err) {
      setBackendError(err instanceof Error ? err.message : 'Failed to update application');
    }
  }, [component, updateComponentMutation, onClose]);

  const canAddExpert = component ? hasLink(component, 'x-add-expert') : false;
  const openAddExpert = useCallback(() => setIsAddExpertOpen(true), []);
  const closeAddExpert = useCallback(() => setIsAddExpertOpen(false), []);

  return (
    <>
      <Modal opened={isOpen} onClose={onClose} title="Edit Application" centered>
        <form onSubmit={handleSubmit(onSubmit)}>
          <Stack gap="md">
            <FormFields register={register} errors={errors} isPending={updateComponentMutation.isPending} />
            {component && (
              <ComponentExpertsList
                componentId={component.id}
                experts={component.experts}
                canAddExpert={canAddExpert}
                onAddClick={openAddExpert}
                disabled={updateComponentMutation.isPending}
              />
            )}
            {backendError && <Alert color="red">{backendError}</Alert>}
            <FormActions isPending={updateComponentMutation.isPending} isValid={isValid} onCancel={onClose} />
          </Stack>
        </form>
      </Modal>
      {component && <AddComponentExpertDialog isOpen={isAddExpertOpen} onClose={closeAddExpert} componentId={component.id} />}
    </>
  );
};
