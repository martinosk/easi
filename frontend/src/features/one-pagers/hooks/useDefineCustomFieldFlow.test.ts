import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import React from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { DefineCustomFieldFormData } from '../../../lib/schemas/onePagerConfiguration';
import type { CustomField, OnePagerConfiguration } from '../types';
import { useDefineCustomFieldFlow } from './useDefineCustomFieldFlow';

vi.mock('../api/onePagersApi', () => ({
  onePagersApi: {
    defineCustomField: vi.fn(),
    setNumberFieldBounds: vi.fn(),
  },
}));

vi.mock('react-hot-toast', () => ({
  default: { success: vi.fn(), error: vi.fn() },
}));

import { onePagersApi } from '../api/onePagersApi';

function createWrapper(queryClient: QueryClient) {
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client: queryClient }, children);
}

function buildConfiguration(overrides: Partial<OnePagerConfiguration> = {}): OnePagerConfiguration {
  return {
    id: 'config-1',
    subjectType: 'application',
    builtInFields: [],
    customFields: [],
    displayOrder: [],
    version: 1,
    createdAt: '2026-01-01T00:00:00Z',
    modifiedAt: '2026-01-01T00:00:00Z',
    modifiedBy: 'admin@example.com',
    _links: {
      self: { href: '/api/v1/one-pagers/configurations/application', method: 'GET' },
      'x-define-custom-field': {
        href: '/api/v1/one-pagers/configurations/application/custom-fields',
        method: 'POST',
      },
    },
    ...overrides,
  };
}

function buildNumberField(overrides: Partial<CustomField> = {}): CustomField {
  return {
    id: 'field-1',
    name: 'Maturity score',
    type: 'number',
    required: false,
    helpText: '',
    active: true,
    _links: {
      'x-set-bounds': {
        href: '/api/v1/one-pagers/configurations/application/custom-fields/field-1/bounds',
        method: 'PUT',
      },
    },
    ...overrides,
  };
}

function formData(overrides: Partial<DefineCustomFieldFormData> = {}): DefineCustomFieldFormData {
  return {
    name: 'Maturity score',
    fieldType: 'number',
    required: false,
    helpText: '',
    options: [],
    min: '',
    max: '',
    ...overrides,
  };
}

describe('useDefineCustomFieldFlow', () => {
  let queryClient: QueryClient;

  beforeEach(() => {
    vi.clearAllMocks();
    queryClient = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } });
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  function renderFlow(configuration: OnePagerConfiguration) {
    return renderHook(() => useDefineCustomFieldFlow('application', configuration), {
      wrapper: createWrapper(queryClient),
    });
  }

  it('defines the field without sending bounds in the define request', async () => {
    const configuration = buildConfiguration();
    vi.mocked(onePagersApi.defineCustomField).mockResolvedValue(buildConfiguration({ version: 2 }));
    const { result } = renderFlow(configuration);

    act(() => {
      result.current.handleSubmit(formData({ min: 0, max: 5 }));
    });

    await waitFor(() =>
      expect(onePagersApi.defineCustomField).toHaveBeenCalledWith(configuration, {
        name: 'Maturity score',
        fieldType: 'number',
        required: false,
        helpText: '',
        options: undefined,
        version: 1,
      }),
    );
  });

  it.each([
    {
      name: 'both bounds are provided',
      formOverrides: { min: 0, max: 5 },
      expectedRequest: { min: 0, max: 5, version: 2 },
    },
    {
      name: 'only the minimum is provided',
      formOverrides: { min: 0, max: '' as const },
      expectedRequest: { min: 0, max: undefined, version: 2 },
    },
  ])('composes a set-bounds call when $name', async ({ formOverrides, expectedRequest }) => {
    const configuration = buildConfiguration();
    const numberField = buildNumberField();
    const updated = buildConfiguration({ version: 2, customFields: [numberField] });
    vi.mocked(onePagersApi.defineCustomField).mockResolvedValue(updated);
    vi.mocked(onePagersApi.setNumberFieldBounds).mockResolvedValue(buildConfiguration({ version: 3 }));
    const { result } = renderFlow(configuration);

    act(() => {
      result.current.handleSubmit(formData(formOverrides));
    });

    await waitFor(() =>
      expect(onePagersApi.setNumberFieldBounds).toHaveBeenCalledWith(numberField, expectedRequest),
    );
  });

  it.each([
    {
      name: 'neither bound is provided',
      customFields: [buildNumberField()],
      formOverrides: { min: '' as const, max: '' as const },
    },
    {
      name: 'the defined field has no x-set-bounds link',
      customFields: [buildNumberField({ _links: {} })],
      formOverrides: { min: 0, max: 5 },
    },
    {
      name: 'the field type is not number',
      customFields: [],
      formOverrides: { fieldType: 'text' as const, name: 'Notes', min: '' as const, max: '' as const },
    },
  ])('does not call set-bounds when $name', async ({ customFields, formOverrides }) => {
    const configuration = buildConfiguration();
    const updated = buildConfiguration({ version: 2, customFields });
    vi.mocked(onePagersApi.defineCustomField).mockResolvedValue(updated);
    const { result } = renderFlow(configuration);

    act(() => {
      result.current.handleSubmit(formData(formOverrides));
    });

    await waitFor(() => expect(onePagersApi.defineCustomField).toHaveBeenCalled());
    expect(onePagersApi.setNumberFieldBounds).not.toHaveBeenCalled();
  });
});
