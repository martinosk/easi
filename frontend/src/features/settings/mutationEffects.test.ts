import { describe, expect, it } from 'vitest';
import { chatQueryKeys } from '../chat/queryKeys';
import { assistantConfigMutationEffects } from './mutationEffects';
import { assistantConfigQueryKeys } from './queryKeys';

describe('assistantConfigMutationEffects', () => {
  it('invalidates the stored configuration when it is saved', () => {
    expect(assistantConfigMutationEffects.update()).toContainEqual(assistantConfigQueryKeys.config());
  });

  it('invalidates the assistant status so the chat entry point re-evaluates', () => {
    expect(assistantConfigMutationEffects.update()).toContainEqual(chatQueryKeys.status());
  });
});
