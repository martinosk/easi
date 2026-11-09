import { zodResolver } from '@hookform/resolvers/zod';
import { Alert, Button, Group, Modal, Select, Stack } from '@mantine/core';
import { useMemo, useState } from 'react';
import { Controller, useForm } from 'react-hook-form';
import type { Capability, CapabilityRealization } from '../../../api/types';
import { toComponentId } from '../../../api/types';
import { type AddRealizationFormData, addRealizationSchema } from '../../../lib/schemas';
import { getPostableRelated } from '../../../utils/xRelated';
import { CreateComponentDialog } from '../../components/components/CreateComponentDialog';
import { useComponents } from '../../components/hooks/useComponents';
import { useAddRealization } from '../hooks/useCapabilities';

const LEVEL_OPTIONS = [
  { value: 'Full', label: 'Full' },
  { value: 'Partial', label: 'Partial' },
  { value: 'Planned', label: 'Planned' },
];

const DEFAULT_VALUES: AddRealizationFormData = { componentId: '', realizationLevel: 'Full' };

interface AddRealizationDialogProps {
  capability: Capability;
  realizations: CapabilityRealization[];
  onClose: () => void;
}

function useApplicationOptions(realizations: CapabilityRealization[]) {
  const { data: components = [] } = useComponents();
  return useMemo(() => {
    const directlyRealising = new Set(realizations.filter((r) => r.origin === 'Direct').map((r) => r.componentId));
    return components
      .filter((component) => !directlyRealising.has(component.id))
      .map((component) => ({ value: component.id, label: component.name }));
  }, [components, realizations]);
}

function canCreateRealizingApplication(capability: Capability): boolean {
  return getPostableRelated(capability).some((entry) => entry.relationType === 'capability-realization');
}

export function AddRealizationDialog({ capability, realizations, onClose }: AddRealizationDialogProps) {
  const [error, setError] = useState<string | null>(null);
  const [creatingApplication, setCreatingApplication] = useState(false);
  const addRealization = useAddRealization();
  const applicationOptions = useApplicationOptions(realizations);
  const {
    control,
    getValues,
    handleSubmit,
    formState: { isValid },
  } = useForm<AddRealizationFormData>({
    resolver: zodResolver(addRealizationSchema),
    defaultValues: DEFAULT_VALUES,
    mode: 'onChange',
  });

  const onSubmit = async (data: AddRealizationFormData) => {
    setError(null);
    try {
      await addRealization.mutateAsync({
        capability,
        request: { componentId: toComponentId(data.componentId), realizationLevel: data.realizationLevel },
      });
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to add application');
    }
  };

  const linkCreatedApplication = async (component: { id: string }) => {
    await addRealization.mutateAsync({
      capability,
      request: { componentId: toComponentId(component.id), realizationLevel: getValues('realizationLevel') },
    });
    onClose();
  };

  if (creatingApplication) {
    return (
      <CreateComponentDialog isOpen onClose={() => setCreatingApplication(false)} onCreated={linkCreatedApplication} />
    );
  }

  return (
    <Modal opened onClose={onClose} title="Add realising application" centered>
      <form onSubmit={handleSubmit(onSubmit)}>
        <Stack gap="md">
          <Controller
            name="componentId"
            control={control}
            render={({ field }) => (
              <Select
                label="Application"
                placeholder="Search applications"
                data={applicationOptions}
                value={field.value || null}
                onChange={(value) => field.onChange(value ?? '')}
                searchable
                required
                nothingFoundMessage="No applications available"
                data-testid="add-realization-application"
              />
            )}
          />
          <Controller
            name="realizationLevel"
            control={control}
            render={({ field }) => (
              <Select
                label="Realization level"
                data={LEVEL_OPTIONS}
                value={field.value}
                onChange={(value) => value && field.onChange(value)}
                allowDeselect={false}
                required
                data-testid="add-realization-level"
              />
            )}
          />
          {canCreateRealizingApplication(capability) && (
            <Group justify="flex-start">
              <Button variant="subtle" size="xs" onClick={() => setCreatingApplication(true)}>
                Create new application
              </Button>
            </Group>
          )}
          {error && <Alert color="red">{error}</Alert>}
          <Group justify="flex-end" gap="sm">
            <Button variant="default" onClick={onClose} disabled={addRealization.isPending}>
              Cancel
            </Button>
            <Button type="submit" disabled={!isValid} loading={addRealization.isPending}>
              Add
            </Button>
          </Group>
        </Stack>
      </form>
    </Modal>
  );
}
