import { zodResolver } from '@hookform/resolvers/zod';
import { Button, Group, Modal, Stack, Textarea, TextInput } from '@mantine/core';
import { useForm } from 'react-hook-form';
import { type RenameCustomFieldFormData, renameCustomFieldSchema } from '../../../lib/schemas/onePagerConfiguration';
import type { CustomField } from '../types';

interface RenameFieldDialogProps {
  field: CustomField;
  isSaving: boolean;
  onSave: (field: CustomField, data: RenameCustomFieldFormData) => void;
  onClose: () => void;
}

export function RenameFieldDialog({ field, isSaving, onSave, onClose }: RenameFieldDialogProps) {
  const {
    register,
    handleSubmit,
    formState: { errors, isValid },
  } = useForm<RenameCustomFieldFormData>({
    resolver: zodResolver(renameCustomFieldSchema),
    defaultValues: { name: field.name, helpText: field.helpText },
    mode: 'onChange',
  });

  const submit = handleSubmit((data) => onSave(field, data));

  return (
    <Modal opened onClose={onClose} title="Rename field" centered data-testid="one-pager-rename-dialog">
      <form onSubmit={submit}>
        <Stack gap="md">
          <TextInput
            label="Name"
            {...register('name')}
            required
            withAsterisk
            autoFocus
            error={errors.name?.message}
            data-testid="one-pager-rename-name-input"
          />
          <Textarea
            label="Help text"
            {...register('helpText')}
            rows={3}
            error={errors.helpText?.message}
            data-testid="one-pager-rename-help-input"
          />
          <Group justify="flex-end" gap="sm">
            <Button variant="default" onClick={onClose} disabled={isSaving}>
              Cancel
            </Button>
            <Button type="submit" loading={isSaving} disabled={!isValid} data-testid="one-pager-rename-submit">
              Save
            </Button>
          </Group>
        </Stack>
      </form>
    </Modal>
  );
}
