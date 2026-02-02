import { describe, it, expect } from 'vitest';
import { computeTreeBulkActions } from './useTreeMultiSelectMenu';
import type { TreeSelectedItem } from './useTreeMultiSelect';
import type { HATEOASLinks } from '../../../api/types';

function makeItem(id: string, links?: HATEOASLinks): TreeSelectedItem {
  return {
    id,
    name: `Item ${id}`,
    type: 'component',
    links,
  };
}

const deleteLinks: HATEOASLinks = {
  self: { href: '/test', method: 'GET' },
  delete: { href: `/test/${Math.random()}`, method: 'DELETE' },
};

const noDeleteLinks: HATEOASLinks = {
  self: { href: '/test', method: 'GET' },
};

describe('computeTreeBulkActions', () => {
  it('returns empty for fewer than 2 items', () => {
    expect(computeTreeBulkActions([makeItem('1', deleteLinks)])).toEqual([]);
    expect(computeTreeBulkActions([])).toEqual([]);
  });

  it('all items have delete link - shows Delete from Model', () => {
    const items = [
      makeItem('1', deleteLinks),
      makeItem('2', deleteLinks),
      makeItem('3', deleteLinks),
    ];
    const actions = computeTreeBulkActions(items);
    expect(actions).toHaveLength(1);
    expect(actions[0]).toEqual({
      type: 'deleteFromModel',
      label: 'Delete from Model (3 items)',
      isDanger: true,
    });
  });

  it('mixed permissions - some without delete - no actions', () => {
    const items = [
      makeItem('1', deleteLinks),
      makeItem('2', noDeleteLinks),
    ];
    const actions = computeTreeBulkActions(items);
    expect(actions).toHaveLength(0);
  });

  it('items with undefined links - no actions', () => {
    const items = [
      makeItem('1', deleteLinks),
      makeItem('2', undefined),
    ];
    const actions = computeTreeBulkActions(items);
    expect(actions).toHaveLength(0);
  });

  it('exactly 2 items both with delete - shows action', () => {
    const items = [
      makeItem('1', deleteLinks),
      makeItem('2', deleteLinks),
    ];
    const actions = computeTreeBulkActions(items);
    expect(actions).toHaveLength(1);
    expect(actions[0].label).toBe('Delete from Model (2 items)');
  });
});
