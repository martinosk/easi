import { describe, expect, it } from 'vitest';
import { compositionSummariesQueryKeys } from '../enterprise-architecture/queryKeys';
import { directionMutationEffects } from './mutationEffects';

const DIRECTION_MUTATIONS = [
  'capture',
  'addSource',
  'removeSource',
  'update',
  'propose',
  'agree',
  'reject',
  'revert',
] as const;

describe('directionMutationEffects', () => {
  it.each(DIRECTION_MUTATIONS)('invalidates the composition summaries after %s', (mutation) => {
    const effects = directionMutationEffects[mutation]('ec-1');

    expect(effects).toContainEqual(compositionSummariesQueryKeys.lists());
  });
});
