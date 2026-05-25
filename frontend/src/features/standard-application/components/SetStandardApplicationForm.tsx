import { Alert, Button, Group, Select, Stack, Textarea } from '@mantine/core';
import { zodResolver } from '@hookform/resolvers/zod';
import { useMemo } from 'react';
import { Controller, useForm } from 'react-hook-form';
import type { EnterpriseCapabilityId } from '../../../api/types';
import {
  setStandardApplicationSchema,
  type SetStandardApplicationFormValues,
} from '../../../lib/schemas';
import { useComponents } from '../../components/hooks/useComponents';
import { useSetStandardApplication } from '../hooks/useStandardApplication';

interface SetStandardApplicationFormProps {
  enterpriseCapabilityId: EnterpriseCapabilityId;
  initialApplicationId?: string;
  initialNarrative?: string;
  onSubmitted: () => void;
  onCancel: () => void;
}

export function SetStandardApplicationForm({
  enterpriseCapabilityId,
  initialApplicationId,
  initialNarrative,
  onSubmitted,
  onCancel,
}: SetStandardApplicationFormProps) {
  const { data: components, isLoading: loadingComponents, error: componentsError } = useComponents();
  const setMutation = useSetStandardApplication();

  const {
    control,
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<SetStandardApplicationFormValues>({
    resolver: zodResolver(setStandardApplicationSchema),
    defaultValues: {
      applicationId: initialApplicationId ?? '',
      narrative: initialNarrative ?? '',
    },
    mode: 'onSubmit',
    reValidateMode: 'onChange',
  });

  const componentOptions = useMemo(
    () =>
      (components ?? []).map((c) => ({
        value: String(c.id),
        label: c.name,
      })),
    [components],
  );

  const onSubmit = handleSubmit((values) => {
    setMutation.mutate(
      {
        enterpriseCapabilityId,
        request: { applicationId: values.applicationId, narrative: values.narrative },
      },
      { onSuccess: () => onSubmitted() },
    );
  });

  return (
    <form onSubmit={onSubmit}>
      <Stack gap="md">
        {componentsError && <Alert color="red">Failed to load applications.</Alert>}
        <Controller
          control={control}
          name="applicationId"
          render={({ field }) => (
            <Select
              label="Application"
              placeholder="Pick the standard application"
              data={componentOptions}
              searchable
              clearable
              disabled={loadingComponents}
              value={field.value}
              onChange={(value) => field.onChange(value ?? '')}
              error={errors.applicationId?.message}
              data-testid="standard-application-picker"
            />
          )}
        />
        <Textarea
          label="Narrative"
          placeholder="Why this application, and what it covers (e.g. 'covers the operational and reporting layers; excludes legacy COBOL flows')."
          autosize
          minRows={3}
          maxLength={1000}
          {...register('narrative')}
          error={errors.narrative?.message}
          data-testid="standard-application-narrative"
        />
        <Group justify="flex-end">
          <Button variant="default" onClick={onCancel} type="button">
            Cancel
          </Button>
          <Button type="submit" disabled={isSubmitting} data-testid="standard-application-submit">
            Save standard
          </Button>
        </Group>
      </Stack>
    </form>
  );
}
