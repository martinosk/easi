import { act, renderHook } from '@testing-library/react';
import type { ReactNode } from 'react';
import { MemoryRouter, useLocation } from 'react-router-dom';
import { describe, expect, it } from 'vitest';
import { useBoardLensState } from './useBoardLensState';

function wrapper(initialEntries: string[]) {
  return ({ children }: { children: ReactNode }) => (
    <MemoryRouter initialEntries={initialEntries}>{children}</MemoryRouter>
  );
}

describe('useBoardLensState', () => {
  it('defaults to the Now lens with no lens param', () => {
    const { result } = renderHook(() => useBoardLensState(), { wrapper: wrapper(['/business-domains']) });
    expect(result.current.lens).toBe('now');
  });

  it('reads the active lens from the URL so a shared link lands on it', () => {
    const { result } = renderHook(() => useBoardLensState(), { wrapper: wrapper(['/business-domains?lens=journey']) });
    expect(result.current.lens).toBe('journey');
  });

  it('reflects a lens change in the URL query parameter', () => {
    const { result } = renderHook(() => ({ state: useBoardLensState(), location: useLocation() }), {
      wrapper: wrapper(['/business-domains']),
    });

    act(() => result.current.state.setLens('target'));

    expect(result.current.state.lens).toBe('target');
    expect(result.current.location.search).toContain('lens=target');
  });

  it('drops the lens param when returning to the Now default', () => {
    const { result } = renderHook(() => ({ state: useBoardLensState(), location: useLocation() }), {
      wrapper: wrapper(['/business-domains?lens=journey']),
    });

    act(() => result.current.state.setLens('now'));

    expect(result.current.location.search).not.toContain('lens');
  });

  it('ignores an invalid lens value in the URL', () => {
    const { result } = renderHook(() => useBoardLensState(), { wrapper: wrapper(['/business-domains?lens=bogus']) });
    expect(result.current.lens).toBe('now');
  });
});
