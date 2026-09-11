import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import React from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { SubjectAttribute, SubjectAttributeSchema } from '../../../api/types';
import { useAttributeSchemaActions } from './useAttributeSchemaActions';

vi.mock('../../../api/metadata', () => ({
  subjectAttributeSchemaApi: {
    renameAttribute: vi.fn(),
    retireAttribute: vi.fn(),
    reactivateAttribute: vi.fn(),
    addOption: vi.fn(),
    retireOption: vi.fn(),
    setBounds: vi.fn(),
  },
}));

vi.mock('react-hot-toast', () => ({
  default: { success: vi.fn(), error: vi.fn() },
}));

import { subjectAttributeSchemaApi } from '../../../api/metadata';

function createWrapper(queryClient: QueryClient) {
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client: queryClient }, children);
}

const attribute: SubjectAttribute = {
  id: 'attr-1',
  name: 'Hosting Region',
  type: 'selection',
  helpText: '',
  active: true,
  options: [{ id: 'opt-1', label: 'EU', active: true, _links: { 'x-retire': { href: '/o', method: 'POST' } } }],
  _links: {},
};

const schema: SubjectAttributeSchema = {
  id: 'schema-1',
  subjectType: 'application',
  attributes: [attribute],
  version: 7,
  createdAt: '',
  modifiedAt: '',
  modifiedBy: 'steward@example.com',
  _links: { self: { href: '/s', method: 'GET' } },
};

describe('useAttributeSchemaActions', () => {
  let queryClient: QueryClient;

  beforeEach(() => {
    vi.clearAllMocks();
    queryClient = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } });
    vi.mocked(subjectAttributeSchemaApi.retireAttribute).mockResolvedValue(schema);
    vi.mocked(subjectAttributeSchemaApi.reactivateAttribute).mockResolvedValue(schema);
    vi.mocked(subjectAttributeSchemaApi.addOption).mockResolvedValue(schema);
    vi.mocked(subjectAttributeSchemaApi.retireOption).mockResolvedValue(schema);
    vi.mocked(subjectAttributeSchemaApi.setBounds).mockResolvedValue(schema);
    vi.mocked(subjectAttributeSchemaApi.renameAttribute).mockResolvedValue(schema);
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  function renderActions(currentSchema: SubjectAttributeSchema | undefined = schema) {
    const onRename = vi.fn();
    const hook = renderHook(() => useAttributeSchemaActions('application', currentSchema, onRename), {
      wrapper: createWrapper(queryClient),
    });
    return { ...hook, onRename };
  }

  it('sends every schema command with the current schema version', async () => {
    const { result } = renderActions();

    act(() => {
      result.current.actions.onRetire(attribute);
      result.current.actions.onAddOption(attribute, 'US');
      result.current.actions.onRetireOption(attribute.options?.[0] as NonNullable<typeof attribute.options>[number]);
      result.current.actions.onSetBounds(attribute, 1, undefined);
      result.current.reactivateAttribute(attribute);
    });

    await waitFor(() => expect(subjectAttributeSchemaApi.retireAttribute).toHaveBeenCalledWith(attribute, { version: 7 }));
    expect(subjectAttributeSchemaApi.addOption).toHaveBeenCalledWith(attribute, { label: 'US', version: 7 });
    expect(subjectAttributeSchemaApi.retireOption).toHaveBeenCalledWith(attribute.options?.[0], { version: 7 });
    expect(subjectAttributeSchemaApi.setBounds).toHaveBeenCalledWith(attribute, { min: 1, max: undefined, version: 7 });
    expect(subjectAttributeSchemaApi.reactivateAttribute).toHaveBeenCalledWith(attribute, { version: 7 });
  });

  it('renames with the attribute type restated and reports completion', async () => {
    const { result } = renderActions();
    const onSaved = vi.fn();

    act(() => {
      result.current.saveRename(attribute, { name: 'Region', helpText: 'Where it runs' }, onSaved);
    });

    await waitFor(() =>
      expect(subjectAttributeSchemaApi.renameAttribute).toHaveBeenCalledWith(attribute, {
        name: 'Region',
        helpText: 'Where it runs',
        type: 'selection',
        version: 7,
      }),
    );
    await waitFor(() => expect(onSaved).toHaveBeenCalled());
  });

  it('delegates rename requests to the dialog opener', () => {
    const { result, onRename } = renderActions();

    act(() => {
      result.current.actions.onRename(attribute);
    });

    expect(onRename).toHaveBeenCalledWith(attribute);
  });

  it('does nothing while the schema has not loaded', () => {
    const { result } = renderActions(undefined);

    act(() => {
      result.current.actions.onRetire(attribute);
      result.current.reactivateAttribute(attribute);
    });

    expect(subjectAttributeSchemaApi.retireAttribute).not.toHaveBeenCalled();
    expect(subjectAttributeSchemaApi.reactivateAttribute).not.toHaveBeenCalled();
  });
});
