import { act, renderHook } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';
import { useDetailGroupLayout } from './useDetailGroupLayout';

const KEY = 'test-detail-layout';
const DEFAULT_ORDER = ['one', 'two', 'three'] as const;

function renderLayout() {
  return renderHook(() => useDetailGroupLayout(KEY, DEFAULT_ORDER));
}

describe('useDetailGroupLayout', () => {
  beforeEach(() => {
    localStorage.clear();
  });

  it('starts with the default order and every group expanded', () => {
    const { result } = renderLayout();

    expect(result.current.order).toEqual(['one', 'two', 'three']);
    expect(result.current.collapsed).toEqual([]);
  });

  it('swaps two groups and remembers the order', () => {
    const { result } = renderLayout();

    act(() => result.current.swap('one', 'two'));
    expect(result.current.order).toEqual(['two', 'one', 'three']);

    act(() => result.current.swap('three', 'one'));
    expect(result.current.order).toEqual(['two', 'three', 'one']);

    const remounted = renderLayout();
    expect(remounted.result.current.order).toEqual(['two', 'three', 'one']);
  });

  it('ignores a swap that names an unknown group', () => {
    const { result } = renderLayout();

    act(() => result.current.swap('one', 'gone'));

    expect(result.current.order).toEqual(['one', 'two', 'three']);
  });

  it('toggles a group collapsed and back and remembers it', () => {
    const { result } = renderLayout();

    act(() => result.current.toggle('two'));
    expect(result.current.collapsed).toEqual(['two']);
    expect(renderLayout().result.current.collapsed).toEqual(['two']);

    act(() => result.current.toggle('two'));
    expect(result.current.collapsed).toEqual([]);
  });

  it('drops unknown ids from a stored arrangement and appends missing ones in default order', () => {
    localStorage.setItem(KEY, JSON.stringify({ order: ['three', 'gone', 'one'], collapsed: ['gone', 'one'] }));

    const { result } = renderLayout();

    expect(result.current.order).toEqual(['three', 'one', 'two']);
    expect(result.current.collapsed).toEqual(['one']);
  });

  it('falls back to the default when the stored value is not an arrangement', () => {
    localStorage.setItem(KEY, 'not json');

    const { result } = renderLayout();

    expect(result.current.order).toEqual(['one', 'two', 'three']);
    expect(result.current.collapsed).toEqual([]);
  });
});
