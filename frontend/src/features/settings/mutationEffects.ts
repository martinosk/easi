import { chatQueryKeys } from '../chat/queryKeys';
import { assistantConfigQueryKeys } from './queryKeys';

export const assistantConfigMutationEffects = {
  update: () => [assistantConfigQueryKeys.config(), chatQueryKeys.status()],
};
