import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook } from '@testing-library/react';
import React from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { ApiError, type SubjectAttribute, type SubjectAttributeSchema } from '../../../api/types';
import { metadataQueryKeys } from '../../../lib/appQueryKeys';
import { onePagersQueryKeys } from '../queryKeys';
import { useDefineAttribute, useRetireAttribute, useSetAttributeBounds } from './useAttributeSchemaMutations';

vi.mock('../../../api/metadata', () => ({
  subjectAttributeSchemaApi: {
    defineAttribute: vi.fn(),
    retireAttribute: vi.fn(),
    setBounds: vi.fn(),
  },
}));

vi.mock('react-hot-toast', () => ({
  default: { success: vi.fn(), error: vi.fn() },
}));

import toast from 'react-hot-toast';
import { subjectAttributeSchemaApi } from '../../../api/metadata';

function createWrapper(queryClient: QueryClient) {
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client: queryClient }, children);
}

function buildSchema(overrides: Partial<SubjectAttributeSchema> = {}): SubjectAttributeSchema {
  return {
    id: 'schema-1',
    subjectType: 'vendor',
    attributes: [],
    version: 1,
    createdAt: '2026-01-01T00:00:00Z',
    modifiedAt: '2026-01-01T00:00:00Z',
    modifiedBy: 'steward@example.com',
    _links: {
      self: { href: '/api/v1/meta-model/subject-types/vendor/attributes', method: 'GET' },
      'x-define': { href: '/api/v1/meta-model/subject-types/vendor/attributes', method: 'POST' },
    },
    ...overrides,
  };
}

function buildAttribute(overrides: Partial<SubjectAttribute> = {}): SubjectAttribute {
  return {
    id: 'attr-1',
    name: 'Score',
    type: 'number',
    helpText: '',
    active: true,
    _links: {
      'x-retire': { href: '/api/v1/meta-model/subject-types/vendor/attributes/attr-1/retire', method: 'POST' },
      'x-set-bounds': { href: '/api/v1/meta-model/subject-types/vendor/attributes/attr-1/bounds', method: 'PUT' },
    },
    ...overrides,
  };
}

describe('useAttributeSchemaMutations', () => {
  let queryClient: QueryClient;

  beforeEach(() => {
    vi.clearAllMocks();
    queryClient = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } });
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('defines an attribute and invalidates both the schema and the one-pager configuration', async () => {
    const schema = buildSchema();
    vi.mocked(subjectAttributeSchemaApi.defineAttribute).mockResolvedValue(buildSchema({ version: 2 }));
    const invalidateQueriesSpy = vi.spyOn(queryClient, 'invalidateQueries');

    const { result } = renderHook(() => useDefineAttribute('vendor'), { wrapper: createWrapper(queryClient) });

    await act(async () => {
      await result.current.mutateAsync({
        schema,
        request: { name: 'Score', type: 'number', helpText: '', min: 0, version: 1 },
      });
    });

    expect(subjectAttributeSchemaApi.defineAttribute).toHaveBeenCalledWith(schema, {
      name: 'Score',
      type: 'number',
      helpText: '',
      min: 0,
      version: 1,
    });
    expect(invalidateQueriesSpy).toHaveBeenCalledWith({ queryKey: metadataQueryKeys.subjectAttributeSchema('vendor') });
    expect(invalidateQueriesSpy).toHaveBeenCalledWith({ queryKey: onePagersQueryKeys.configuration('vendor') });
    expect(toast.success).toHaveBeenCalledWith('Custom field defined');
  });

  it('sets bounds through the attribute link', async () => {
    const attribute = buildAttribute();
    vi.mocked(subjectAttributeSchemaApi.setBounds).mockResolvedValue(buildSchema({ version: 2 }));

    const { result } = renderHook(() => useSetAttributeBounds('vendor'), { wrapper: createWrapper(queryClient) });

    await act(async () => {
      await result.current.mutateAsync({ attribute, request: { min: 0, max: 5, version: 1 } });
    });

    expect(subjectAttributeSchemaApi.setBounds).toHaveBeenCalledWith(attribute, { min: 0, max: 5, version: 1 });
    expect(toast.success).toHaveBeenCalledWith('Bounds updated');
  });

  it('surfaces a schema conflict message and refetches on 409', async () => {
    const attribute = buildAttribute();
    vi.mocked(subjectAttributeSchemaApi.retireAttribute).mockRejectedValue(new ApiError('Version conflict', 409));
    const invalidateQueriesSpy = vi.spyOn(queryClient, 'invalidateQueries');

    const { result } = renderHook(() => useRetireAttribute('vendor'), { wrapper: createWrapper(queryClient) });

    await act(async () => {
      await expect(result.current.mutateAsync({ attribute, request: { version: 1 } })).rejects.toThrow();
    });

    expect(invalidateQueriesSpy).toHaveBeenCalledWith({ queryKey: metadataQueryKeys.subjectAttributeSchema('vendor') });
    expect(toast.error).toHaveBeenCalledWith('Attribute schema was changed by someone else. Refreshed with the latest version.');
  });

  it('shows a generic error message on non-conflict failures', async () => {
    const attribute = buildAttribute();
    vi.mocked(subjectAttributeSchemaApi.retireAttribute).mockRejectedValue(new Error('Server exploded'));

    const { result } = renderHook(() => useRetireAttribute('vendor'), { wrapper: createWrapper(queryClient) });

    await act(async () => {
      await expect(result.current.mutateAsync({ attribute, request: { version: 1 } })).rejects.toThrow();
    });

    expect(toast.error).toHaveBeenCalledWith('Server exploded');
  });
});
