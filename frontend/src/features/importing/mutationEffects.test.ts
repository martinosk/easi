import { describe, expect, it } from 'vitest';
import { onePagersQueryKeys } from '../one-pagers/queryKeys';
import { importsMutationEffects } from './mutationEffects';

describe('importsMutationEffects', () => {
  it('invalidates one-pager completeness for the subject types an import creates', () => {
    const effects = importsMutationEffects.completed();

    expect(effects).toContainEqual(onePagersQueryKeys.completenessForSubjectType('application'));
    expect(effects).toContainEqual(onePagersQueryKeys.completenessForSubjectType('capability'));
  });
});
