import { viewsQueryKeys } from './queryKeys';

export const viewsMutationEffects = {
  create: () => [
    viewsQueryKeys.lists(),
  ],

  delete: (viewId: string) => [
    viewsQueryKeys.lists(),
    viewsQueryKeys.detail(viewId),
  ],

  rename: (viewId: string) => [
    viewsQueryKeys.lists(),
    viewsQueryKeys.detail(viewId),
  ],

  setDefault: () => [
    viewsQueryKeys.lists(),
  ],

  changeVisibility: (viewId: string) => [
    viewsQueryKeys.lists(),
    viewsQueryKeys.detail(viewId),
  ],

  updateDetail: (viewId: string) => [
    viewsQueryKeys.detail(viewId),
  ],
};
