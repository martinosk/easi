import { describe, expect, it } from 'vitest';
import { moveMilestone } from './milestoneOrder';

describe('moveMilestone', () => {
  const ids = ['a', 'b', 'c', 'd'];

  it('moves an item later in the list', () => {
    expect(moveMilestone(ids, 0, 2)).toEqual(['b', 'c', 'a', 'd']);
  });

  it('moves an item earlier in the list', () => {
    expect(moveMilestone(ids, 3, 1)).toEqual(['a', 'd', 'b', 'c']);
  });

  it('returns null for a no-op move so no request is sent', () => {
    expect(moveMilestone(ids, 2, 2)).toBeNull();
  });

  it('returns null when the target is outside the list', () => {
    expect(moveMilestone(ids, 0, -1)).toBeNull();
    expect(moveMilestone(ids, 3, 4)).toBeNull();
  });

  it('does not mutate the input', () => {
    moveMilestone(ids, 0, 3);
    expect(ids).toEqual(['a', 'b', 'c', 'd']);
  });
});
