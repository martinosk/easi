import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import React from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { SubjectAttributeSchema } from '../../../api/types';
import type { DefineCustomFieldFormData } from '../../../lib/schemas/onePagerConfiguration';
import type { CustomField, OnePagerConfiguration } from '../types';
import { useDefineCustomFieldFlow } from './useDefineCustomFieldFlow';

vi.mock('../../../api/metadata', () => ({
  subjectAttributeSchemaApi: {
    defineAttribute: vi.fn(),
  },
}));

vi.mock('../api/onePagersApi', () => ({
  onePagersApi: {
    changeFieldRequirement: vi.fn(),
  },
}));

vi.mock('react-hot-toast', () => ({
  default: { success: vi.fn(), error: vi.fn() },
}));

import { subjectAttributeSchemaApi } from '../../../api/metadata';
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
      'x-attribute-schema': { href: '/api/v1/meta-model/subject-types/application/attributes', method: 'GET' },
    },
    ...overrides,
  };
}

function buildSchema(overrides: Partial<SubjectAttributeSchema> = {}): SubjectAttributeSchema {
  return {
    id: 'schema-1',
    subjectType: 'application',
    attributes: [],
    version: 3,
    createdAt: '2026-01-01T00:00:00Z',
    modifiedAt: '2026-01-01T00:00:00Z',
    modifiedBy: 'steward@example.com',
    _links: {
      self: { href: '/api/v1/meta-model/subject-types/application/attributes', method: 'GET' },
      'x-define': { href: '/api/v1/meta-model/subject-types/application/attributes', method: 'POST' },
    },
    ...overrides,
  };
}

function includedField(overrides: Partial<CustomField> = {}): CustomField {
  return {
    id: 'attr-new',
    name: 'Maturity score',
    type: 'number',
    required: false,
    helpText: '',
    active: true,
    included: true,
    _links: {
      'x-set-requirement': {
        href: '/api/v1/one-pagers/configurations/application/custom-fields/attr-new/requirement',
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

  function renderFlow(configuration: OnePagerConfiguration | undefined, schema: SubjectAttributeSchema | undefined) {
    return renderHook(({ config }) => useDefineCustomFieldFlow('application', config, schema), {
      wrapper: createWrapper(queryClient),
      initialProps: { config: configuration },
    });
  }

  it.each([
    {
      name: 'a number field with bounds',
      form: formData({ min: 0, max: 5 }),
      expected: { name: 'Maturity score', type: 'number', helpText: '', options: undefined, min: 0, max: 5, version: 3 },
    },
    {
      name: 'a number field with only the minimum',
      form: formData({ min: 0 }),
      expected: { name: 'Maturity score', type: 'number', helpText: '', options: undefined, min: 0, max: undefined, version: 3 },
    },
    {
      name: 'a selection field with its options',
      form: formData({ name: 'Region', fieldType: 'selection', options: ['EU', 'US'] }),
      expected: { name: 'Region', type: 'selection', helpText: '', options: ['EU', 'US'], min: undefined, max: undefined, version: 3 },
    },
  ])('defines $name through the MetaModel schema in one request', async ({ form, expected }) => {
    const schema = buildSchema();
    vi.mocked(subjectAttributeSchemaApi.defineAttribute).mockResolvedValue(buildSchema({ version: 4 }));
    const { result } = renderFlow(buildConfiguration(), schema);

    act(() => {
      result.current.handleSubmit(form);
    });

    await waitFor(() => expect(subjectAttributeSchemaApi.defineAttribute).toHaveBeenCalledWith(schema, expected));
    expect(onePagersApi.changeFieldRequirement).not.toHaveBeenCalled();
  });

  it('marks the field required on the one-pager once the refreshed configuration offers it', async () => {
    const schema = buildSchema();
    const definedAttribute = { id: 'attr-new', name: 'Maturity score', type: 'number' as const, helpText: '', active: true };
    vi.mocked(subjectAttributeSchemaApi.defineAttribute).mockResolvedValue(
      buildSchema({ version: 4, attributes: [definedAttribute] }),
    );
    vi.mocked(onePagersApi.changeFieldRequirement).mockResolvedValue(buildConfiguration({ version: 3 }));
    const { result, rerender } = renderFlow(buildConfiguration(), schema);

    act(() => {
      result.current.handleSubmit(formData({ required: true }));
    });
    await waitFor(() => expect(subjectAttributeSchemaApi.defineAttribute).toHaveBeenCalled());
    expect(onePagersApi.changeFieldRequirement).not.toHaveBeenCalled();

    const refreshed = buildConfiguration({ version: 2, customFields: [includedField()] });
    rerender({ config: refreshed });

    await waitFor(() =>
      expect(onePagersApi.changeFieldRequirement).toHaveBeenCalledWith(includedField(), { required: true, version: 2 }),
    );
    expect(onePagersApi.changeFieldRequirement).toHaveBeenCalledTimes(1);
  });

  it('gates a required new field behind the impact preview when the configuration offers one', () => {
    const configuration = buildConfiguration({
      _links: {
        ...buildConfiguration()._links,
        'x-impact-preview': { href: '/api/v1/one-pagers/configurations/application/impact-preview', method: 'GET' },
      },
    });
    const { result } = renderFlow(configuration, buildSchema());

    act(() => {
      result.current.handleSubmit(formData({ required: true }));
    });

    expect(result.current.pendingNewField?.name).toBe('Maturity score');
    expect(subjectAttributeSchemaApi.defineAttribute).not.toHaveBeenCalled();

    act(() => {
      result.current.cancelPendingField();
    });
    expect(result.current.pendingNewField).toBeNull();
  });

  it('does nothing when the schema has not loaded', () => {
    const { result } = renderFlow(buildConfiguration(), undefined);

    act(() => {
      result.current.handleSubmit(formData());
    });

    expect(subjectAttributeSchemaApi.defineAttribute).not.toHaveBeenCalled();
  });
});
