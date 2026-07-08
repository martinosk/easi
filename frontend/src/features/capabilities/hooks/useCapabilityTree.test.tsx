import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { renderHook, waitFor } from '@testing-library/react';
import React from 'react';
import { describe, expect, it, vi } from 'vitest';
import { toCapabilityId } from '../../../api/types';
import { buildCapabilityAt as cap } from '../../../test/helpers/entityBuilders';
import { capabilitiesApi } from '../api';
import { buildCapabilityTree, useCapabilityTree } from './useCapabilityTree';

vi.mock('../api', () => ({
  capabilitiesApi: {
    getAll: vi.fn(),
  },
}));

describe('buildCapabilityTree', () => {
  it('nests children under their parents and sorts roots by name', () => {
    const tree = buildCapabilityTree([
      cap('b', 'Bravo', 'L1'),
      cap('a', 'Alpha', 'L1'),
      cap('a1', 'Alpha One', 'L2', 'a'),
    ]);

    expect(tree.map((n) => n.capability.name)).toEqual(['Alpha', 'Bravo']);
    expect(tree[0].children.map((n) => n.capability.name)).toEqual(['Alpha One']);
  });

  it('drops non-L1 orphans by default', () => {
    const tree = buildCapabilityTree([cap('a', 'Alpha', 'L1'), cap('x1', 'Orphan Two', 'L2', 'missing')]);

    expect(tree.map((n) => n.capability.name)).toEqual(['Alpha']);
  });

  it('roots orphans of any level when orphanRoots is any-level', () => {
    const tree = buildCapabilityTree([cap('a', 'Alpha', 'L1'), cap('x1', 'Orphan Two', 'L2', 'missing')], {
      orphanRoots: 'any-level',
    });

    expect(tree.map((n) => n.capability.name)).toEqual(['Alpha', 'Orphan Two']);
  });
});

describe('useCapabilityTree', () => {
  it('exposes the tree and childless L1 roots as orphanedL1Ids', async () => {
    vi.mocked(capabilitiesApi.getAll).mockResolvedValue([
      cap('a', 'Alpha', 'L1'),
      cap('a1', 'Alpha One', 'L2', 'a'),
      cap('b', 'Bravo', 'L1'),
    ]);
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const wrapper = ({ children }: { children: React.ReactNode }) => (
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    );

    const { result } = renderHook(() => useCapabilityTree(), { wrapper });
    await waitFor(() => expect(result.current.isLoading).toBe(false));

    expect(result.current.tree.map((n) => n.capability.name)).toEqual(['Alpha', 'Bravo']);
    expect(result.current.orphanedL1Ids).toEqual(new Set([toCapabilityId('b')]));
  });
});
