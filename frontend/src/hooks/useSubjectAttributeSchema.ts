import { useQuery } from '@tanstack/react-query';
import { subjectAttributeSchemaApi } from '../api/metadata';
import { metadataQueryKeys } from '../lib/appQueryKeys';

export function useSubjectAttributeSchema(subjectType: string, enabled = true) {
  return useQuery({
    queryKey: metadataQueryKeys.subjectAttributeSchema(subjectType),
    queryFn: () => subjectAttributeSchemaApi.getSchema(subjectType),
    staleTime: Infinity,
    enabled,
  });
}
