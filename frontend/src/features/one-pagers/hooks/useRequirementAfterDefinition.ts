import { useEffect, useState } from 'react';
import { hasLink } from '../../../utils/hateoas';
import type { OnePagerConfiguration, OnePagerSubjectType } from '../types';
import { useChangeFieldRequirement } from './useOnePagerMutations';

export function useRequirementAfterDefinition(
  subjectType: OnePagerSubjectType,
  configuration: OnePagerConfiguration | undefined,
) {
  const requireField = useChangeFieldRequirement(subjectType);
  const [awaitedFieldId, setAwaitedFieldId] = useState<string | null>(null);

  useEffect(() => {
    if (!awaitedFieldId || !configuration) return;
    const field = configuration.customFields.find((candidate) => candidate.id === awaitedFieldId);
    if (!field || !hasLink(field, 'x-set-requirement')) return;
    setAwaitedFieldId(null);
    requireField.mutate({ field, request: { required: true, version: configuration.version } });
  }, [configuration, awaitedFieldId, requireField]);

  return {
    awaitRequirementFor: (fieldId: string | undefined) => setAwaitedFieldId(fieldId ?? null),
    isPending: requireField.isPending || awaitedFieldId !== null,
  };
}
