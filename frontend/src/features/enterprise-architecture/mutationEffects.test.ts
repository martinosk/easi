import { describe, expect, it } from 'vitest';
import { enterpriseCapabilitiesMutationEffects } from './mutationEffects';
import { compositionSummariesQueryKeys } from './queryKeys';

describe('enterpriseCapabilitiesMutationEffects', () => {
  it('invalidates the composition summaries when an enterprise capability is created', () => {
    expect(enterpriseCapabilitiesMutationEffects.create()).toContainEqual(compositionSummariesQueryKeys.lists());
  });

  it('invalidates the composition summaries when an enterprise capability is deleted', () => {
    expect(enterpriseCapabilitiesMutationEffects.delete('ec-1')).toContainEqual(compositionSummariesQueryKeys.lists());
  });
});
