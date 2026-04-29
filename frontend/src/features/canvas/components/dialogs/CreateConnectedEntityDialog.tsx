import { zodResolver } from '@hookform/resolvers/zod';
import { Alert, Button, Group, Modal, Select, Stack, Text, Textarea, TextInput } from '@mantine/core';
import React, { useEffect, useMemo } from 'react';
import { Controller, useForm } from 'react-hook-form';
import { z } from 'zod';
import type { HATEOASLink, HATEOASLinks } from '../../../../api/types';
import { componentNameSchema, componentDescriptionSchema } from '../../../../lib/schemas';
import { relationTypeSchema } from '../../../../lib/schemas';

const RELATION_TYPE_OPTIONS = [
  { value: 'Triggers', label: 'Triggers' },
  { value: 'Serves', label: 'Serves' },
];

export type ConnectedEntityActionType =
  | 'x-add-relation'
  | 'x-set-origin-acquired-via'
  | 'x-set-origin-purchased-from'
  | 'x-set-origin-built-by';

interface ActionOption {
  value: ConnectedEntityActionType;
  label: string;
}

const ACTION_DEFINITIONS: { key: ConnectedEntityActionType; label: string }[] = [
  { key: 'x-add-relation', label: 'Create related component' },
  { key: 'x-set-origin-acquired-via', label: 'Acquired via' },
  { key: 'x-set-origin-purchased-from', label: 'Purchased from' },
  { key: 'x-set-origin-built-by', label: 'Built by' },
];

function deriveAvailableActions(links: HATEOASLinks): ActionOption[] {
  return ACTION_DEFINITIONS.filter(({ key }) => links[key] != null).map(({ key, label }) => ({
    value: key,
    label,
  }));
}

const createConnectedEntitySchema = z.object({
  name: componentNameSchema,
  description: componentDescriptionSchema,
  actionType: z.string().min(1, 'Action type is required'),
  relationType: relationTypeSchema.optional(),
});

export type CreateConnectedEntityFormData = z.infer<typeof createConnectedEntitySchema>;

export interface CreateConnectedEntitySubmitData {
  name: string;
  description: string;
  actionType: ConnectedEntityActionType;
  relationType?: string;
  actionLink: HATEOASLink;
}

export interface CreateConnectedEntityDialogProps {
  isOpen: boolean;
  sourceNodeId: string;
  sourceNodeType: string;
  handlePosition: string;
  links: HATEOASLinks;
  onSubmit: (data: CreateConnectedEntitySubmitData) => void;
  onClose: () => void;
}

export const CreateConnectedEntityDialog: React.FC<CreateConnectedEntityDialogProps> = ({
  isOpen,
  links,
  onSubmit,
  onClose,
}) => {
  const availableActions = useMemo(() => deriveAvailableActions(links), [links]);
  const hasActions = availableActions.length > 0;
  const singleAction = availableActions.length === 1;

  const {
    register,
    handleSubmit,
    control,
    reset,
    watch,
    formState: { errors, isValid },
  } = useForm<CreateConnectedEntityFormData>({
    resolver: zodResolver(createConnectedEntitySchema),
    defaultValues: {
      name: '',
      description: '',
      actionType: singleAction ? availableActions[0].value : '',
      relationType: 'Triggers',
    },
    mode: 'onChange',
  });

  useEffect(() => {
    if (isOpen) {
      reset({
        name: '',
        description: '',
        actionType: singleAction ? availableActions[0].value : '',
        relationType: 'Triggers',
      });
    }
  }, [isOpen, reset, singleAction, availableActions]);

  const selectedActionType = watch('actionType') as ConnectedEntityActionType;
  const showRelationType = selectedActionType === 'x-add-relation';

  const handleFormSubmit = (data: CreateConnectedEntityFormData) => {
    const actionLink = links[data.actionType];
    if (!actionLink) return;

    onSubmit({
      name: data.name,
      description: data.description || '',
      actionType: data.actionType as ConnectedEntityActionType,
      relationType: data.actionType === 'x-add-relation' ? data.relationType : undefined,
      actionLink,
    });
  };

  if (!hasActions) {
    return (
      <Modal opened={isOpen} onClose={onClose} title="Create Connected Entity" centered data-testid="create-connected-entity-dialog">
        <Stack gap="md">
          <Text data-testid="no-actions-message">No actions available for this entity.</Text>
          <Group justify="flex-end">
            <Button variant="default" onClick={onClose} data-testid="create-connected-entity-close">
              Close
            </Button>
          </Group>
        </Stack>
      </Modal>
    );
  }

  return (
    <Modal opened={isOpen} onClose={onClose} title="Create Connected Entity" centered data-testid="create-connected-entity-dialog">
      <form onSubmit={handleSubmit(handleFormSubmit)}>
        <Stack gap="md">
          {!singleAction && (
            <Controller
              name="actionType"
              control={control}
              render={({ field }) => (
                <Select
                  label="Action"
                  placeholder="Select action"
                  data={availableActions}
                  required
                  withAsterisk
                  error={errors.actionType?.message}
                  data-testid="connected-entity-action-select"
                  {...field}
                />
              )}
            />
          )}

          {showRelationType && (
            <Controller
              name="relationType"
              control={control}
              render={({ field }) => (
                <Select
                  label="Relation Type"
                  data={RELATION_TYPE_OPTIONS}
                  required
                  withAsterisk
                  data-testid="connected-entity-relation-type-select"
                  {...field}
                />
              )}
            />
          )}

          <TextInput
            label="Name"
            placeholder="Enter name"
            required
            withAsterisk
            {...register('name')}
            error={errors.name?.message}
            data-testid="connected-entity-name-input"
          />

          <Textarea
            label="Description"
            placeholder="Enter description (optional)"
            {...register('description')}
            rows={3}
            error={errors.description?.message}
            data-testid="connected-entity-description-input"
          />

          <Group justify="flex-end" gap="sm">
            <Button variant="default" onClick={onClose} data-testid="create-connected-entity-cancel">
              Cancel
            </Button>
            <Button type="submit" disabled={!isValid} data-testid="create-connected-entity-submit">
              Create
            </Button>
          </Group>
        </Stack>
      </form>
    </Modal>
  );
};
