import { zodResolver } from '@hookform/resolvers/zod';
import { Alert, Button, Group, Modal, Stack, Textarea, TextInput } from '@mantine/core';
import React, { useLayoutEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import {
  type CreateEnterpriseCapabilityFormData,
  createEnterpriseCapabilitySchema,
} from '../../../lib/schemas';
import type { CreateEnterpriseCapabilityRequest } from '../types';

interface CreateEnterpriseCapabilityModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (request: CreateEnterpriseCapabilityRequest) => Promise<void>;
}

const DEFAULT_VALUES: CreateEnterpriseCapabilityFormData = {
  name: '',
  description: '',
  category: '',
};

export const CreateEnterpriseCapabilityModal = React.memo<CreateEnterpriseCapabilityModalProps>(
  ({ isOpen, onClose, onSubmit }) => {
    const [error, setError] = useState<string | null>(null);
    const [isSubmitting, setIsSubmitting] = useState(false);

    const {
      register,
      handleSubmit,
      reset,
      formState: { errors, isValid },
    } = useForm<CreateEnterpriseCapabilityFormData>({
      resolver: zodResolver(createEnterpriseCapabilitySchema),
      defaultValues: DEFAULT_VALUES,
      mode: 'onChange',
    });

    useLayoutEffect(() => {
      if (!isOpen) return;
      reset(DEFAULT_VALUES);
      setError(null);
    }, [isOpen, reset]);

    const handleClose = () => {
      reset(DEFAULT_VALUES);
      setError(null);
      onClose();
    };

    const submit = async (data: CreateEnterpriseCapabilityFormData) => {
      setError(null);
      setIsSubmitting(true);
      try {
        await onSubmit({
          name: data.name,
          description: data.description || undefined,
          category: data.category || undefined,
        });
        reset(DEFAULT_VALUES);
        onClose();
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to create enterprise capability');
      } finally {
        setIsSubmitting(false);
      }
    };

    return (
      <Modal
        opened={isOpen}
        onClose={handleClose}
        title="Create Enterprise Capability"
        centered
        data-testid="create-capability-modal"
      >
        <form onSubmit={handleSubmit(submit)}>
          <Stack gap="md">
            <TextInput
              label="Name"
              placeholder="e.g., Customer Management"
              required
              withAsterisk
              autoFocus
              disabled={isSubmitting}
              error={errors.name?.message}
              data-testid="capability-name-input"
              {...register('name')}
            />
            <TextInput
              label="Category"
              placeholder="e.g., Core Business"
              disabled={isSubmitting}
              error={errors.category?.message}
              data-testid="capability-category-input"
              {...register('category')}
            />
            <Textarea
              label="Description"
              placeholder="Describe what this enterprise capability represents..."
              rows={3}
              disabled={isSubmitting}
              error={errors.description?.message}
              data-testid="capability-description-input"
              {...register('description')}
            />
            {error && (
              <Alert color="red" data-testid="create-error-message">
                {error}
              </Alert>
            )}
            <Group justify="flex-end" gap="sm">
              <Button
                variant="default"
                onClick={handleClose}
                disabled={isSubmitting}
                data-testid="create-cancel-btn"
              >
                Cancel
              </Button>
              <Button type="submit" loading={isSubmitting} disabled={!isValid} data-testid="create-submit-btn">
                Create
              </Button>
            </Group>
          </Stack>
        </form>
      </Modal>
    );
  },
);

CreateEnterpriseCapabilityModal.displayName = 'CreateEnterpriseCapabilityModal';
