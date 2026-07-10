import { useMutation, useQueryClient } from '@tanstack/react-query';
import toast from 'react-hot-toast';
import { invalidateFor } from '../../../lib/invalidateFor';
import { onePagersApi } from '../api/onePagersApi';
import { onePagersMutationEffects } from '../mutationEffects';
import type { FieldValue, OnePagerFacts, OnePagerSubjectType, ValueEnvelope } from '../types';

export interface FieldValueRecord {
  fieldId: string;
  value: ValueEnvelope;
}

export interface SaveOnePagerFactsInput {
  facts: OnePagerFacts;
  records: FieldValueRecord[];
  clears: FieldValue[];
}

async function saveFacts({ facts, records, clears }: SaveOnePagerFactsInput): Promise<void> {
  for (const record of records) {
    await onePagersApi.recordFieldValue(facts, record.fieldId, { value: record.value });
  }
  for (const fieldValue of clears) {
    await onePagersApi.clearFieldValue(fieldValue);
  }
}

function errorMessage(err: unknown): string {
  return err instanceof Error ? err.message : 'Failed to update one-pager';
}

export function useSaveOnePagerFacts(subjectType: OnePagerSubjectType, subjectId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: saveFacts,
    onSuccess: () => {
      invalidateFor(queryClient, onePagersMutationEffects.facts(subjectType, subjectId));
      toast.success('One-Pager updated');
    },
    onError: (err) => {
      invalidateFor(queryClient, onePagersMutationEffects.facts(subjectType, subjectId));
      toast.error(errorMessage(err));
    },
  });
}
