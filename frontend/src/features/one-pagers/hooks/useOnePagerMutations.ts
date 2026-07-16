import { useMutation, useQueryClient } from '@tanstack/react-query';
import toast from 'react-hot-toast';
import { ApiError } from '../../../api/types';
import { invalidateFor } from '../../../lib/invalidateFor';
import { onePagersApi } from '../api/onePagersApi';
import { onePagersMutationEffects } from '../mutationEffects';
import type {
  AddSelectionOptionRequest,
  BuiltInField,
  ChangeRequirementRequest,
  CustomField,
  DefineCustomFieldRequest,
  OnePagerConfiguration,
  OnePagerSubjectType,
  RenameCustomFieldRequest,
  ReorderFieldsRequest,
  SelectionOption,
  SetNumberFieldBoundsRequest,
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

export function useDefineCustomField(subjectType: OnePagerSubjectType) {
  return useOnePagerMutation<{ configuration: OnePagerConfiguration; request: DefineCustomFieldRequest }>({
    call: ({ configuration, request }) => onePagersApi.defineCustomField(configuration, request),
    subjectType,
    successMessage: 'Custom field defined',
    failureMessage: 'Failed to define custom field',
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

export function useRenameCustomField(subjectType: OnePagerSubjectType) {
  return useOnePagerMutation<{ field: CustomField; request: RenameCustomFieldRequest }>({
    call: ({ field, request }) => onePagersApi.renameCustomField(field, request),
    subjectType,
    successMessage: 'Field renamed',
    failureMessage: 'Failed to rename field',
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

export function useRetireCustomField(subjectType: OnePagerSubjectType) {
  return useOnePagerMutation<{ field: CustomField; request: VersionRequest }>({
    call: ({ field, request }) => onePagersApi.retireCustomField(field, request),
    subjectType,
    successMessage: 'Field retired',
    failureMessage: 'Failed to retire field',
  });
}

export function useReactivateCustomField(subjectType: OnePagerSubjectType) {
  return useOnePagerMutation<{ field: CustomField; request: VersionRequest }>({
    call: ({ field, request }) => onePagersApi.reactivateCustomField(field, request),
    subjectType,
    successMessage: 'Field reactivated',
    failureMessage: 'Failed to reactivate field',
  });
}

export function useAddSelectionOption(subjectType: OnePagerSubjectType) {
  return useOnePagerMutation<{ field: CustomField; request: AddSelectionOptionRequest }>({
    call: ({ field, request }) => onePagersApi.addSelectionOption(field, request),
    subjectType,
    successMessage: 'Option added',
    failureMessage: 'Failed to add option',
  });
}

export function useRetireSelectionOption(subjectType: OnePagerSubjectType) {
  return useOnePagerMutation<{ option: SelectionOption; request: VersionRequest }>({
    call: ({ option, request }) => onePagersApi.retireSelectionOption(option, request),
    subjectType,
    successMessage: 'Option retired',
    failureMessage: 'Failed to retire option',
  });
}

export function useSetNumberFieldBounds(subjectType: OnePagerSubjectType) {
  return useOnePagerMutation<{ field: CustomField; request: SetNumberFieldBoundsRequest }>({
    call: ({ field, request }) => onePagersApi.setNumberFieldBounds(field, request),
    subjectType,
    successMessage: 'Bounds updated',
    failureMessage: 'Failed to update bounds',
  });
}
