import { describe, it, expect, vi } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useTreeMultiSelect } from './useTreeMultiSelect';
import type { TreeSelectedItem, MultiDragPayload } from './useTreeMultiSelect';
import type { HATEOASLinks } from '../../../api/types';

const deleteLinks: HATEOASLinks = {
  self: { href: '/test', method: 'GET' },
  delete: { href: '/test', method: 'DELETE' },
};

function makeItem(id: string, type: TreeSelectedItem['type'] = 'component', name?: string): TreeSelectedItem {
  return { id, name: name ?? `Item ${id}`, type, links: deleteLinks };
}

function makeVisibleItems(ids: string[], type: TreeSelectedItem['type'] = 'component'): TreeSelectedItem[] {
  return ids.map((id) => makeItem(id, type));
}

function makeMouseEvent(overrides: Partial<React.MouseEvent> = {}): React.MouseEvent {
  return {
    ctrlKey: false,
    metaKey: false,
    shiftKey: false,
    preventDefault: () => {},
    stopPropagation: () => {},
    ...overrides,
  } as unknown as React.MouseEvent;
}

function expectSelected(
  hook: { current: ReturnType<typeof useTreeMultiSelect> },
  selectedIds: string[],
  notSelectedIds: string[] = []
) {
  for (const id of selectedIds) {
    expect(hook.current.isMultiSelected(id)).toBe(true);
  }
  for (const id of notSelectedIds) {
    expect(hook.current.isMultiSelected(id)).toBe(false);
  }
}

function setup() {
  const { result } = renderHook(() => useTreeMultiSelect());

  const ctrlClick = (item: TreeSelectedItem, sectionId: string, visible: TreeSelectedItem[]) => {
    act(() => { result.current.handleItemClick(item, sectionId, visible, makeMouseEvent({ ctrlKey: true })); });
  };

  const shiftClick = (item: TreeSelectedItem, sectionId: string, visible: TreeSelectedItem[]) => {
    act(() => { result.current.handleItemClick(item, sectionId, visible, makeMouseEvent({ shiftKey: true })); });
  };

  const plainClick = (item: TreeSelectedItem, sectionId: string, visible: TreeSelectedItem[]) => {
    let outcome: 'multi' | 'single' = 'multi';
    act(() => { outcome = result.current.handleItemClick(item, sectionId, visible, makeMouseEvent()); });
    return outcome;
  };

  const ctrlClickTwo = (visible: TreeSelectedItem[], sectionId = 'apps') => {
    ctrlClick(visible[0], sectionId, visible);
    ctrlClick(visible[1], sectionId, visible);
  };

  return { result, ctrlClick, shiftClick, plainClick, ctrlClickTwo };
}

function makeDragEvent(): React.DragEvent {
  return {
    dataTransfer: { setData: vi.fn(), setDragImage: vi.fn(), effectAllowed: '' },
    preventDefault: vi.fn(),
    stopPropagation: vi.fn(),
  } as unknown as React.DragEvent;
}

describe('useTreeMultiSelect', () => {
  it('plain click clears multi-selection and returns single', () => {
    const { result, ctrlClickTwo, plainClick } = setup();
    const visible = makeVisibleItems(['a', 'b', 'c']);

    ctrlClickTwo(visible);
    expect(result.current.selectionCount).toBe(2);

    const outcome = plainClick(makeItem('c'), 'apps', visible);

    expect(outcome).toBe('single');
    expect(result.current.selectionCount).toBe(0);
  });

  it('Ctrl+click toggles item and returns multi', () => {
    const { result } = setup();
    const visible = makeVisibleItems(['a', 'b']);

    let outcome: 'multi' | 'single' = 'single';
    act(() => { outcome = result.current.handleItemClick(makeItem('a'), 'apps', visible, makeMouseEvent({ ctrlKey: true })); });

    expect(outcome).toBe('multi');
    expect(result.current.selectionCount).toBe(1);
    expectSelected(result, ['a']);
  });

  it('Ctrl+click deselects already-selected item', () => {
    const { result, ctrlClickTwo, ctrlClick } = setup();
    const visible = makeVisibleItems(['a', 'b']);

    ctrlClickTwo(visible);
    expect(result.current.selectionCount).toBe(2);

    ctrlClick(makeItem('a'), 'apps', visible);

    expect(result.current.selectionCount).toBe(1);
    expectSelected(result, ['b'], ['a']);
  });

  it('Shift+click selects range within section', () => {
    const { result, ctrlClick, shiftClick } = setup();
    const visible = makeVisibleItems(['a', 'b', 'c', 'd']);

    ctrlClick(makeItem('a'), 'apps', visible);
    shiftClick(makeItem('c'), 'apps', visible);

    expectSelected(result, ['a', 'b', 'c'], ['d']);
  });

  it('Shift+click does not cross sections - preserves other section', () => {
    const { result, ctrlClick, shiftClick } = setup();

    ctrlClick(makeItem('app-a', 'component'), 'apps', makeVisibleItems(['app-a', 'app-b']));
    shiftClick(makeItem('cap-x', 'capability'), 'capabilities', makeVisibleItems(['cap-x', 'cap-y', 'cap-z'], 'capability'));

    expectSelected(result, ['app-a', 'cap-x']);
  });

  it('Shift+click with no anchor selects from first to clicked', () => {
    const { result, shiftClick } = setup();
    const visible = makeVisibleItems(['a', 'b', 'c', 'd']);

    shiftClick(makeItem('c'), 'apps', visible);

    expectSelected(result, ['a', 'b', 'c'], ['d']);
  });

  it('cross-section Ctrl+click works', () => {
    const { result, ctrlClick } = setup();

    ctrlClick(makeItem('app-1', 'component'), 'apps', makeVisibleItems(['app-1']));
    ctrlClick(makeItem('cap-1', 'capability'), 'capabilities', makeVisibleItems(['cap-1'], 'capability'));
    ctrlClick(makeItem('v-1', 'vendor'), 'vendors', makeVisibleItems(['v-1'], 'vendor'));

    expect(result.current.selectionCount).toBe(3);
    expectSelected(result, ['app-1', 'cap-1', 'v-1']);
  });

  it('clearMultiSelection clears all', () => {
    const { result, ctrlClickTwo } = setup();

    ctrlClickTwo(makeVisibleItems(['a', 'b']));
    expect(result.current.selectionCount).toBe(2);

    act(() => { result.current.clearMultiSelection(); });

    expect(result.current.selectionCount).toBe(0);
  });

  it('getSelectedItems returns all selected items', () => {
    const { result, ctrlClick } = setup();

    ctrlClick(makeItem('a', 'component', 'App A'), 'apps', makeVisibleItems(['a']));
    ctrlClick(makeItem('b', 'capability', 'Cap B'), 'caps', makeVisibleItems(['b'], 'capability'));

    const items = result.current.getSelectedItems();
    expect(items).toHaveLength(2);
    expect(items.map((i) => i.id).sort()).toEqual(['a', 'b']);
  });

  it('Ctrl+click after single click includes the previously single-selected item', () => {
    const { result, plainClick, ctrlClick } = setup();
    const visible = makeVisibleItems(['a', 'b', 'c']);

    plainClick(makeItem('a'), 'apps', visible);
    expect(result.current.selectionCount).toBe(0);

    ctrlClick(makeItem('b'), 'apps', visible);

    expect(result.current.selectionCount).toBe(2);
    expectSelected(result, ['a', 'b']);
  });

  it('Cmd+click (metaKey) works the same as Ctrl+click', () => {
    const { result } = setup();

    let outcome: 'multi' | 'single' = 'single';
    act(() => { outcome = result.current.handleItemClick(makeItem('a'), 'apps', makeVisibleItems(['a']), makeMouseEvent({ metaKey: true })); });

    expect(outcome).toBe('multi');
    expectSelected(result, ['a']);
  });

  describe('buildMultiDragPayload', () => {
    it('returns JSON with all selected items grouped by type', () => {
      const { result, ctrlClick } = setup();

      ctrlClick(makeItem('app-1', 'component', 'App A'), 'apps', makeVisibleItems(['app-1']));
      ctrlClick(makeItem('cap-1', 'capability', 'Cap B'), 'caps', makeVisibleItems(['cap-1'], 'capability'));

      const payload: MultiDragPayload = JSON.parse(result.current.buildMultiDragPayload());
      expect(payload.items).toHaveLength(2);

      const ids = payload.items.map((i) => i.id).sort();
      expect(ids).toEqual(['app-1', 'cap-1']);

      const app = payload.items.find((i) => i.id === 'app-1')!;
      expect(app.type).toBe('component');
      expect(app.name).toBe('App A');

      const cap = payload.items.find((i) => i.id === 'cap-1')!;
      expect(cap.type).toBe('capability');
      expect(cap.name).toBe('Cap B');
    });

    it('returns empty items array when nothing is selected', () => {
      const { result } = setup();

      const payload: MultiDragPayload = JSON.parse(result.current.buildMultiDragPayload());
      expect(payload.items).toEqual([]);
    });
  });

  describe('handleDragStart', () => {
    it('returns true and sets multiDragItems when item is in multi-selection', () => {
      const { result, ctrlClickTwo } = setup();

      ctrlClickTwo(makeVisibleItems(['a', 'b']));

      const event = makeDragEvent();
      let handled = false;
      act(() => { handled = result.current.handleDragStart(event, 'a'); });

      expect(handled).toBe(true);
      expect(event.dataTransfer.setData).toHaveBeenCalledWith('multiDragItems', expect.any(String));

      const payload: MultiDragPayload = JSON.parse(
        (event.dataTransfer.setData as ReturnType<typeof vi.fn>).mock.calls[0][1]
      );
      expect(payload.items).toHaveLength(2);
    });

    it('returns false when fewer than 2 items are selected', () => {
      const { result, ctrlClick } = setup();

      ctrlClick(makeItem('a'), 'apps', makeVisibleItems(['a']));

      const event = makeDragEvent();
      let handled = false;
      act(() => { handled = result.current.handleDragStart(event, 'a'); });

      expect(handled).toBe(false);
      expect(event.dataTransfer.setData).not.toHaveBeenCalled();
    });

    it('returns false when dragged item is not in the selection', () => {
      const { result, ctrlClickTwo } = setup();

      ctrlClickTwo(makeVisibleItems(['a', 'b', 'c']));

      const event = makeDragEvent();
      let handled = false;
      act(() => { handled = result.current.handleDragStart(event, 'c'); });

      expect(handled).toBe(false);
      expect(event.dataTransfer.setData).not.toHaveBeenCalled();
    });
  });
});
