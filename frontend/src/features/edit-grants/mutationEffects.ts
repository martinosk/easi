import { editGrantsQueryKeys } from './queryKeys';

export const editGrantsMutationEffects = {
  create: () => [
    editGrantsQueryKeys.mine(),
    editGrantsQueryKeys.all,
  ],

  revoke: () => [
    editGrantsQueryKeys.mine(),
    editGrantsQueryKeys.all,
  ],
};
