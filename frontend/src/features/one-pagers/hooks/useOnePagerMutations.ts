import { useMutation, useQueryClient } from '@tanstack/react-query';
import toast from 'react-hot-toast';
import { ApiError } from '../../../api/types';
import { invalidateFor } from '../../../lib/invalidateFor';
import { onePagersApi } from '../api/onePagersApi';
import { onePagersMutationEffects } from '../mutationEffects';
import type {
  BuiltInField,
  ChangeRequirementRequest,
  CustomField,
  OnePagerConfiguration,
  OnePagerSubjectType,
  ReorderFieldsRequest,
  VersionRequest,
} from '../types';

const CONFLICT_MESSAGE = 'Configuration was changed by someone else. Refreshed with the latest version.';

export function isOnePagerConflict(error: unknown): boolean {
  return error instanceof ApiError && error.statusCode === 409;
}

function getErrorMessage(err: unknown, fallback: string): string {
  if (err instanceof Error) return err.message;
  return fallback;
}

interface OnePagerMutationConfig<TVars> {
  call: (vars: TVars) => Promise<OnePagerConfiguration>;
  subjectType: OnePagerSubjectType;
  successMessage: string;
  failureMessage: string;
}

function useOnePagerMutation<TVars>({
  call,
  subjectType,
  successMessage,
  failureMessage,
}: OnePagerMutationConfig<TVars>) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: call,
    onSuccess: () => {
      invalidateFor(queryClient, onePagersMutationEffects.configuration(subjectType));
      toast.success(successMessage);
    },
    onError: (err) => {
      invalidateFor(queryClient, onePagersMutationEffects.configuration(subjectType));
      toast.error(isOnePagerConflict(err) ? CONFLICT_MESSAGE : getErrorMessage(err, failureMessage));
    },
  });
}

export function useReorderFields(subjectType: OnePagerSubjectType) {
  return useOnePagerMutation<{ configuration: OnePagerConfiguration; request: ReorderFieldsRequest }>({
    call: ({ configuration, request }) => onePagersApi.reorderFields(configuration, request),
    subjectType,
    successMessage: 'Field order updated',
    failureMessage: 'Failed to reorder fields',
  });
}

export function useIncludeBuiltInField(subjectType: OnePagerSubjectType) {
  return useOnePagerMutation<{ field: BuiltInField; request: VersionRequest }>({
    call: ({ field, request }) => onePagersApi.includeBuiltInField(field, request),
    subjectType,
    successMessage: 'Field included',
    failureMessage: 'Failed to include field',
  });
}

export function useExcludeBuiltInField(subjectType: OnePagerSubjectType) {
  return useOnePagerMutation<{ field: BuiltInField; request: VersionRequest }>({
    call: ({ field, request }) => onePagersApi.excludeBuiltInField(field, request),
    subjectType,
    successMessage: 'Field excluded',
    failureMessage: 'Failed to exclude field',
  });
}

export function useChangeFieldRequirement(subjectType: OnePagerSubjectType) {
  return useOnePagerMutation<{ field: CustomField; request: ChangeRequirementRequest }>({
    call: ({ field, request }) => onePagersApi.changeFieldRequirement(field, request),
    subjectType,
    successMessage: 'Requirement updated',
    failureMessage: 'Failed to change requirement',
  });
}

export function useChangeBuiltInFieldRequirement(subjectType: OnePagerSubjectType) {
  return useOnePagerMutation<{ field: BuiltInField; request: ChangeRequirementRequest }>({
    call: ({ field, request }) => onePagersApi.changeBuiltInFieldRequirement(field, request),
    subjectType,
    successMessage: 'Requirement updated',
    failureMessage: 'Failed to change requirement',
  });
}
