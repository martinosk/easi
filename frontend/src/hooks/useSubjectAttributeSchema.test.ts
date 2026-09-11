import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { renderHook, waitFor } from '@testing-library/react';
import React from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { SubjectAttributeSchema } from '../api/types';
import { metadataQueryKeys } from '../lib/appQueryKeys';
import { useSubjectAttributeSchema } from './useSubjectAttributeSchema';

vi.mock('../api/metadata', () => ({
  subjectAttributeSchemaApi: { getSchema: vi.fn() },
}));

import { subjectAttributeSchemaApi } from '../api/metadata';

function createWrapper(queryClient: QueryClient) {
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client: queryClient }, children);
}

const schema: SubjectAttributeSchema = {
  id: 'schema-1',
  subjectType: 'vendor',
  attributes: [],
  version: 1,
  createdAt: '',
  modifiedAt: '',
  modifiedBy: 'steward@example.com',
  _links: { self: { href: '/api/v1/meta-model/subject-types/vendor/attributes', method: 'GET' } },
};

describe('useSubjectAttributeSchema', () => {
  let queryClient: QueryClient;

  beforeEach(() => {
    vi.clearAllMocks();
    queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  });

  it('fetches the schema for a subject type under the metadata key', async () => {
    vi.mocked(subjectAttributeSchemaApi.getSchema).mockResolvedValue(schema);

    const { result } = renderHook(() => useSubjectAttributeSchema('vendor'), { wrapper: createWrapper(queryClient) });

    await waitFor(() => expect(result.current.data).toEqual(schema));
    expect(subjectAttributeSchemaApi.getSchema).toHaveBeenCalledWith('vendor');
    expect(queryClient.getQueryData(metadataQueryKeys.subjectAttributeSchema('vendor'))).toEqual(schema);
  });

  it('does not fetch while disabled', () => {
    renderHook(() => useSubjectAttributeSchema('vendor', false), { wrapper: createWrapper(queryClient) });

    expect(subjectAttributeSchemaApi.getSchema).not.toHaveBeenCalled();
  });
});
