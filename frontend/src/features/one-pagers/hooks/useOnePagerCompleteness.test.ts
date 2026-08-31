import { QueryClientProvider } from '@tanstack/react-query';
import { renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import React from 'react';
import { describe, expect, it } from 'vitest';
import { createTestQueryClient } from '../../../test/helpers';
import { seedOnePagerCompleteness } from '../../../test/mocks/onePagerCompleteness';
import { useOnePagerCompleteness } from './useOnePagerCompleteness';

function wrapper({ children }: { children: ReactNode }) {
  const client = createTestQueryClient();
  return React.createElement(QueryClientProvider, { client }, children);
}

describe('useOnePagerCompleteness', () => {
  it('indexes completeness by subject id for the subject type', async () => {
    seedOnePagerCompleteness('application', [
      { subjectId: 'comp-a', complete: true },
      { subjectId: 'comp-b', complete: false },
    ]);

    const { result } = renderHook(() => useOnePagerCompleteness('application'), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data!.get('comp-a')).toBe(true);
    expect(result.current.data!.get('comp-b')).toBe(false);
  });

  it('yields an empty lookup when the subject type has no required field', async () => {
    seedOnePagerCompleteness('vendor', []);

    const { result } = renderHook(() => useOnePagerCompleteness('vendor'), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data!.size).toBe(0);
    expect(result.current.data!.get('v-1')).toBeUndefined();
  });

  it('keeps subject types isolated from one another', async () => {
    seedOnePagerCompleteness('application', [{ subjectId: 'shared-id', complete: false }]);
    seedOnePagerCompleteness('capability', [{ subjectId: 'shared-id', complete: true }]);

    const { result } = renderHook(() => useOnePagerCompleteness('capability'), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data!.get('shared-id')).toBe(true);
  });
});
