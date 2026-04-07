import { auditQueryKeys } from '../audit/queryKeys';
import { relationsQueryKeys } from './queryKeys';

export const relationsMutationEffects = {
  create: () => [relationsQueryKeys.lists()],

  update: (relationId: string) => [
    relationsQueryKeys.lists(),
    relationsQueryKeys.detail(relationId),
    auditQueryKeys.history(relationId),
  ],

  delete: (relationId: string) => [relationsQueryKeys.lists(), relationsQueryKeys.detail(relationId)],
};
