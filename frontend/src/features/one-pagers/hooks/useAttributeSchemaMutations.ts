import { useMutation, useQueryClient } from '@tanstack/react-query';
import toast from 'react-hot-toast';
import { subjectAttributeSchemaApi } from '../../../api/metadata';
import type {
  AddAttributeOptionRequest,
  DefineSubjectAttributeRequest,
  RenameSubjectAttributeRequest,
  SchemaVersionRequest,
  SetAttributeBoundsRequest,
  SubjectAttribute,
  SubjectAttributeOption,
  SubjectAttributeSchema,
} from '../../../api/types';
import { invalidateFor } from '../../../lib/invalidateFor';
import { onePagersMutationEffects } from '../mutationEffects';
import type { OnePagerSubjectType } from '../types';
import { isOnePagerConflict } from './useOnePagerMutations';

const CONFLICT_MESSAGE = 'Attribute schema was changed by someone else. Refreshed with the latest version.';

function getErrorMessage(err: unknown, fallback: string): string {
  if (err instanceof Error) return err.message;
  return fallback;
}

interface SchemaMutationConfig<TVars> {
  call: (vars: TVars) => Promise<SubjectAttributeSchema>;
  subjectType: OnePagerSubjectType;
  successMessage: string;
  failureMessage: string;
}

function useSchemaMutation<TVars>({ call, subjectType, successMessage, failureMessage }: SchemaMutationConfig<TVars>) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: call,
    onSuccess: () => {
      invalidateFor(queryClient, onePagersMutationEffects.attributeSchema(subjectType));
      toast.success(successMessage);
    },
    onError: (err) => {
      invalidateFor(queryClient, onePagersMutationEffects.attributeSchema(subjectType));
      toast.error(isOnePagerConflict(err) ? CONFLICT_MESSAGE : getErrorMessage(err, failureMessage));
    },
  });
}

export function useDefineAttribute(subjectType: OnePagerSubjectType) {
  return useSchemaMutation<{ schema: SubjectAttributeSchema; request: DefineSubjectAttributeRequest }>({
    call: ({ schema, request }) => subjectAttributeSchemaApi.defineAttribute(schema, request),
    subjectType,
    successMessage: 'Custom field defined',
    failureMessage: 'Failed to define custom field',
  });
}

export function useRenameAttribute(subjectType: OnePagerSubjectType) {
  return useSchemaMutation<{ attribute: SubjectAttribute; request: RenameSubjectAttributeRequest }>({
    call: ({ attribute, request }) => subjectAttributeSchemaApi.renameAttribute(attribute, request),
    subjectType,
    successMessage: 'Field renamed',
    failureMessage: 'Failed to rename field',
  });
}

export function useRetireAttribute(subjectType: OnePagerSubjectType) {
  return useSchemaMutation<{ attribute: SubjectAttribute; request: SchemaVersionRequest }>({
    call: ({ attribute, request }) => subjectAttributeSchemaApi.retireAttribute(attribute, request),
    subjectType,
    successMessage: 'Field retired',
    failureMessage: 'Failed to retire field',
  });
}

export function useReactivateAttribute(subjectType: OnePagerSubjectType) {
  return useSchemaMutation<{ attribute: SubjectAttribute; request: SchemaVersionRequest }>({
    call: ({ attribute, request }) => subjectAttributeSchemaApi.reactivateAttribute(attribute, request),
    subjectType,
    successMessage: 'Field reactivated',
    failureMessage: 'Failed to reactivate field',
  });
}

export function useAddAttributeOption(subjectType: OnePagerSubjectType) {
  return useSchemaMutation<{ attribute: SubjectAttribute; request: AddAttributeOptionRequest }>({
    call: ({ attribute, request }) => subjectAttributeSchemaApi.addOption(attribute, request),
    subjectType,
    successMessage: 'Option added',
    failureMessage: 'Failed to add option',
  });
}

export function useRetireAttributeOption(subjectType: OnePagerSubjectType) {
  return useSchemaMutation<{ option: SubjectAttributeOption; request: SchemaVersionRequest }>({
    call: ({ option, request }) => subjectAttributeSchemaApi.retireOption(option, request),
    subjectType,
    successMessage: 'Option retired',
    failureMessage: 'Failed to retire option',
  });
}

export function useSetAttributeBounds(subjectType: OnePagerSubjectType) {
  return useSchemaMutation<{ attribute: SubjectAttribute; request: SetAttributeBoundsRequest }>({
    call: ({ attribute, request }) => subjectAttributeSchemaApi.setBounds(attribute, request),
    subjectType,
    successMessage: 'Bounds updated',
    failureMessage: 'Failed to update bounds',
  });
}
