import { act, renderHook } from '@testing-library/react';
import type { ReactNode } from 'react';
import { MemoryRouter, useLocation } from 'react-router-dom';
import { beforeEach, describe, expect, it } from 'vitest';
import { toBusinessDomainId } from '../../../api/types';
import { buildBusinessDomain } from '../../../test/helpers/entityBuilders';
import { resolveMapDomainId, useMapDepth, useMapDomain, useMapShowApps, useViewMode } from './useMapViewState';

beforeEach(() => {
  localStorage.clear();
});

function routerWrapper(initialEntries: string[] = ['/business-domains']) {
  return ({ children }: { children: ReactNode }) => (
    <MemoryRouter initialEntries={initialEntries}>{children}</MemoryRouter>
  );
}

function renderViewMode(initialEntries?: string[]) {
  return renderHook(() => ({ viewMode: useViewMode(), location: useLocation() }), {
    wrapper: routerWrapper(initialEntries),
  });
}

describe('useViewMode', () => {
  it('defaults to the board view', () => {
    const { result } = renderViewMode();
    expect(result.current.viewMode[0]).toBe('board');
  });

  it('persists the chosen view across hook instances', () => {
    const first = renderViewMode();
    act(() => first.result.current.viewMode[1]('map'));
    expect(first.result.current.viewMode[0]).toBe('map');

    const second = renderViewMode();
    expect(second.result.current.viewMode[0]).toBe('map');
  });

  it('falls back to board when the stored value is invalid', () => {
    localStorage.setItem('business-domains-view', 'bogus');
    const { result } = renderViewMode();
    expect(result.current.viewMode[0]).toBe('board');
  });

  it('lands on the timeline from a deep link with the presentation param', () => {
    const { result } = renderViewMode(['/business-domains?presentation=timeline']);
    expect(result.current.viewMode[0]).toBe('timeline');
  });

  it('reflects switching to the timeline in the URL', () => {
    const { result } = renderViewMode();

    act(() => result.current.viewMode[1]('timeline'));

    expect(result.current.viewMode[0]).toBe('timeline');
    expect(result.current.location.search).toContain('presentation=timeline');
  });

  it('drops the presentation param when leaving the timeline', () => {
    const { result } = renderViewMode(['/business-domains?presentation=timeline']);

    act(() => result.current.viewMode[1]('board'));

    expect(result.current.viewMode[0]).toBe('board');
    expect(result.current.location.search).not.toContain('presentation');
  });

  it('keeps the stored board or map preference while on the timeline', () => {
    const first = renderViewMode();
    act(() => first.result.current.viewMode[1]('map'));
    act(() => first.result.current.viewMode[1]('timeline'));

    const second = renderViewMode();
    expect(second.result.current.viewMode[0]).toBe('map');
  });

  it('ignores an invalid presentation value in the URL', () => {
    const { result } = renderViewMode(['/business-domains?presentation=bogus']);
    expect(result.current.viewMode[0]).toBe('board');
  });
});

describe('useMapDepth', () => {
  it('defaults to depth 2', () => {
    const { result } = renderHook(() => useMapDepth());
    expect(result.current[0]).toBe(2);
  });

  it('persists the chosen depth across hook instances', () => {
    const first = renderHook(() => useMapDepth());
    act(() => first.result.current[1](4));

    const second = renderHook(() => useMapDepth());
    expect(second.result.current[0]).toBe(4);
  });

  it('falls back to depth 2 when the stored value is out of range', () => {
    localStorage.setItem('business-domains-map-depth', '9');
    const { result } = renderHook(() => useMapDepth());
    expect(result.current[0]).toBe(2);
  });
});

describe('useMapShowApps', () => {
  it('hides apps by default', () => {
    const { result } = renderHook(() => useMapShowApps());
    expect(result.current[0]).toBe(false);
  });

  it('persists the choice across hook instances', () => {
    const first = renderHook(() => useMapShowApps());
    act(() => first.result.current[1](true));

    const second = renderHook(() => useMapShowApps());
    expect(second.result.current[0]).toBe(true);
  });
});

describe('resolveMapDomainId', () => {
  const domains = [buildBusinessDomain(), buildBusinessDomain()];

  it('keeps a stored id that matches an existing domain', () => {
    expect(resolveMapDomainId(String(domains[1].id), domains)).toBe(domains[1].id);
  });

  it('falls back to the first domain when the stored id no longer exists', () => {
    expect(resolveMapDomainId('gone', domains)).toBe(domains[0].id);
  });

  it('falls back to the first domain when nothing is stored', () => {
    expect(resolveMapDomainId(null, domains)).toBe(domains[0].id);
  });

  it('resolves to null when there are no domains', () => {
    expect(resolveMapDomainId('anything', [])).toBeNull();
  });
});

describe('useMapDomain', () => {
  const domains = [buildBusinessDomain({ id: toBusinessDomainId('dom-a') }), buildBusinessDomain({ id: toBusinessDomainId('dom-b') })];

  it('selects the first domain by default', () => {
    const { result } = renderHook(() => useMapDomain(domains));
    expect(result.current[0]).toBe(domains[0].id);
  });

  it('persists the chosen domain across hook instances', () => {
    const first = renderHook(() => useMapDomain(domains));
    act(() => first.result.current[1](domains[1].id));
    expect(first.result.current[0]).toBe(domains[1].id);

    const second = renderHook(() => useMapDomain(domains));
    expect(second.result.current[0]).toBe(domains[1].id);
  });

  it('falls back to the first domain when the stored domain was deleted', () => {
    localStorage.setItem('business-domains-map-domain', 'deleted-domain');
    const { result } = renderHook(() => useMapDomain(domains));
    expect(result.current[0]).toBe(domains[0].id);
  });
});
