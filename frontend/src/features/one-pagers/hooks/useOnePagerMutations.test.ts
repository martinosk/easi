import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook } from '@testing-library/react';
import React from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { ApiError } from '../../../api/types';
import { onePagersQueryKeys } from '../queryKeys';
import type { CustomField, OnePagerConfiguration } from '../types';
import {
  useChangeBuiltInFieldRequirement,
  useChangeFieldRequirement,
  useIncludeBuiltInField,
  useReorderFields,
} from './useOnePagerMutations';

vi.mock('../api/onePagersApi', () => ({
  onePagersApi: {
    reorderFields: vi.fn(),
    includeBuiltInField: vi.fn(),
    changeBuiltInFieldRequirement: vi.fn(),
    changeFieldRequirement: vi.fn(),
  },
}));

vi.mock('react-hot-toast', () => ({
  default: { success: vi.fn(), error: vi.fn() },
}));

import toast from 'react-hot-toast';
import { onePagersApi } from '../api/onePagersApi';

function createWrapper(queryClient: QueryClient) {
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client: queryClient }, children);
}

function buildConfiguration(overrides: Partial<OnePagerConfiguration> = {}): OnePagerConfiguration {
  return {
    id: 'config-1',
    subjectType: 'vendor',
    builtInFields: [],
    customFields: [],
    displayOrder: [],
    version: 1,
    createdAt: '2026-01-01T00:00:00Z',
    modifiedAt: '2026-01-01T00:00:00Z',
    modifiedBy: 'admin@example.com',
    _links: {
      self: { href: '/api/v1/one-pagers/configurations/vendor', method: 'GET' },
      'x-reorder': { href: '/api/v1/one-pagers/configurations/vendor/display-order', method: 'PUT' },
      'x-attribute-schema': { href: '/api/v1/meta-model/subject-types/vendor/attributes', method: 'GET' },
    },
    ...overrides,
  };
}

function buildCustomField(overrides: Partial<CustomField> = {}): CustomField {
  return {
    id: 'field-1',
    name: 'Contract',
    type: 'link',
    required: false,
    helpText: '',
    active: true,
    included: true,
    _links: {
      'x-set-requirement': {
        href: '/api/v1/one-pagers/configurations/vendor/custom-fields/field-1/requirement',
        method: 'PUT',
      },
    },
    ...overrides,
  };
}

describe('useOnePagerMutations', () => {
  let queryClient: QueryClient;

  beforeEach(() => {
    vi.clearAllMocks();
    queryClient = new QueryClient({
      defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
    });
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('changes a custom field requirement and invalidates the configuration query', async () => {
    const field = buildCustomField();
    vi.mocked(onePagersApi.changeFieldRequirement).mockResolvedValue(buildConfiguration({ version: 2 }));
    const invalidateQueriesSpy = vi.spyOn(queryClient, 'invalidateQueries');

    const { result } = renderHook(() => useChangeFieldRequirement('vendor'), { wrapper: createWrapper(queryClient) });

    await act(async () => {
      await result.current.mutateAsync({ field, request: { required: true, version: 1 } });
    });

    expect(onePagersApi.changeFieldRequirement).toHaveBeenCalledWith(field, { required: true, version: 1 });
    expect(invalidateQueriesSpy).toHaveBeenCalledWith({ queryKey: onePagersQueryKeys.configuration('vendor') });
    expect(toast.success).toHaveBeenCalledWith('Requirement updated');
  });

  it('reorders fields', async () => {
    const configuration = buildConfiguration();
    vi.mocked(onePagersApi.reorderFields).mockResolvedValue(buildConfiguration({ version: 2 }));

    const { result } = renderHook(() => useReorderFields('vendor'), { wrapper: createWrapper(queryClient) });

    await act(async () => {
      await result.current.mutateAsync({
        configuration,
        request: { order: [{ kind: 'custom', id: 'field-1' }], version: 1 },
      });
    });

    expect(onePagersApi.reorderFields).toHaveBeenCalledWith(configuration, {
      order: [{ kind: 'custom', id: 'field-1' }],
      version: 1,
    });
    expect(toast.success).toHaveBeenCalledWith('Field order updated');
  });

  it('includes a built-in field', async () => {
    const field = { id: 'name', label: 'Name', included: false, required: false, _links: { 'x-include': { href: '/x', method: 'POST' as const } } };
    vi.mocked(onePagersApi.includeBuiltInField).mockResolvedValue(buildConfiguration());

    const { result } = renderHook(() => useIncludeBuiltInField('vendor'), { wrapper: createWrapper(queryClient) });

    await act(async () => {
      await result.current.mutateAsync({ field, request: { version: 1 } });
    });

    expect(onePagersApi.includeBuiltInField).toHaveBeenCalledWith(field, { version: 1 });
  });

  it('changes a built-in field requirement', async () => {
    const field = {
      id: 'experts',
      label: 'Experts',
      included: true,
      required: false,
      _links: {
        'x-set-requirement': {
          href: '/api/v1/one-pagers/configurations/vendor/built-in-fields/experts/requirement',
          method: 'PUT' as const,
        },
      },
    };
    vi.mocked(onePagersApi.changeBuiltInFieldRequirement).mockResolvedValue(buildConfiguration({ version: 2 }));

    const { result } = renderHook(() => useChangeBuiltInFieldRequirement('vendor'), {
      wrapper: createWrapper(queryClient),
    });

    await act(async () => {
      await result.current.mutateAsync({ field, request: { required: true, version: 1 } });
    });

    expect(onePagersApi.changeBuiltInFieldRequirement).toHaveBeenCalledWith(field, { required: true, version: 1 });
    expect(toast.success).toHaveBeenCalledWith('Requirement updated');
  });

  it('surfaces a conflict message and refetches on 409', async () => {
    const field = buildCustomField();
    const conflict = new ApiError('Version conflict', 409);
    vi.mocked(onePagersApi.changeFieldRequirement).mockRejectedValue(conflict);
    const invalidateQueriesSpy = vi.spyOn(queryClient, 'invalidateQueries');

    const { result } = renderHook(() => useChangeFieldRequirement('vendor'), { wrapper: createWrapper(queryClient) });

    await act(async () => {
      await expect(result.current.mutateAsync({ field, request: { required: true, version: 1 } })).rejects.toThrow();
    });

    expect(invalidateQueriesSpy).toHaveBeenCalledWith({ queryKey: onePagersQueryKeys.configuration('vendor') });
    expect(toast.error).toHaveBeenCalledWith('Configuration was changed by someone else. Refreshed with the latest version.');
  });

  it('shows a generic error message on non-conflict failures', async () => {
    const field = buildCustomField();
    vi.mocked(onePagersApi.changeFieldRequirement).mockRejectedValue(new Error('Server exploded'));

    const { result } = renderHook(() => useChangeFieldRequirement('vendor'), { wrapper: createWrapper(queryClient) });

    await act(async () => {
      await expect(result.current.mutateAsync({ field, request: { required: true, version: 1 } })).rejects.toThrow();
    });

    expect(toast.error).toHaveBeenCalledWith('Server exploded');
  });
});
