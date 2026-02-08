import { layoutsQueryKeys } from './queryKeys';

export const layoutsMutationEffects = {
  upsert: (contextType: string, contextRef: string) => [
    layoutsQueryKeys.detail(contextType, contextRef),
  ],

  delete: (contextType: string, contextRef: string) => [
    layoutsQueryKeys.detail(contextType, contextRef),
  ],

  updatePreferences: (contextType: string, contextRef: string) => [
    layoutsQueryKeys.detail(contextType, contextRef),
  ],

  updateElement: (contextType: string, contextRef: string) => [
    layoutsQueryKeys.detail(contextType, contextRef),
  ],
};
