import { describe, expect, it } from 'vitest';
import { stewardshipMutationEffects } from './mutationEffects';
import { stewardshipQueryKeys } from './queryKeys';

describe('stewardshipMutationEffects', () => {
  it.each(['assign', 'release'] as const)('%s invalidates the domain stewardships', (mutation) => {
    expect(stewardshipMutationEffects[mutation]('domain-1')).toContainEqual(stewardshipQueryKeys.domain('domain-1'));
  });
});
