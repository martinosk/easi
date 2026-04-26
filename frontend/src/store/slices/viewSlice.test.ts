import { describe, expect, it } from 'vitest';
import { toViewId } from '../../api/types';
import { createViewSlice, type ViewActions, type ViewState } from './viewSlice';

type Slice = ViewState & ViewActions;

function createStore() {
  let state: Slice;
  const setState = (partial: Partial<Slice> | ((s: Slice) => Partial<Slice>)) => {
    const update = typeof partial === 'function' ? partial(state) : partial;
    state = { ...state, ...update };
  };
  const getState = () => state;
  state = createViewSlice(setState as never, getState as never, {} as never);
  return { getState };
}

const v1 = toViewId('view-1');
const v2 = toViewId('view-2');
const v3 = toViewId('view-3');

describe('viewSlice — openViewIds', () => {
  it('starts with an empty openViewIds list', () => {
    const { getState } = createStore();
    expect(getState().openViewIds).toEqual([]);
  });

  it('openView appends a viewId to the end', () => {
    const { getState } = createStore();
    getState().openView(v1);
    getState().openView(v2);

    expect(getState().openViewIds).toEqual([v1, v2]);
  });

  it('openView is idempotent', () => {
    const { getState } = createStore();
    getState().openView(v1);
    getState().openView(v1);

    expect(getState().openViewIds).toEqual([v1]);
  });

  it('closeView removes the given viewId', () => {
    const { getState } = createStore();
    getState().setOpenViewIds([v1, v2, v3]);

    getState().closeView(v2);

    expect(getState().openViewIds).toEqual([v1, v3]);
  });

  it('closeView is a no-op when the id is not present', () => {
    const { getState } = createStore();
    getState().setOpenViewIds([v1]);

    getState().closeView(v2);

    expect(getState().openViewIds).toEqual([v1]);
  });

  it('setOpenViewIds replaces the entire list', () => {
    const { getState } = createStore();
    getState().openView(v1);

    getState().setOpenViewIds([v2, v3]);

    expect(getState().openViewIds).toEqual([v2, v3]);
  });
});
